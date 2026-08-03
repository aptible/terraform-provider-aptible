package aptible

import (
	"context"

	"github.com/aptible/aptible-api-go/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"aptible_app":          resourceApp(),
			"aptible_database":     resourceDatabase(),
			"aptible_environment":  resourceEnvironment(),
			"aptible_endpoint":     resourceEndpoint(),
			"aptible_replica":      resourceReplica(),
			"aptible_log_drain":    resourceLogDrain(),
			"aptible_metric_drain": resourceMetricDrain(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"aptible_environment":             dataSourceEnvironment(),
			"aptible_backup_retention_policy": dataSourceBackupRetentionPolicy(),
			"aptible_stack":                   dataSourceStack(),
		},
		ConfigureContextFunc: providerConfigureWithContext,
	}
}

func providerConfigureWithContext(_ context.Context, _ *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token, err := helpers.GetToken()
	if err != nil {
		return nil, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "There was an error when initializing the provider.",
			Detail:   err.Error(),
		}}
	}

	return newProviderMetadata(token, helpers.GetAPIRoot()), nil
}
