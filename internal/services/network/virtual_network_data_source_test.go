package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type VirtualNetworkDataSource struct{}

func TestAccVirtualNetworkDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_virtual_network", "test")
	r := VirtualNetworkDataSource{}

	name := fmt.Sprintf("acctestvnet-%d", data.RandomInteger)

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").HasValue(name),
				check.That(data.ResourceName).Key("location").HasValue(location.Normalize(data.Locations.Primary)),
				check.That(data.ResourceName).Key("dns_servers.0").HasValue("10.0.0.4"),
				check.That(data.ResourceName).Key("address_space.0").HasValue("10.0.0.0/16"),
				check.That(data.ResourceName).Key("subnets.0").HasValue("subnet1"),
			),
		},
	})
}

func (VirtualNetworkDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvnet-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  dns_servers         = ["10.0.0.4"]

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }
}

data "azurestack_virtual_network" "test" {
  resource_group_name = azurestack_resource_group.test.name
  name                = azurestack_virtual_network.test.name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
