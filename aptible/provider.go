package aptible

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("APTIBLE_ACCESS_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"aptible_app":      resourceApp(),
			"aptible_endpoint": resourceEndpoint(),
			"aptible_database": resourceDatabase(),
			"aptible_replica":  resourceReplica(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"aptible_environment": dataSourceEnvironment(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	attrs := aptible.ClientAttrs{
		TokenString: d.Get("token").(string),
	}
	client, err := aptible.SetUpClient(attrs)
	if err != nil {
		return nil, err
	}
	return client, nil
}
