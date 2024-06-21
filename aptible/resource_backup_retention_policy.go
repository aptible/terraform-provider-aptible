package aptible

import (
	"context"
	"log"
	"strconv"

	"github.com/aptible/aptible-api-go/aptibleapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBackupRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupRetentionPolicyCreate,
		ReadContext:   resourceBackupRetentionPolicyRead,
		UpdateContext: resourceBackupRetentionPolicyCreate,
		DeleteContext: resourceBackupRetentionPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBackupRetentionPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"env_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
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
	}
}

func resourceBackupRetentionPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*providerMetadata)
	client := meta.Client
	ctx = meta.APIContext(ctx)

	envId := int32(d.Get("env_id").(int))
	daily := int32(d.Get("daily").(int))
	monthly := int32(d.Get("monthly").(int))
	yearly := int32(d.Get("yearly").(int))
	makeCopy := d.Get("make_copy").(bool)
	keepFinal := d.Get("keep_final").(bool)

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

	return resourceBackupRetentionPolicyRead(ctx, d, meta)
}

func resourceBackupRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		d.SetId("")
		return nil
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

func resourceBackupRetentionPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// The environment must have a backup retention policy so only delete it from
	// the TF state (stop managing via TF)
	d.SetId("")
	return nil
}

func resourceBackupRetentionPolicyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	envId, _ := strconv.Atoi(d.Id())
	_ = d.Set("env_id", envId)
	if err := diagnosticsToError(resourceBackupRetentionPolicyRead(ctx, d, m)); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
