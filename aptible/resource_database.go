package aptible

import (
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate, // POST
		Read:   resourceDatabaseRead,   // GET
		Update: resourceDatabaseUpdate, // PUT
		Delete: resourceDatabaseDelete, // DELETE
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
				ForceNew: true,
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
			"connection_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	envID := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)

	attrs := aptible.DBCreateAttrs{
		Handle:        &handle,
		Type:          d.Get("database_type").(string),
		ContainerSize: int64(d.Get("container_size").(int)),
		DiskSize:      int64(d.Get("disk_size").(int)),
	}

	database, err := client.CreateDatabase(envID, attrs)
	if err != nil {
		log.Println(err)
		return err
	}

	_ = d.Set("database_id", database.ID)

	d.SetId(handle)
	return resourceDatabaseRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("database_id").(int))

	database, err := client.GetDatabase(databaseID)
	if err != nil {
		log.Println(err)
		return err
	}

	if database.Deleted {
		d.SetId("")
		log.Println("Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	_ = d.Set("container_size", database.ContainerSize)
	_ = d.Set("disk_size", database.DiskSize)
	_ = d.Set("connection_url", database.ConnectionURL)
	_ = d.Set("handle", database.Handle)
	_ = d.Set("env_id", database.EnvironmentID)
	_ = d.Set("database_type", database.Type)

	d.SetId(database.Handle)

	return nil
}

func resourceDatabaseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	databaseID, _ := strconv.Atoi(d.Id())
	_ = d.Set("database_id", databaseID)
	err := resourceDatabaseRead(d, meta)
	return []*schema.ResourceData{d}, err
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("database_id").(int))
	containerSize := int64(d.Get("container_size").(int))
	diskSize := int64(d.Get("disk_size").(int))

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = containerSize
	}

	if d.HasChange("disk_size") {
		updates.DiskSize = diskSize
	}

	err := client.UpdateDatabase(databaseID, updates)
	if err != nil {
		return err
	}

	return resourceDatabaseRead(d, meta)
}

func resourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("database_id").(int))

	err := client.DeleteDatabase(databaseID)
	if err != nil {
		log.Println(err)
		return err
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
