package main

import (
	"github.com/aptible/go-deploy/aptible"
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"aptible_app":      resourceApp(),
			"aptible_endpoint": resourceEndpoint(),
			"aptible_db":       resourceDatabase(),
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
