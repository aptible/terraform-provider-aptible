package aptible

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate, // POST
		ReadContext:   resourceDatabaseRead,   // GET
		UpdateContext: resourceDatabaseUpdate, // PUT
		DeleteContext: resourceDatabaseDelete, // DELETE
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseImport,
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
				ValidateFunc: validation.IntInSlice(validContainerSizes),
				Default:      1024,
			},
			"disk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 16000),
				Default:      10,
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

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	envID := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)
	version := d.Get("version").(string)
	databaseType := d.Get("database_type").(string)
	var diags diag.Diagnostics

	attrs := aptible.DBCreateAttrs{
		Handle:        &handle,
		Type:          databaseType,
		ContainerSize: int64(d.Get("container_size").(int)),
		DiskSize:      int64(d.Get("disk_size").(int)),
	}

	if version != "" {
		image, err := client.GetDatabaseImageByTypeAndVersion(databaseType, version)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "There was an error when completing the request.",
				Detail:   "There was an error when trying to get database images by type and version.",
			})
			log.Println(err)
			return diags
		}
		attrs.DatabaseImageID = image.ID
	}

	database, err := client.CreateDatabase(envID, attrs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when completing the request.",
			Detail:   "There was an error when trying to create the database.",
		})
		log.Println(err)
		return diags
	}

	_ = d.Set("database_id", database.ID)

	d.SetId(strconv.Itoa(int(database.ID)))
	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("database_id").(int))
	var diags diag.Diagnostics

	database, err := client.GetDatabase(databaseID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when completing the request.",
			Detail:   "There was an error when trying to find the database.",
		})
		log.Println(err)
		return diags
	}

	if database.Deleted {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Database was deleted outside terraform.",
			Detail:   "Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.",
		})
		log.Println("Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return diags
	}

	_ = d.Set("container_size", database.ContainerSize)
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

func resourceDatabaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	databaseID, _ := strconv.Atoi(d.Id())
	_ = d.Set("database_id", databaseID)
	var diags diag.Diagnostics

	diags = append(diags, resourceDatabaseRead(ctx, d, meta)...)
	if diags.HasError() {
		return nil, errors.New("unable to read existing resources")
	}
	return []*schema.ResourceData{d}, nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("database_id").(int))
	containerSize := int64(d.Get("container_size").(int))
	diskSize := int64(d.Get("disk_size").(int))
	handle := d.Get("handle").(string)
	var diags diag.Diagnostics

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = containerSize
	}

	if d.HasChange("disk_size") {
		updates.DiskSize = diskSize
	}

	if d.HasChange("handle") {
		log.Printf("Updating handle to %s\n", handle)
		updates.Handle = handle
	}

	if !d.HasChangesExcept("handle") {
		updates.OnlyChangingHandle = true
	}

	err := client.UpdateDatabase(databaseID, updates)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when completing the request.",
			Detail:   "There was an error when trying to update the database.",
		})
		return diags
	}

	if d.HasChange("handle") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "You must restart the database to see changes",
			Detail:   fmt.Sprintf("In order for the new database name (%s) to appear in log drain and metric drain destinations, you must restart the database.", handle),
		})
		log.Printf(fmt.Sprintf("[WARN] In order for the new database name (%s) to appear in log drain and metric drain destinations, you must restart the database.", handle))
	}

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("database_id").(int))
	var diags diag.Diagnostics

	err := client.DeleteDatabase(databaseID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when completing the request.",
			Detail:   "There was an error when trying to delete the database.",
		})
		log.Println(err)
		return diags
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
