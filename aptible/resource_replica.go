package aptible

import (
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/aptible/go-deploy/aptible"
)

func resourceReplica() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicaCreate, // POST
		Read:   resourceReplicaRead,   // GET
		Update: resourceReplicaUpdate, // PUT
		Delete: resourceReplicaDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"primary_db_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1024,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"replica_id": {
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

func resourceReplicaCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	handle := d.Get("handle").(string)

	attrs := aptible.ReplicateAttrs{
		EnvID:         int64(d.Get("env_id").(int)),
		DatabaseID:    int64(d.Get("primary_db_id").(int)),
		ReplicaHandle: handle,
		ContainerSize: int64(d.Get("container_size").(int)),
		DiskSize:      int64(d.Get("disk_size").(int)),
	}

	replica, err := client.CreateReplica(attrs)
	if err != nil {
		log.Println(err)
		return err
	}

	d.Set("replica_id", replica.ID)
	d.SetId(handle)
	d.Set("connection_url", replica.ConnectionURL)
	return resourceReplicaRead(d, meta)
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceReplicaRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	replica_id := int64(d.Get("replica_id").(int))
	replica, deleted, err := client.GetReplica(replica_id)
	if deleted {
		d.SetId("")
		log.Println("Replica with ID: " + strconv.Itoa(int(replica_id)) + " was deleted outside of Terraform. \nNow removing it from Terraform state.")
		return nil
	}
	if err != nil {
		log.Println(err)
		return err
	}

	d.Set("container_size", replica.ContainerSize)
	d.Set("disk_size", replica.DiskSize)

	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceReplicaUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	replica_id := int64(d.Get("replica_id").(int))
	container_size := int64(d.Get("container_size").(int))
	disk_size := int64(d.Get("disk_size").(int))

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = container_size
	}

	if d.HasChange("disk_size") {
		updates.DiskSize = disk_size
	}

	err := client.UpdateReplica(replica_id, updates)
	if err != nil {
		return err
	}

	return resourceReplicaRead(d, meta)
}

func resourceReplicaDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	replica_id := int64(d.Get("replica_id").(int))
	err := client.DeleteReplica(replica_id)
	if err != nil {
		log.Println(err)
		return err
	}

	d.SetId("")
	return nil
}
