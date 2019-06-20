package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceArmSubnet_basic(t *testing.T) {
	resourceName := "data.azurestack_subnet.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArmSubnet_basic(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_group_name"),
					resource.TestCheckResourceAttrSet(resourceName, "virtual_network_name"),
					resource.TestCheckResourceAttrSet(resourceName, "address_prefix"),
					resource.TestCheckResourceAttr(resourceName, "network_security_group_id", ""),
					resource.TestCheckResourceAttr(resourceName, "route_table_id", ""),
				),
			},
		},
	})
}

func TestAccDataSourceArmSubnet_networkSecurityGroup(t *testing.T) {
	dataSourceName := "data.azurestack_subnet.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArmSubnet_networkSecurityGroup(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_group_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "virtual_network_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "address_prefix"),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_security_group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "route_table_id", ""),
				),
			},
		},
	})
}

func TestAccDataSourceArmSubnet_routeTable(t *testing.T) {
	dataSourceName := "data.azurestack_subnet.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArmSubnet_routeTable(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_group_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "virtual_network_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "address_prefix"),
					resource.TestCheckResourceAttr(dataSourceName, "network_security_group_id", ""),
					resource.TestCheckResourceAttrSet(dataSourceName, "route_table_id"),
				),
			},
		},
	})
}

func testAccDataSourceArmSubnet_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctest%d-private"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.0.0/24"
}

data "azurestack_subnet" "test" {
  name                 = "${azurestack_subnet.test.name}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
}
`, rInt, location, rInt, rInt)
}

func testAccDataSourceArmSubnet_networkSecurityGroup(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                      = "acctest%d-private"
  resource_group_name       = "${azurestack_resource_group.test.name}"
  virtual_network_name      = "${azurestack_virtual_network.test.name}"
  address_prefix            = "10.0.0.0/24"
  network_security_group_id = "${azurestack_network_security_group.test.id}"
}

data "azurestack_subnet" "test" {
  name                 = "${azurestack_subnet.test.name}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccDataSourceArmSubnet_routeTable(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctest-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name                   = "acctest-%d"
    address_prefix         = "10.100.0.0/14"
    next_hop_type          = "VirtualAppliance"
    next_hop_in_ip_address = "10.10.1.1"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
  route_table_id       = "${azurestack_route_table.test.id}"
}

data "azurestack_subnet" "test" {
  name                 = "${azurestack_subnet.test.name}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
}
`, rInt, location, rInt, rInt, rInt, rInt)
}
