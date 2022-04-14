package network_test

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type VirtualNetworkGatewayConnectionDataSource struct{}

func TestAccDataSourceVirtualNetworkGatewayConnection_sitetosite(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_virtual_network_gateway_connection", "test")
	r := VirtualNetworkGatewayConnectionDataSource{}
	sharedKey := "4-v3ry-53cr37-1p53c-5h4r3d-k3y"

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.sitetosite(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("shared_key").HasValue(sharedKey),
				check.That(data.ResourceName).Key("type").HasValue(string(network.IPsec)),
			),
		},
	})
}

func TestAccDataSourceVirtualNetworkGatewayConnection_ipsecpolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_virtual_network_gateway_connection", "test")
	r := VirtualNetworkGatewayConnectionDataSource{}
	sharedKey := "4-v3ry-53cr37-1p53c-5h4r3d-k3y"

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.ipsecpolicy(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("shared_key").HasValue(sharedKey),
				check.That(data.ResourceName).Key("type").HasValue(string(network.IPsec)),
				check.That(data.ResourceName).Key("routing_weight").HasValue("20"),
				check.That(data.ResourceName).Key("ipsec_policy.0.dh_group").HasValue(string(network.DHGroup14)),
				check.That(data.ResourceName).Key("ipsec_policy.0.ike_encryption").HasValue(string(network.AES256)),
				check.That(data.ResourceName).Key("ipsec_policy.0.ike_integrity").HasValue(string(network.IkeIntegritySHA256)),
				check.That(data.ResourceName).Key("ipsec_policy.0.ipsec_encryption").HasValue(string(network.IpsecEncryptionAES256)),
				check.That(data.ResourceName).Key("ipsec_policy.0.ipsec_integrity").HasValue(string(network.IpsecIntegritySHA256)),
				check.That(data.ResourceName).Key("ipsec_policy.0.pfs_group").HasValue(string(network.PfsGroupPFS2048)),
				check.That(data.ResourceName).Key("ipsec_policy.0.sa_datasize").HasValue("102400000"),
				check.That(data.ResourceName).Key("ipsec_policy.0.sa_lifetime").HasValue("27000"),
			),
		},
	})
}

func (VirtualNetworkGatewayConnectionDataSource) sitetosite(data acceptance.TestData) string {
	return fmt.Sprintf(`
variable "random" {
  default = "%d"
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-${var.random}"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test" {
  name                 = "GatewaySubnet"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  type                = "Vpn"
  vpn_type            = "RouteBased"
  sku                 = "Basic"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}

resource "azurestack_local_network_gateway" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  gateway_address     = "168.62.225.%d"
  address_space       = ["10.1.1.0/24"]
}

resource "azurestack_virtual_network_gateway_connection" "test" {
  name                       = "acctest-${var.random}"
  location                   = azurestack_resource_group.test.location
  resource_group_name        = azurestack_resource_group.test.name
  type                       = "IPsec"
  virtual_network_gateway_id = azurestack_virtual_network_gateway.test.id
  local_network_gateway_id   = azurestack_local_network_gateway.test.id
  shared_key                 = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}

data "azurestack_virtual_network_gateway_connection" "test" {
  name                = azurestack_virtual_network_gateway_connection.test.name
  resource_group_name = azurestack_virtual_network_gateway_connection.test.resource_group_name
}
`, data.RandomInteger, data.Locations.Primary, acctest.RandIntRange(2, 253))
}

func (VirtualNetworkGatewayConnectionDataSource) ipsecpolicy(data acceptance.TestData) string {
	return fmt.Sprintf(`
variable "random" {
  default = "%d"
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-${var.random}"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test" {
  name                 = "GatewaySubnet"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  type                = "Vpn"
  vpn_type            = "RouteBased"
  sku                 = "Standard"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}

resource "azurestack_local_network_gateway" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  gateway_address     = "168.62.225.%d"
  address_space       = ["10.1.1.0/24"]
}

resource "azurestack_virtual_network_gateway_connection" "test" {
  name                               = "acctest-${var.random}"
  location                           = azurestack_resource_group.test.location
  resource_group_name                = azurestack_resource_group.test.name
  type                               = "IPsec"
  virtual_network_gateway_id         = azurestack_virtual_network_gateway.test.id
  local_network_gateway_id           = azurestack_local_network_gateway.test.id
  use_policy_based_traffic_selectors = true
  routing_weight                     = 20

  ipsec_policy {
    dh_group         = "DHGroup14"
    ike_encryption   = "AES256"
    ike_integrity    = "SHA256"
    ipsec_encryption = "AES256"
    ipsec_integrity  = "SHA256"
    pfs_group        = "PFS2048"
    sa_datasize      = 102400000
    sa_lifetime      = 27000
  }

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}

data "azurestack_virtual_network_gateway_connection" "test" {
  name                = azurestack_virtual_network_gateway_connection.test.name
  resource_group_name = azurestack_virtual_network_gateway_connection.test.resource_group_name
}
`, data.RandomInteger, data.Locations.Primary, acctest.RandIntRange(2, 253))
}
