package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create:        resourceDatabaseCreate, // POST
		Read:          resourceDatabaseRead,   // GET
		UpdateContext: resourceDatabaseUpdate, // PUT
		Delete:        resourceDatabaseDelete, // DELETE
		Importer: &schema.ResourceImporter{
			State: resourceDatabaseImport,
		},

		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"database_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validDBTypes, false),
				Default:      "postgresql",
				ForceNew:     true,
			},
			"container_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateContainerSize,
				Default:      1024,
			},
			"container_profile": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateContainerProfile,
				Default:      "m5",
			},
			"disk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateDiskSize,
				Default:      10,
			},
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"database_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_connection_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"version": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressDefaultDatabaseVersion,
			},
			"database_image_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_urls": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).Client
	legacy := meta.(*providerMetadata).LegacyClient
	envID := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)
	version := d.Get("version").(string)
	databaseType := d.Get("database_type").(string)
	ctx := context.Background()

	create := aptibleapi.NewCreateDatabaseRequestWithDefaults()
	create.SetHandle(handle)
	createDb := client.DatabasesAPI.CreateDatabase(ctx, int32(envID))
	create.SetType(databaseType)
	create.SetInitialDiskSize(int32(d.Get("disk_size").(int)))
	create.SetInitialContainerSize(int32(d.Get("container_size").(int)))

	if version != "" {
		image, err := legacy.GetDatabaseImageByTypeAndVersion(databaseType, version)
		if err != nil {
			log.Println(err)
			return generateErrorFromClientError(err)
		}
		create.SetDatabaseImageId(int32(image.ID))
	}

	db, _, err := createDb.CreateDatabaseRequest(*create).Execute()
	if err != nil {
		return err
	}

	createOp := client.OperationsAPI.CreateOperationForDatabase(ctx, db.Id)
	payload := aptibleapi.NewCreateOperationRequest("provision")
	payload.SetProvisionedIops(int32(d.Get("iops").(int)))
	payload.SetInstanceProfile(d.Get("container_profile").(string))
	op, _, err := createOp.CreateOperationRequest(*payload).Execute()
	if err != nil {
		return err
	}

	success, err := legacy.WaitForOperation(int64(op.Id))
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("Could not provision database")
	}

	_ = d.Set("database_id", db.Id)

	d.SetId(strconv.Itoa(int(db.Id)))
	return resourceDatabaseRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).LegacyClient
	databaseID := int64(d.Get("database_id").(int))

	database, err := client.GetDatabase(databaseID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}

	if database.Deleted {
		d.SetId("")
		log.Println("Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	_ = d.Set("container_size", database.ContainerSize)
	_ = d.Set("container_profile", database.ContainerProfile)
	_ = d.Set("disk_size", database.DiskSize)
	_ = d.Set("default_connection_url", database.DefaultConnection)
	_ = d.Set("connection_urls", database.ConnectionURLs)
	_ = d.Set("handle", database.Handle)
	_ = d.Set("env_id", database.EnvironmentID)
	_ = d.Set("database_type", database.Type)
	_ = d.Set("database_image_id", database.DatabaseImage.ID)
	_ = d.Set("version", database.DatabaseImage.Version)

	return nil
}

func resourceDatabaseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	databaseID, _ := strconv.Atoi(d.Id())
	_ = d.Set("database_id", databaseID)
	err := resourceDatabaseRead(d, meta)
	return []*schema.ResourceData{d}, err
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).LegacyClient
	databaseID := int64(d.Get("database_id").(int))
	containerSize := int64(d.Get("container_size").(int))
	containerProfile := d.Get("container_profile").(string)
	diskSize := int64(d.Get("disk_size").(int))
	handle := d.Get("handle").(string)
	var diags diag.Diagnostics

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = containerSize
	}
	if d.HasChange("container_profile") {
		updates.ContainerProfile = containerProfile
	}
	if d.HasChange("disk_size") {
		updates.DiskSize = diskSize
	}

	if d.HasChange("handle") {
		log.Printf("[INFO] Updating handle to %s\n", handle)
		updates.Handle = handle
	}

	// if only changing the handle, you can skip running an operation needlessly
	// below can be hard to read, but it basically means if nothing besides handle was changed
	if !d.HasChangesExcept("handle") {
		updates.SkipOperationUpdate = true
	}

	err := client.UpdateDatabase(databaseID, updates)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to update the database.",
			Detail:   generateErrorFromClientError(err).Error(),
		})
		return diags
	}

	if d.HasChange("handle") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("You must reload the database to see changes. In order for the new database name (%s) to appear in log drain and metric drain destinations, you must reload the database.  You can use the CLI to do this with: 'aptible db:reload %s'", handle, handle),
		})
		log.Printf("[WARN] In order for the new database name (%s) to appear in log drain and metric drain destinations, you must restart the database.\n", handle)
	}

	if err := resourceDatabaseRead(d, meta); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to retrieve the updated state of the database.",
			Detail:   err.Error(),
		})
	}

	return diags
}

func resourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).LegacyClient
	databaseID := int64(d.Get("database_id").(int))

	err := client.DeleteDatabase(databaseID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}

	d.SetId("")
	return nil
}

var validDBTypes = []string{
	"couchdb",
	"elasticsearch",
	"influxdb",
	"mongodb",
	"mysql",
	"postgresql",
	"rabbitmq",
	"redis",
	"sftp",
}
