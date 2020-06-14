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
			"db_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validDBTypes, false),
				Default:      "postgresql",
				ForceNew:     true,
			},
			"container_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(512, 7168),
				Default:      1024,
			},
			"disk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(10, 200),
				Default:      10,
			},
			"db_id": {
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
		Type:          d.Get("db_type").(string),
		ContainerSize: int64(d.Get("container_size").(int)),
		DiskSize:      int64(d.Get("disk_size").(int)),
	}

	database, err := client.CreateDatabase(envID, attrs)
	if err != nil {
		log.Println(err)
		return err
	}

	_ = d.Set("db_id", database.ID)

	d.SetId(handle)
	return resourceDatabaseRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("db_id").(int))

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
	d.SetId(database.Handle)

	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	databaseID := int64(d.Get("db_id").(int))
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
	databaseID := int64(d.Get("db_id").(int))

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
