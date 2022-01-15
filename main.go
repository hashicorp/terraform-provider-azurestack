package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-provider-azurestack/azurestack"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: azurestack.Provider,
	})
}
