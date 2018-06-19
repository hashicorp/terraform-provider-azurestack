package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: azurestack.Provider})
}
