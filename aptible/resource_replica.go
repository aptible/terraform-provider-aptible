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
	// Linter gets upset because of the mixed context and non-context methods
	// lintignore:S024
	return &schema.Resource{
		CreateContext: resourceReplicaCreate, // POST
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
			"enable_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	legacy := m.LegacyClient
	client := m.Client
	ctx = m.APIContext(ctx)
	diags := diag.Diagnostics{}

	handle := d.Get("handle").(string)
	databaseID := int32(d.Get("primary_database_id").(int))
	iops := int32(d.Get("iops").(int))
	containerSize := int32(d.Get("container_size").(int))
	diskSize := int32(d.Get("disk_size").(int))
	profile := d.Get("container_profile").(string)
	enableBackups := d.Get("enable_backups").(bool)
	envID := int32(d.Get("env_id").(int))

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
	if d.HasChange("disk_size") {
		payload.SetDiskSize(diskSize)
	}

	op, _, err := client.
		OperationsAPI.
		CreateOperationForDatabase(ctx, databaseID).
		CreateOperationRequest(*payload).
		Execute()
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create replicate operation",
			Detail:   err.Error(),
		})
	}

	deleted, err := legacy.WaitForOperation(int64(op.Id))
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to replicate database",
			Detail:   err.Error(),
		})
	}
	if deleted {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to replicate database",
			Detail:   fmt.Sprintf("The database with ID: %d was unexpectedly deleted", databaseID),
		})
	}

	repl, err := legacy.GetReplicaFromHandle(int64(databaseID), handle)
	if err != nil {
		return generateDiagnosticsFromClientError(err)
	}

	// At this point the replica exists so it should be persisted in the state
	_ = d.Set("replica_id", repl.ID)
	d.SetId(strconv.Itoa(int(repl.ID)))

	operation := repl.Embedded.LastOperation
	if operation == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Replica provision operation not found",
		})
		return append(diags, diag.FromErr(resourceReplicaRead(d, meta))...)
	}
	operationID := (*operation).ID
	deleted, err = legacy.WaitForOperation(operationID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to provision replica",
			Detail:   err.Error(),
		})
		return append(diags, diag.FromErr(resourceReplicaRead(d, meta))...)
	}
	if deleted {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to provision replica",
			Detail:   fmt.Sprintf("the replica with handle: %s was unexpectedly deleted", handle),
		})
		return append(diags, diag.FromErr(resourceReplicaRead(d, meta))...)
	}

	// Enable backups defaults to `true` in the backend so we only need to do
	// something if it is being set to false.
	// A replica is created in the backend via sweetness, so we need to wait until
	// after the replication is complete to update the db object.
	if !enableBackups {
		_, err := client.
			DatabasesAPI.
			PatchDatabase(ctx, int32(repl.ID)).
			UpdateDatabaseRequest(
				aptibleapi.UpdateDatabaseRequest{EnableBackups: &enableBackups},
			).
			Execute()
		if err != nil {
			// Making this a warning as the resource is not broken so we don't want it
			// to be tainted (replaced on next apply). Rather, allowing the create to
			// complete will result in a non-empty plan after apply as hydrating the
			// state using the read method will set enable_backups to true which is
			// inconsistent with the desired state. Subsequent apply's will attempt to
			// update the replica to match the desired state.
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Error disabling backups on replica with handle: %s", handle),
				Detail:   err.Error(),
			})
		}
	}

	return append(diags, diag.FromErr(resourceReplicaRead(d, meta))...)
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
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		log.Println("Database with ID: " + strconv.Itoa(int(databaseID)) + " was deleted outside of Terraform. Now removing it from Terraform state.")
		return nil
	}
	if err != nil {
		return err
	}

	imageID := ExtractIdFromLink(database.Links.DatabaseImage.GetHref())
	if imageID == 0 {
		return fmt.Errorf("could not find database image ID")
	}
	serviceID := ExtractIdFromLink(database.Links.Service.GetHref())
	if serviceID == 0 {
		return fmt.Errorf("could not find database service ID")
	}
	accountID := ExtractIdFromLink(database.Links.Account.GetHref())
	if accountID == 0 {
		return fmt.Errorf("could not find database account ID")
	}

	service, _, err := client.ServicesAPI.GetServiceWithOperationStatus(ctx, serviceID).Execute()
	if err != nil {
		return err
	}

	containerSize := service.GetContainerMemoryLimitMb()
	profile := service.GetInstanceClass()
	primaryDatabaseID := ExtractIdFromLink(database.Links.InitializeFrom.GetHref())

	_ = d.Set("container_size", containerSize)
	_ = d.Set("disk_size", database.Embedded.Disk.GetSize())
	_ = d.Set("default_connection_url", database.GetConnectionUrl())
	_ = d.Set("handle", database.GetHandle())
	_ = d.Set("env_id", accountID)
	_ = d.Set("primary_database_id", primaryDatabaseID)
	_ = d.Set("container_profile", profile)
	_ = d.Set("iops", database.Embedded.Disk.GetProvisionedIops())
	_ = d.Set("enable_backups", database.GetEnableBackups())
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
				Detail:   "The database was unexpectedly deleted",
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
