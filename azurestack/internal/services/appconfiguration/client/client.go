package client

import (
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/common"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/appconfiguration/sdk/configurationstores"
)

type Client struct {
	ConfigurationStoresClient *configurationstores.ConfigurationStoresClient
}

func NewClient(o *common.ClientOptions) *Client {
	configurationStores := configurationstores.NewConfigurationStoresClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&configurationStores.Client, o.ResourceManagerAuthorizer)

	return &Client{
		ConfigurationStoresClient: &configurationStores,
	}
}
