package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceArmVirtualNetwork_basic(t *testing.T) {
	dataSourceName := "data.azurestack_virtual_network.test"
	ri := acctest.RandInt()

	name := fmt.Sprintf("acctestvnet-%d", ri)
	config := testAccDataSourceArmVirtualNetwork_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "dns_servers.0", "10.0.0.4"),
					resource.TestCheckResourceAttr(dataSourceName, "address_spaces.0", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(dataSourceName, "subnets.0", "subnet1"),
				),
			},
		},
	})
}

// Peering not in scope
func TestAccDataSourceArmVirtualNetwork_peering(t *testing.T) {

	t.Skip()

	dataSourceName := "data.azurestack_virtual_network.test"
	ri := acctest.RandInt()

	virtualNetworkName := fmt.Sprintf("acctestvnet-1-%d", ri)
	location := testLocation()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArmVirtualNetwork_peering(ri, location),
			},
			{
				Config: testAccDataSourceArmVirtualNetwork_peeringWithDataSource(ri, location),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", virtualNetworkName),
					resource.TestCheckResourceAttr(dataSourceName, "address_spaces.0", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(dataSourceName, "vnet_peerings.%", "1"),
				),
			},
		},
	})
}

func testAccDataSourceArmVirtualNetwork_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvnet-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  dns_servers         = ["10.0.0.4"]

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }
}

data "azurestack_virtual_network" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  name                = "${azurestack_virtual_network.test.name}"
}
`, rInt, location, rInt)
}

func testAccDataSourceArmVirtualNetwork_peering(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_virtual_network" "test1" {
  name                = "acctestvnet-1-%d"
  address_space       = ["10.0.1.0/24"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_virtual_network" "test2" {
  name                = "acctestvnet-2-%d"
  address_space       = ["10.0.2.0/24"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_virtual_network_peering" "test1" {
  name                      = "peer-1to2"
  resource_group_name       = "${azurestack_resource_group.test.name}"
  virtual_network_name      = "${azurestack_virtual_network.test1.name}"
  remote_virtual_network_id = "${azurestack_virtual_network.test2.id}"
}
`, rInt, location, rInt, rInt)
}

func testAccDataSourceArmVirtualNetwork_peeringWithDataSource(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_virtual_network" "test1" {
  name                = "acctestvnet-1-%d"
  address_space       = ["10.0.1.0/24"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_virtual_network" "test2" {
  name                = "acctestvnet-2-%d"
  address_space       = ["10.0.2.0/24"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_virtual_network_peering" "test1" {
  name                      = "peer-1to2"
  resource_group_name       = "${azurestack_resource_group.test.name}"
  virtual_network_name      = "${azurestack_virtual_network.test1.name}"
  remote_virtual_network_id = "${azurestack_virtual_network.test2.id}"
}

data "azurestack_virtual_network" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  name                = "${azurestack_virtual_network.test1.name}"
}
`, rInt, location, rInt, rInt)
}
