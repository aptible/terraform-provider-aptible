package aptible

import (
	"log"
	"strconv"

	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceReplica() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicaCreate, // POST
		Read:   resourceReplicaRead,   // GET
		Update: resourceReplicaUpdate, // PUT
		Delete: resourceReplicaDelete, // DELETE
		Importer: &schema.ResourceImporter{
			State: resourceReplicaImport,
		},

		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"primary_database_id": {
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
			"replica_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_connection_url": {
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
		DatabaseID:    int64(d.Get("primary_database_id").(int)),
		ReplicaHandle: handle,
		ContainerSize: int64(d.Get("container_size").(int)),
		DiskSize:      int64(d.Get("disk_size").(int)),
	}

	replica, err := client.CreateReplica(attrs)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}

	_ = d.Set("replica_id", replica.ID)
	d.SetId(strconv.Itoa(int(replica.ID)))
	_ = d.Set("default_connection_url", replica.DefaultConnection)
	return resourceReplicaRead(d, meta)
}

func resourceReplicaImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	replicaID, _ := strconv.Atoi(d.Id())
	_ = d.Set("replica_id", replicaID)
	err := resourceReplicaRead(d, meta)
	return []*schema.ResourceData{d}, err
}

// syncs Terraform state with changes made via the API outside of Terraform
func resourceReplicaRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	replicaID := int64(d.Get("replica_id").(int))
	replica, err := client.GetReplica(replicaID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}
	if replica.Deleted {
		d.SetId("")
		log.Println("Replica with ID: " + strconv.Itoa(int(replicaID)) + " was deleted outside of Terraform. \nNow removing it from Terraform state.")
		return nil
	}

	_ = d.Set("container_size", replica.ContainerSize)
	_ = d.Set("disk_size", replica.DiskSize)
	_ = d.Set("default_connection_url", replica.DefaultConnection)
	_ = d.Set("handle", replica.Handle)
	_ = d.Set("env_id", replica.EnvironmentID)
	_ = d.Set("primary_database_id", replica.InitializeFromID)
	d.SetId(strconv.Itoa(int(replica.ID)))

	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceReplicaUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	replicaID := int64(d.Get("replica_id").(int))
	containerSize := int64(d.Get("container_size").(int))
	diskSize := int64(d.Get("disk_size").(int))

	updates := aptible.DBUpdates{}

	if d.HasChange("container_size") {
		updates.ContainerSize = containerSize
	}

	if d.HasChange("disk_size") {
		updates.DiskSize = diskSize
	}

	err := client.UpdateReplica(replicaID, updates)
	if err != nil {
		return generateErrorFromClientError(err)
	}

	return resourceReplicaRead(d, meta)
}

func resourceReplicaDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*aptible.Client)
	replicaID := int64(d.Get("replica_id").(int))
	err := client.DeleteReplica(replicaID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}

	d.SetId("")
	return nil
}
