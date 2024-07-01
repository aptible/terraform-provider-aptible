package aptible

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
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
				Default:  3000,
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
	iops := int32(d.Get("iops").(int))
	profile := d.Get("container_profile").(string)
	diskSize := int32(d.Get("disk_size").(int))
	containerSize := int32(d.Get("container_size").(int))

	ctx := context.Background()
	ctx = meta.(*providerMetadata).APIContext(ctx)

	create := aptibleapi.NewCreateDatabaseRequestWithDefaults()
	create.SetHandle(handle)
	createDb := client.DatabasesAPI.CreateDatabase(ctx, int32(envID))
	create.SetType(databaseType)

	if diskSize != 0 {
		create.SetInitialDiskSize(diskSize)
	}
	if containerSize != 0 {
		create.SetInitialContainerSize(containerSize)
	}

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

	payload := aptibleapi.NewCreateOperationRequest("provision")
	if containerSize != 0 {
		payload.SetContainerSize(containerSize)
	}
	if iops != 0 {
		payload.SetProvisionedIops(iops)
	}
	if profile != "" {
		payload.SetInstanceProfile(profile)
	}

	op, _, err := client.
		OperationsAPI.
		CreateOperationForDatabase(ctx, db.Id).
		CreateOperationRequest(*payload).
		Execute()
	if err != nil {
		return err
	}

	del, err := legacy.WaitForOperation(int64(op.Id))
	if err != nil {
		return err
	}
	if del {
		return fmt.Errorf("the replica with handle: %s was unexpectedly deleted", handle)
	}

	_ = d.Set("database_id", db.Id)

	d.SetId(strconv.Itoa(int(db.Id)))
	return resourceDatabaseRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).Client
	databaseID := int32(d.Get("database_id").(int))
	ctx := context.Background()
	ctx = meta.(*providerMetadata).APIContext(ctx)

	database, resp, err := client.DatabasesAPI.GetDatabase(ctx, databaseID).Execute()
	if err != nil {
		return generateErrorFromClientError(err)
	}
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		log.Println("Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	urls := []string{}
	creds := database.Embedded.DatabaseCredentials
	for _, cred := range creds {
		urls = append(urls, cred.ConnectionUrl)
	}

	imageID := ExtractIdFromLink(*database.Links.DatabaseImage.Href)
	if imageID == 0 {
		return fmt.Errorf("Could not find database image ID")
	}
	serviceID := ExtractIdFromLink(*database.Links.Service.Href)
	if serviceID == 0 {
		return fmt.Errorf("Could not find database service ID")
	}
	accountID := ExtractIdFromLink(*database.Links.Account.Href)
	if accountID == 0 {
		return fmt.Errorf("Could not find database account ID")
	}

	image, _, err := client.ImagesAPI.GetDatabaseImage(ctx, imageID).Execute()
	if err != nil {
		return generateErrorFromClientError(err)
	}

	service, _, err := client.ServicesAPI.GetService(ctx, serviceID).Execute()
	if err != nil {
		return generateErrorFromClientError(err)
	}

	containerSize := service.ContainerMemoryLimitMb.Get()
	profile := service.InstanceClass

	_ = d.Set("container_size", containerSize)
	_ = d.Set("container_profile", profile)
	_ = d.Set("iops", database.Embedded.Disk.ProvisionedIops)
	_ = d.Set("disk_size", database.Embedded.Disk.Size)
	_ = d.Set("default_connection_url", database.ConnectionUrl.Get())
	_ = d.Set("connection_urls", urls)
	_ = d.Set("handle", database.Handle)
	_ = d.Set("env_id", accountID)
	_ = d.Set("database_type", database.Type.Get())
	_ = d.Set("database_image_id", imageID)
	_ = d.Set("version", image.Version)

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
	client := meta.(*providerMetadata).Client
	legacy := meta.(*providerMetadata).LegacyClient
	databaseID := int32(d.Get("database_id").(int))
	containerSize := int32(d.Get("container_size").(int))
	profile := d.Get("container_profile").(string)
	diskSize := int32(d.Get("disk_size").(int))
	iops := int32(d.Get("iops").(int))
	handle := d.Get("handle").(string)
	var diags diag.Diagnostics

	ctx = meta.(*providerMetadata).APIContext(ctx)
	payload := aptibleapi.NewCreateOperationRequest("restart")

	if d.HasChange("container_size") {
		payload.SetContainerSize(containerSize)
	}
	if d.HasChange("iops") {
		payload.SetProvisionedIops(iops)
	}
	if d.HasChange("container_profile") {
		payload.SetInstanceProfile(profile)
	}
	if d.HasChange("disk_size") {
		payload.SetDiskSize(diskSize)
	}

	if d.HasChange("handle") {
		_, err := client.
			DatabasesAPI.
			PatchDatabase(ctx, databaseID).
			UpdateDatabaseRequest(
				aptibleapi.UpdateDatabaseRequest{Handle: &handle},
			).
			Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to update the database handle.",
				Detail:   generateErrorFromClientError(err).Error(),
			})
			return diags
		}
	}

	if d.HasChangeExcept("handle") {
		op, _, err := client.
			OperationsAPI.
			CreateOperationForDatabase(ctx, databaseID).
			CreateOperationRequest(*payload).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to update the database.",
				Detail:   generateErrorFromClientError(err).Error(),
			})
			return diags
		}

		del, err := legacy.WaitForOperation(int64(op.Id))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to update the database.",
				Detail:   generateErrorFromClientError(err).Error(),
			})
			return diags
		}
		if del {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("The replica with handle: %s was unexpectedly deleted", handle),
				Detail:   generateErrorFromClientError(err).Error(),
			})
			return diags
		}
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
