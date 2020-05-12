package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/aptible/terraform-provider-aptible/aptible"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aptible.Provider})
}
