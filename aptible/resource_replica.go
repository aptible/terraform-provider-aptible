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
)

func resourceReplica() *schema.Resource {
	return &schema.Resource{
		Create:        resourceReplicaCreate, // POST
		Read:          resourceReplicaRead,   // GET
		UpdateContext: resourceReplicaUpdate, // PUT
		Delete:        resourceReplicaDelete, // DELETE
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
				Default:  3000,
			},
			"replica_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_connection_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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
	ctx = meta.(*providerMetadata).APIContext(ctx)

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

	op, _, err := client.
		OperationsAPI.
		CreateOperationForDatabase(ctx, databaseID).
		CreateOperationRequest(*payload).
		Execute()
	if err != nil {
		return err
	}

	deleted, err := legacy.WaitForOperation(int64(op.Id))
	if err != nil {
		return generateErrorFromClientError(err)
	}
	if deleted {
		return fmt.Errorf("the replica with handle: %s was unexpectedly deleted", handle)
	}

	repl, err := legacy.GetReplicaFromHandle(int64(databaseID), handle)
	if err != nil {
		return generateErrorFromClientError(err)
	}

	operation := repl.Embedded.LastOperation
	if operation == nil {
		return fmt.Errorf("Could not find provision operation for replica database")
	}
	operationID := (*operation).ID
	deleted, err = legacy.WaitForOperation(operationID)
	if err != nil {
		return generateErrorFromClientError(err)
	}
	if deleted {
		return fmt.Errorf("the replica with handle: %s was unexpectedly deleted", handle)
	}

	_ = d.Set("replica_id", repl.ID)
	d.SetId(strconv.Itoa(int(repl.ID)))
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

	imageID := ExtractIdFromLink(database.Links.DatabaseImage.GetHref())
	if imageID == 0 {
		return fmt.Errorf("Could not find database image ID")
	}
	serviceID := ExtractIdFromLink(database.Links.Service.GetHref())
	if serviceID == 0 {
		return fmt.Errorf("Could not find database service ID")
	}
	accountID := ExtractIdFromLink(database.Links.Account.GetHref())
	if accountID == 0 {
		return fmt.Errorf("Could not find database account ID")
	}

	service, _, err := client.ServicesAPI.GetService(ctx, serviceID).Execute()
	if err != nil {
		return generateErrorFromClientError(err)
	}

	containerSize := service.ContainerMemoryLimitMb.Get()
	profile := service.InstanceClass
	primaryDatabaseID := ExtractIdFromLink(database.Links.InitializeFrom.GetHref())

	_ = d.Set("container_size", containerSize)
	_ = d.Set("disk_size", database.Embedded.Disk.GetSize())
	_ = d.Set("default_connection_url", database.GetConnectionUrl())
	_ = d.Set("handle", database.GetHandle())
	_ = d.Set("env_id", accountID)
	_ = d.Set("primary_database_id", primaryDatabaseID)
	_ = d.Set("container_profile", profile)
	_ = d.Set("iops", database.Embedded.Disk.GetProvisionedIops())
	d.SetId(strconv.Itoa(int(database.Id)))

	return nil
}

// changes state of actual resource based on changes made in a Terraform config file
func resourceReplicaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).Client
	legacy := meta.(*providerMetadata).LegacyClient
	ctx = meta.(*providerMetadata).APIContext(ctx)
	databaseID := int32(d.Get("replica_id").(int))
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

	if err := resourceReplicaRead(d, meta); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when trying to retrieve the updated state of the database.",
			Detail:   err.Error(),
		})
	}

	return diags
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
