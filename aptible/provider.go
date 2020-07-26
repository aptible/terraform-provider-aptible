package aptible

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
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
	client, err := aptible.SetUpClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}
