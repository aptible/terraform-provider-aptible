package aptible

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBackupRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupRetentionPolicyRead,
		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"daily": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"monthly": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"yearly": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"make_copy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"keep_final": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceBackupRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*providerMetadata)
	client := meta.Client
	ctx = meta.APIContext(ctx)

	// Policies are identified by environment ID
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
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error fetching backup retention policy",
				Detail:   fmt.Sprintf("Environment with ID %v does not have a backup retention policy", envId),
			},
		}
	}

	policy := policies[0]

	d.SetId(strconv.Itoa(int(envId)))
	_ = d.Set("policy_id", int(policy.Id))
	_ = d.Set("env_id", int(envId))
	_ = d.Set("daily", int(policy.Daily))
	_ = d.Set("monthly", int(policy.Monthly))
	_ = d.Set("yearly", int(policy.Yearly))
	_ = d.Set("make_copy", policy.MakeCopy)
	_ = d.Set("keep_final", policy.KeepFinal)

	return nil
}
