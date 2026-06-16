package aptible

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,        // POST
		ReadContext:   resourceDatabaseReadContext,   // GET
		UpdateContext: resourceDatabaseUpdate,        // PUT
		DeleteContext: resourceDatabaseDeleteContext, // DELETE
		Importer: &schema.ResourceImporter{
			State: resourceDatabaseImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
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
				StateFunc:    normalizeContainerProfile,
				Default:      "m",
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
			"enable_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	legacy := m.LegacyClient
	client := m.Client
	ctx = m.APIContext(ctx)
	diags := diag.Diagnostics{}

	envID := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)
	version := d.Get("version").(string)
	databaseType := d.Get("database_type").(string)
	iops := int32(d.Get("iops").(int))
	profile := d.Get("container_profile").(string)
	diskSize := int32(d.Get("disk_size").(int))
	containerSize := int32(d.Get("container_size").(int))
	enableBackups := d.Get("enable_backups").(bool)

	create := aptibleapi.NewCreateDatabaseRequest(handle, databaseType)

	if diskSize != 0 {
		create.SetInitialDiskSize(diskSize)
	}
	if containerSize != 0 {
		create.SetInitialContainerSize(containerSize)
	}
	if !enableBackups {
		create.SetEnableBackups(enableBackups)
	}

	if version != "" {
		image, err := legacy.GetDatabaseImageByTypeAndVersion(databaseType, version)
		if err != nil {
			log.Println(err)
			return generateDiagnosticsFromClientError(err)
		}
		create.SetDatabaseImageId(int32(image.ID))
	}

	db, _, err := client.DatabasesAPI.
		CreateDatabase(ctx, int32(envID)).
		CreateDatabaseRequest(*create).
		Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error creating database with handle: %s", handle),
			Detail:   err.Error(),
		})
	}

	_ = d.Set("database_id", db.Id)
	d.SetId(strconv.Itoa(int(db.Id)))

	payload := aptibleapi.NewCreateOperationRequest("provision")
	if diskSize != 0 {
		payload.SetDiskSize(diskSize)
	}
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
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error creating provision operation for database with handle: %s", handle),
			Detail:   err.Error(),
		})
	} else {
		createCtx, createCancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
		defer createCancel()
		_, err = waitForOperationWithContext(createCtx, legacy, int64(op.Id))
		if err != nil {
			// Do not return so that the read method can hydrate the state
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to provision database with handle: %s", handle),
				Detail:   err.Error(),
			})
		}
	}

	return append(diags, resourceDatabaseReadContext(ctx, d, meta)...)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.Client
	ctx = m.APIContext(ctx)
	databaseID := int32(d.Get("database_id").(int))

	database, resp, err := client.DatabasesAPI.GetDatabase(ctx, databaseID).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		log.Println("Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	urls := []string{}
	creds := database.Embedded.GetDatabaseCredentials()
	for _, cred := range creds {
		urls = append(urls, cred.ConnectionUrl)
	}

	imageID := ExtractIdFromLink(database.Links.DatabaseImage.GetHref())
	if imageID == 0 {
		return diag.FromErr(fmt.Errorf("Could not find database image ID"))
	}
	serviceID := ExtractIdFromLink(database.Links.Service.GetHref())
	if serviceID == 0 {
		return diag.FromErr(fmt.Errorf("Could not find database service ID"))
	}
	accountID := ExtractIdFromLink(database.Links.Account.GetHref())
	if accountID == 0 {
		return diag.FromErr(fmt.Errorf("Could not find database account ID"))
	}

	image, _, err := client.ImagesAPI.GetDatabaseImage(ctx, imageID).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	service, _, err := client.ServicesAPI.GetServiceWithOperationStatus(ctx, serviceID).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	containerSize := service.GetContainerMemoryLimitMb()
	profile := service.GetInstanceClass()

	_ = d.Set("container_size", containerSize)
	_ = d.Set("container_profile", normalizeContainerProfile(profile))
	_ = d.Set("iops", database.Embedded.Disk.GetProvisionedIops())
	_ = d.Set("disk_size", database.Embedded.Disk.GetSize())
	_ = d.Set("default_connection_url", database.GetConnectionUrl())
	_ = d.Set("connection_urls", urls)
	_ = d.Set("handle", database.GetHandle())
	_ = d.Set("env_id", accountID)
	_ = d.Set("database_type", database.Type.Get())
	_ = d.Set("database_image_id", imageID)
	_ = d.Set("version", image.GetVersion())
	_ = d.Set("enable_backups", database.EnableBackups)

	return nil
}

func resourceDatabaseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	databaseID, _ := strconv.Atoi(d.Id())
	_ = d.Set("database_id", databaseID)
	err := diagnosticsToError(resourceDatabaseReadContext(context.Background(), d, meta))
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
	enableBackups := d.Get("enable_backups").(bool)
	needsOperation := false
	var diags diag.Diagnostics

	ctx = meta.(*providerMetadata).APIContext(ctx)
	payload := aptibleapi.NewCreateOperationRequest("restart")

	if d.HasChange("container_size") {
		needsOperation = true
		payload.SetContainerSize(containerSize)
	}
	if d.HasChange("iops") {
		needsOperation = true
		payload.SetProvisionedIops(iops)
	}
	if d.HasChange("container_profile") {
		needsOperation = true
		payload.SetInstanceProfile(profile)
	}
	if d.HasChange("disk_size") {
		needsOperation = true
		payload.SetDiskSize(diskSize)
	}
	if d.HasChange("enable_backups") {
		_, err := client.
			DatabasesAPI.
			PatchDatabase(ctx, databaseID).
			UpdateDatabaseRequest(
				aptibleapi.UpdateDatabaseRequest{EnableBackups: &enableBackups},
			).
			Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to set enable_backups.",
				Detail:   err.Error(),
			})
			return diags
		}
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
				Detail:   err.Error(),
			})
			return diags
		}
	}

	if needsOperation {
		op, _, err := client.
			OperationsAPI.
			CreateOperationForDatabase(ctx, databaseID).
			CreateOperationRequest(*payload).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when trying to update the database.",
				Detail:   err.Error(),
			})
			return diags
		}

		updateCtx, updateCancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
		defer updateCancel()
		del, err := waitForOperationWithContext(updateCtx, legacy, int64(op.Id))
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
				Detail:   "The replica was unexpectedly deleted",
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

	return append(diags, resourceDatabaseReadContext(ctx, d, meta)...)
}

func resourceDatabaseDeleteContext(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).LegacyClient
	databaseID := int64(d.Get("database_id").(int))

	err := client.DeleteDatabase(databaseID)
	if err != nil {
		log.Println(err)
		return diag.FromErr(generateErrorFromClientError(err))
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
