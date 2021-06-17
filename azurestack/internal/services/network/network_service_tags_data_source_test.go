package network_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
)

type NetworkServiceTagsDataSource struct {
}

func TestAccDataSourceAzureRMServiceTags_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_network_service_tags", "test")
	r := NetworkServiceTagsDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("address_prefixes.#").Exists(),
			),
		},
	})
}

func TestAccDataSourceAzureRMServiceTags_region(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_network_service_tags", "test")
	r := NetworkServiceTagsDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.region(),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("address_prefixes.#").Exists(),
			),
		},
	})
}

func (NetworkServiceTagsDataSource) basic() string {
	return `data "azurerm_network_service_tags" "test" {
  location = "westcentralus"
  service  = "AzureKeyVault"
}`
}

func (NetworkServiceTagsDataSource) region() string {
	return `data "azurerm_network_service_tags" "test" {
  location        = "westcentralus"
  service         = "AzureKeyVault"
  location_filter = "australiacentral"
}`
}
