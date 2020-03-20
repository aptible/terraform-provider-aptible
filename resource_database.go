package main

import (
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aptible/go-deploy/aptible"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate, // POST
		Read:   resourceDatabaseRead,   // GET
		Update: resourceDatabaseUpdate, // PUT
		Delete: resourceDatabaseDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"env_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "postgresql",
				ForceNew: true,
			},
			"container_size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1024,
			},
			"disk_size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"db_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aptible.Client)
	env_id := int64(d.Get("env_id").(int))
	handle := d.Get("handle").(string)
	db_type := d.Get("db_type").(string)
	container_size := int64(d.Get("container_size").(int))
	disk_size := int64(d.Get("disk_size").(int))

	attrs := aptible.DBCreateAttrs{
		Handle:        &handle,
		Type:          db_type,
		ContainerSize: container_size,
		DiskSize:      disk_size,
	}

	payload, err := client.CreateDatabase(env_id, attrs)
	if err != nil {
		AppLogger.Println(err)
		return err
	}

	d.Set("db_id", *payload.ID)
	d.Set("connection_url", *payload.ConnectionURL)
	d.SetId(handle)
	return resourceDatabaseRead(d, m)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aptible.Client)
	db_id := int64(d.Get("db_id").(int))
	payload, deleted, err := client.GetDatabase(db_id)
	if err != nil {
		AppLogger.Println(err)
		return err
	}
	if deleted {
		d.SetId("")
		AppLogger.Println("Database with ID: " + strconv.Itoa(int(db_id)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}

	op := payload.Embedded.LastOperation
	d.Set("container_size", op.ContainerSize)
	d.Set("disk_size", op.DiskSize)
	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceDatabaseUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aptible.Client)
	db_id := int64(d.Get("db_id").(int))
	container_size := int64(d.Get("container_size").(int))
	disk_size := int64(d.Get("disk_size").(int))

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = container_size
	}

	if d.HasChange("disk_size") {
		updates.DiskSize = disk_size
	}

	err := client.UpdateDatabase(db_id, updates)
	if err != nil {
		return err
	}

	return resourceDatabaseRead(d, m)
}

func resourceDatabaseDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aptible.Client)
	db_id := int64(d.Get("db_id").(int))
	err := client.DeleteDatabase(db_id)
	if err != nil {
		AppLogger.Println(err)
		return err
	}

	d.SetId("")
	return nil
}
