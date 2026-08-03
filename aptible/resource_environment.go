package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aptible/aptible-api-go/aptibleapi"

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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
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
	if diags := validateBackupRetentionPolicy(d); diags != nil {
		return diags
	}

	m := meta.(*client)
	client := m.APIClient

	handle := d.Get("handle").(string)
	stackID := int32(d.Get("stack_id").(int))

	orgID := d.Get("org_id").(string)
	if orgID == "" {
		// Look up org from auth API
		orgID, _ = m.GetOrganizationFromAuthAPI()
	}

	if orgID == "" {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Unable to determine organization ID",
			Detail:   "Unable to infer organization ID. You may have to specify org_id explicitly.",
		}}
	}

	// Determine environment type based on whether stack is shared (public)
	stack, _, err := client.StacksAPI.GetStack(ctx, stackID).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching stack: %w", err))
	}

	envType := "production"
	if stack.Public {
		envType = "development"
	}

	account, _, err := client.AccountsAPI.CreateAccount(ctx).
		CreateAccountRequest(aptibleapi.CreateAccountRequest{
			Type:           envType,
			Handle:         handle,
			OrganizationId: orgID,
			StackId:        &stackID,
		}).Execute()
	if err != nil {
		log.Println("There was an error when completing the request to create the environment.\n[ERROR] -", err)
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(int(account.Id)))
	_ = d.Set("env_id", int(account.Id))

	if diags := createBackupRetentionPolicy(ctx, d, meta); diags != nil {
		return diags
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*client)
	client := m.APIClient

	envID := int32(d.Get("env_id").(int))
	log.Println("Getting environment with ID: " + strconv.Itoa(int(envID)))

	account, resp, err := client.AccountsAPI.GetAccount(ctx, envID).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		log.Println(err)
		return diag.FromErr(err)
	}

	_ = d.Set("handle", account.Handle)
	_ = d.Set("env_id", int(account.Id))

	if account.Links != nil {
		if account.Links.Stack != nil && account.Links.Stack.Href != nil {
			_ = d.Set("stack_id", int(ExtractIdFromLink(*account.Links.Stack.Href)))
		}
		if account.Links.Organization != nil && account.Links.Organization.Href != nil {
			href := *account.Links.Organization.Href
			segments := strings.Split(href, "/")
			if len(segments) > 0 {
				_ = d.Set("org_id", segments[len(segments)-1])
			}
		}
	}

	return readBackupRetentionPolicy(ctx, d, meta)
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if diags := validateBackupRetentionPolicy(d); diags != nil {
		return diags
	}

	m := meta.(*client)
	client := m.APIClient

	handle := d.Get("handle").(string)
	envID := int32(d.Get("env_id").(int))

	_, err := client.AccountsAPI.UpdateAccount(ctx, envID).
		UpdateAccountRequest(aptibleapi.UpdateAccountRequest{
			Handle: &handle,
		}).Execute()
	if err != nil {
		log.Println("There was an error when completing the request to update the environment.\n[ERROR] -", err)
		return diag.FromErr(err)
	}

	if diags := createBackupRetentionPolicy(ctx, d, meta); diags != nil {
		return diags
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	readDiags := resourceEnvironmentRead(ctx, d, meta)
	if !readDiags.HasError() {
		m := meta.(*client)
		client := m.APIClient
		envID := int32(d.Get("env_id").(int))

		// First deprovision any tail log drains
		log.Println("Checking for tail type log drains for environment ID: ", envID)
		drainResp, _, listErr := client.LogDrainsAPI.ListLogDrainsForAccount(ctx, envID).Execute()
		if listErr != nil {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Error fetching log drains",
				Detail:   listErr.Error(),
			}}
		}

		for _, drain := range drainResp.Embedded.LogDrains {
			if drain.DrainType == "tail" {
				_, drainErr := m.DeleteLogDrain(ctx, drain.Id)
				if drainErr != nil {
					log.Println("There was an error when completing the request to destroy the log drain.\n[ERROR] -", drainErr)
					return diag.FromErr(drainErr)
				}
			}
		}

		// Delete the environment
		_, err := client.AccountsAPI.DeleteAccount(ctx, envID).Execute()
		if err != nil {
			log.Println("There was an error when completing the request to destroy the environment.\n[ERROR] -", err)
			return diag.FromErr(err)
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
	m := meta.(*client)
	client := m.APIClient
	envId := int32(d.Get("env_id").(int))

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
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Error creating backup retention policy",
			Detail:   err.Error(),
		}}
	}

	return nil
}

func readBackupRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*client)
	client := m.APIClient
	envId := int32(d.Get("env_id").(int))

	log.Printf("Getting backup retention policy for environment with ID: %d\n", envId)

	resp, _, err := client.BackupRetentionPoliciesAPI.
		ListBackupRetentionPoliciesForAccount(ctx, envId).
		Execute()
	if err != nil {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Error fetching backup retention policy",
			Detail:   err.Error(),
		}}
	}

	policies := resp.Embedded.BackupRetentionPolicies
	if len(policies) == 0 {
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
