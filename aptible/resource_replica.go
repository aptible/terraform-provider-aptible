package aptible

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				ValidateFunc: validateContainerSize,
				Default:      1024,
			},
			"disk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateDiskSize,
				Default:      10,
			},
			"container_profile": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateContainerProfile,
				Default:      "m5",
			},
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
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
	client := meta.(*providerMetadata).Client
	legacy := meta.(*providerMetadata).LegacyClient
	handle := d.Get("handle").(string)
	databaseID := int32(d.Get("primary_database_id").(int))
	iops := int32(d.Get("iops").(int))
	containerSize := int32(d.Get("container_size").(int))
	diskSize := int32(d.Get("disk_size").(int))
	profile := d.Get("container_profile").(string)
	envID := int32(d.Get("env_id").(int))
	ctx := context.Background()

	createOp := client.OperationsAPI.CreateOperationForDatabase(ctx, databaseID)
	payload := aptibleapi.NewCreateOperationRequest("replicate")
	payload.SetHandle(handle)
	payload.SetDestinationAccountId(envID)
	if iops != 0 {
		payload.SetProvisionedIops(iops)
	}
	if profile != "" {
		payload.SetInstanceProfile(profile)
	}
	if containerSize != 0 {
		payload.SetContainerSize(containerSize)
	}
	if diskSize != 0 {
		payload.SetDiskSize(diskSize)
	}

	op, _, err := createOp.CreateOperationRequest(*payload).Execute()
	if err != nil {
		return err
	}

	deleted, err := legacy.WaitForOperation(int64(op.Id))
	if err != nil {
		return err
	}

	repl, err := legacy.GetReplicaFromHandle(int64(databaseID), handle)
	if err != nil {
		return generateErrorFromClientError(err)
	}
	replica, err := legacy.GetReplica(repl.ID)
	if deleted {
		return fmt.Errorf("the replica with handle: %s was unexpectedly deleted", handle)
	}
	if err != nil {
		return err
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
	databaseID := int32(d.Get("replica_id").(int))

	client := meta.(*providerMetadata).Client
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
	primaryDatabaseID := ExtractIdFromLink(*database.Links.GetInitializeFrom().Href)

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
	_ = d.Set("primary_database_id", primaryDatabaseID)

	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceReplicaUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerMetadata).LegacyClient
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
	client := meta.(*providerMetadata).LegacyClient
	replicaID := int64(d.Get("replica_id").(int))
	err := client.DeleteReplica(replicaID)
	if err != nil {
		log.Println(err)
		return generateErrorFromClientError(err)
	}

	d.SetId("")
	return nil
}
