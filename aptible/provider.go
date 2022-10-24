package aptible

import (
	"context"
	"log"

	"github.com/aptible/go-deploy/aptible"
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
			"aptible_environment": dataSourceEnvironment(),
			"aptible_stack":       dataSourceStack(),
		},
		ConfigureContextFunc: providerConfigureWithContext,
	}
}

func providerConfigureWithContext(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	client, err := aptible.SetUpClient()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when initializing the provider.",
			Detail:   "There was an error when initializing the provider.",
		})
		log.Println("[ERR] Error in attempting to start the provider", err)
		return nil, diags
	}
	return client, nil
}
