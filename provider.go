package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/reggregory/go-deploy/aptible"
)

func Provider() *schema.Provider {
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
