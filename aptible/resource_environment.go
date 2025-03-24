package aptible

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/go-deploy/aptible"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEnvironmentImport,
		},
		Schema: map[string]*schema.Schema{
			"env_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"org_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"stack_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"backup_retention_policy": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"monthly": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"yearly": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"make_copy": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"keep_final": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Sets don't support validation functions at this time so validate on apply
	if diags := validateBackupRetentionPolicy(d); diags != nil {
		return diags
	}

	client := meta.(*providerMetadata).LegacyClient
	handle := d.Get("handle").(string)
	stackID := int64(d.Get("stack_id").(int))

	// there are a few  scenarios where org_id can be fully gathered from the data provided above
	// 1. If it is provided explicitly
	// 2. It is using a dedicated stack, and so when creating an environment, use the stack id to get stack
	//    and check that stack's org id and use it
	// 3. The user only belongs to one organization, so fall back to that. If they belong to multiple organizations
	// 	  they will have to specify it explicitly at this point (if these points no longer work for inference.)
	orgID := d.Get("org_id").(string) // scenario #1 outlined above
	if orgID == "" {
		stack, err := client.GetStack(stackID) // scenario #2 outlined above
		if err != nil {
			log.Println("There was an error trying to retrieve the stack with the stack id provided to determine"+
				"an organization id.\n[ERROR] - ", err)
			return generateDiagnosticsFromClientError(err)
		}
		orgID = stack.OrganizationID
		if orgID == "" {
			org, err := client.GetOrganization() // scenario #3 outlined above
			if err != nil {
				log.Println("There was an error trying to retrieve an organization id (org_id). You can "+
					"either specify it explicitly or review the error message to attempt to fix the issue. "+
					"\n[ERROR] - ", err)
				return generateDiagnosticsFromClientError(err)
			}
			orgID = org.ID
		}
	}

	if orgID == "" {
		errorMessage := "[ERROR] - Unable to infer organization ID from stack or user. You may have to specify it explicitly"
		log.Println(errorMessage)
		return generateDiagnosticsFromClientError(errors.New(errorMessage))
	}

	data := aptible.EnvironmentCreateAttrs{
		Handle: handle,
	}

	environment, err := client.CreateEnvironment(orgID, stackID, data)
	if err != nil {
		log.Println("There was an error when completing the request to create the environment.\n[ERROR] -", err)
		return generateDiagnosticsFromClientError(err)
	}

	d.SetId(strconv.Itoa(int(environment.ID)))
	_ = d.Set("env_id", environment.ID)

	if diags := createBackupRetentionPolicy(ctx, d, meta); diags != nil {
		return diags
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerMetadata).LegacyClient
	envID := int64(d.Get("env_id").(int))

	log.Println("Getting environment with ID: " + strconv.Itoa(int(envID)))

	environment, err := client.GetEnvironment(envID)
	if err != nil {
		log.Println(err)
		return generateDiagnosticsFromClientError(err)
	}
	if environment.Deleted {
		d.SetId("")
		return nil
	}

	_ = d.Set("handle", environment.Handle)
	_ = d.Set("env_id", int(environment.ID))
	_ = d.Set("stack_id", environment.StackID)
	_ = d.Set("org_id", environment.OrganizationID)

	return readBackupRetentionPolicy(ctx, d, meta)
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Sets don't support validation functions at this time so validate on apply
	if diags := validateBackupRetentionPolicy(d); diags != nil {
		return diags
	}

	client := meta.(*providerMetadata).LegacyClient
	handle := d.Get("handle").(string)
	envId := int64(d.Get("env_id").(int))
	environmentUpdates := aptible.EnvironmentUpdates{
		Handle: handle,
	}

	if err := client.UpdateEnvironment(envId, environmentUpdates); err != nil {
		log.Println("There was an error when completing the request to update the environment.\n[ERROR] -", err)
		return generateDiagnosticsFromClientError(err)
	}

	// Creating a new backup retention policy replaces the existing one
	if diags := createBackupRetentionPolicy(ctx, d, meta); diags != nil {
		return diags
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	readDiags := resourceEnvironmentRead(ctx, d, meta)
	if !readDiags.HasError() {
		envID := int64(d.Get("env_id").(int))
		client := meta.(*providerMetadata).LegacyClient

		// First we need to run deprovision operations on any tail drains
		log.Println("Checking for an tail type log drain for environment ID: ", envID)

		resp, err := client.ListLogDrainsForAccount(envID)

		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Error fetching log drains",
					Detail:   err.Error(),
				},
			}
		}

		drains := resp.Embedded.LogDrains
		if len(drains) == 0 {
			log.Println("No log drains found")
			return nil
		} else {
			for _, drain := range drains {
				if drain.DrainType != "tail" {
					log.Println("Found drain of unexpected type: ", drain.DrainType)
					return nil
				}

				_, err := client.DeleteLogDrain(drain.ID)

				if err != nil {
					log.Println("There was an error when completing the request to destroy the log drain.\n[ERROR] -", err)
					return generateErrorFromClientError(err)
				}
			}
		}

		// Now we should be okay to delete the environment.
		err := client.DeleteEnvironment(envID)
		if err != nil {
			log.Println("There was an error when completing the request to destroy the environment.\n[ERROR] -", err)
			return generateDiagnosticsFromClientError(err)
		}

		d.SetId("")
		return nil
	}
	d.SetId("")
	return nil
}

func resourceEnvironmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	envID, _ := strconv.Atoi(d.Id())
	_ = d.Set("env_id", envID)
	if err := diagnosticsToError(resourceEnvironmentRead(ctx, d, meta)); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// Backup retention policy
func validateBackupRetentionPolicy(d *schema.ResourceData) diag.Diagnostics {
	if d.Get("backup_retention_policy").(*schema.Set).Len() > 1 {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Multiple backup_retention_policy",
				Detail:   "Environments may only have one backup retention policy",
			},
		}
	}

	return nil
}

func createBackupRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*providerMetadata)
	client := m.Client
	ctx = m.APIContext(ctx)
	envId := int32(d.Get("env_id").(int))

	// Only modify the policy if it has changed
	if !d.HasChange("backup_retention_policy") {
		log.Println("No change in retention policy detected")
		return nil
	}

	policies := d.Get("backup_retention_policy").(*schema.Set).List()
	if len(policies) < 1 {
		return nil
	}
	policy := policies[0].(map[string]interface{})

	daily := int32(policy["daily"].(int))
	monthly := int32(policy["monthly"].(int))
	yearly := int32(policy["yearly"].(int))
	makeCopy := policy["make_copy"].(bool)
	keepFinal := policy["keep_final"].(bool)

	_, err := client.BackupRetentionPoliciesAPI.
		CreateBackupRetentionPolicy(ctx, envId).
		CreateBackupRetentionPolicyRequest(aptibleapi.CreateBackupRetentionPolicyRequest{
			Daily:     &daily,
			Monthly:   &monthly,
			Yearly:    &yearly,
			MakeCopy:  &makeCopy,
			KeepFinal: &keepFinal,
		}).
		Execute()
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error creating backup retention policy",
				Detail:   err.Error(),
			},
		}
	}

	return nil
}

func readBackupRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	met := meta.(*providerMetadata)
	client := met.Client
	ctx = met.APIContext(ctx)
	envId := int32(d.Get("env_id").(int))

	log.Printf("Getting backup retention policy for environment with ID: %d\n", envId)

	resp, _, err := client.BackupRetentionPoliciesAPI.
		ListBackupRetentionPoliciesForAccount(ctx, envId).
		Execute()
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error fetching backup retention policy",
				Detail:   err.Error(),
			},
		}
	}

	policies := resp.Embedded.BackupRetentionPolicies
	if len(policies) == 0 {
		// No results, the environment's policy must have been deleted
		return nil
	}

	policy := policies[0]
	policyData := make(map[string]interface{})

	policyData["daily"] = int(policy.Daily)
	policyData["monthly"] = int(policy.Monthly)
	policyData["yearly"] = int(policy.Yearly)
	policyData["make_copy"] = policy.MakeCopy
	policyData["keep_final"] = policy.KeepFinal

	_ = d.Set("backup_retention_policy", []map[string]interface{}{policyData})

	return nil
}
