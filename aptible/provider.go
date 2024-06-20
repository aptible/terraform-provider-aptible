package aptible

import (
	"context"
	"log"
	"os"

	"github.com/aptible/aptible-api-go/aptibleapi"
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

	token, err := aptible.GetToken()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "There was an error when initializing the provider.",
			Detail:   "There was an error when initializing the provider.",
		})
		log.Println("[ERR] Error in attempting to start the provider", err)
		return nil, diags
	}

	return &providerMeta{
		LegacyClient: client,
		Client:       aptibleapi.NewAPIClient(aptibleapi.NewAPIConfiguration()),
		Token:        token,
	}, nil
}

type providerMeta struct {
	LegacyClient *aptible.Client
	Client       *aptibleapi.APIClient
	Token        string
}

// Configures the provided context to work with aptibleapi.APIClient requests
func (m *providerMeta) APIContext(ctx context.Context) context.Context {
	// Override the default API url with APTIBLE_API_ROOT_URL, if non-empty
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
