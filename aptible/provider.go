package aptible

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/reggregory/go-deploy/aptible"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"aptible_app":      resourceApp(),
			"aptible_endpoint": resourceEndpoint(),
			"aptible_db":       resourceDatabase(),
			"aptible_replica":  resourceReplica(),
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
