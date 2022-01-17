package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type SubnetDataSource struct{}

func TestAccDataSourceSubnet_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subnet", "test")
	r := SubnetDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("virtual_network_name").Exists(),
				check.That(data.ResourceName).Key("address_prefix").Exists(),
				check.That(data.ResourceName).Key("network_security_group_id").HasValue(""),
				check.That(data.ResourceName).Key("route_table_id").HasValue(""),
			),
		},
	})
}

func TestAccDataSourceSubnet_networkSecurityGroup(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subnet", "test")
	r := SubnetDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			// since the network security group association is a separate resource this forces it
			Config: r.networkSecurityGroupDependencies(data),
		},
		{
			Config: r.networkSecurityGroup(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("virtual_network_name").Exists(),
				check.That(data.ResourceName).Key("address_prefix").Exists(),
				check.That(data.ResourceName).Key("network_security_group_id").Exists(),
				check.That(data.ResourceName).Key("route_table_id").HasValue(""),
			),
		},
	})
}

func TestAccDataSourceSubnet_routeTable(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subnet", "test")
	r := SubnetDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			// since the route table association is a separate resource this forces it
			Config: r.routeTableDependencies(data),
		},
		{
			Config: r.routeTable(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("virtual_network_name").Exists(),
				check.That(data.ResourceName).Key("address_prefix").Exists(),
				check.That(data.ResourceName).Key("route_table_id").Exists(),
				check.That(data.ResourceName).Key("network_security_group_id").HasValue(""),
			),
		},
	})
}

func (r SubnetDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.0.0/24"
}

data "azurestack_subnet" "test" {
  name                 = azurestack_subnet.test.name
  virtual_network_name = azurestack_subnet.test.virtual_network_name
  resource_group_name  = azurestack_subnet.test.resource_group_name
}
`, r.template(data))
}

func (r SubnetDataSource) networkSecurityGroupDependencies(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.0.0/24"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

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

resource "azurestack_subnet_network_security_group_association" "test" {
  subnet_id                 = azurestack_subnet.test.id
  network_security_group_id = azurestack_network_security_group.test.id
}
`, r.template(data), data.RandomInteger)
}

func (r SubnetDataSource) networkSecurityGroup(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_subnet" "test" {
  name                 = azurestack_subnet.test.name
  virtual_network_name = azurestack_subnet.test.virtual_network_name
  resource_group_name  = azurestack_subnet.test.resource_group_name
}
`, r.networkSecurityGroupDependencies(data))
}

func (r SubnetDataSource) routeTableDependencies(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.0.0/24"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  route {
    name                   = "first"
    address_prefix         = "10.100.0.0/14"
    next_hop_type          = "VirtualAppliance"
    next_hop_in_ip_address = "10.10.1.1"
  }
}

resource "azurestack_subnet_route_table_association" "test" {
  subnet_id      = azurestack_subnet.test.id
  route_table_id = azurestack_route_table.test.id
}
`, r.template(data), data.RandomInteger)
}

func (r SubnetDataSource) routeTable(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_subnet" "test" {
  name                 = azurestack_subnet.test.name
  virtual_network_name = azurestack_subnet.test.virtual_network_name
  resource_group_name  = azurestack_subnet.test.resource_group_name
}
`, r.routeTableDependencies(data))
}

func (SubnetDataSource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  address_space       = ["10.0.0.0/16"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
