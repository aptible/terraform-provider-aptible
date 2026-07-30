package aptible

import (
	"context"
	"log"
	"os"

	"github.com/aptible/aptible-api-go/aptibleapi"
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

	return &providerMetadata{
		Client: aptibleapi.NewAPIClient(aptibleapi.NewAPIConfiguration()),
		Token:  token,
	}, nil
}

type providerMetadata struct {
	Client *aptibleapi.APIClient
	Token  string
}

func (m *providerMetadata) APIContext(ctx context.Context) context.Context {
	if url := os.Getenv("APTIBLE_API_ROOT_URL"); url != "" {
		ctx = context.WithValue(ctx, aptibleapi.ContextServerVariables, map[string]string{"url": url})
	}

	if m.Token == "" {
		log.Fatalln("Could not read token: Please run aptible login or set APTIBLE_ACCESS_TOKEN")
		return ctx
	}

	return context.WithValue(ctx, aptibleapi.ContextAPIKeys, map[string]aptibleapi.APIKey{
		"token": {
			Prefix: "Bearer",
			Key:    m.Token,
		},
	})
}
