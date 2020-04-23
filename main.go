package main

import (
	"github.com/aptible/terrform-provider-aptible/aptible"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aptible.Provider})
}
