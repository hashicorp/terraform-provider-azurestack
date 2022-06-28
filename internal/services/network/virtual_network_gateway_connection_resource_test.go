package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type VirtualNetworkGatewayConnectionResource struct{}

func TestAccVirtualNetworkGatewayConnection_sitetosite(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test")
	r := VirtualNetworkGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sitetosite(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetworkGatewayConnection_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test")
	r := VirtualNetworkGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sitetosite(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_virtual_network_gateway_connection"),
		},
	})
}

func TestAccVirtualNetworkGatewayConnection_sitetositeWithoutSharedKey(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test")
	r := VirtualNetworkGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sitetositeWithoutSharedKey(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetworkGatewayConnection_ipsecpolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test")
	r := VirtualNetworkGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.ipsecpolicy(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func (t VirtualNetworkGatewayConnectionResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	gatewayName := state.Attributes["name"]
	resourceGroup := state.Attributes["resource_group_name"]

	resp, err := clients.Network.VnetGatewayConnectionsClient.Get(ctx, resourceGroup, gatewayName)
	if err != nil {
		return nil, fmt.Errorf("reading Virtual Network Gateway Connection (%s): %+v", state.ID, err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (VirtualNetworkGatewayConnectionResource) sitetosite(data acceptance.TestData) string {
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

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

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

  gateway_address = "168.62.225.%d"
  address_space   = ["10.1.1.0/24"]
}

resource "azurestack_virtual_network_gateway_connection" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type                       = "IPsec"
  virtual_network_gateway_id = azurestack_virtual_network_gateway.test.id
  local_network_gateway_id   = azurestack_local_network_gateway.test.id

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
`, data.RandomInteger, data.Locations.Primary, acctest.RandIntRange(2, 253))
}

func (VirtualNetworkGatewayConnectionResource) sitetositeWithoutSharedKey(data acceptance.TestData) string {
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

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

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

  gateway_address = "168.62.225.%d"
  address_space   = ["10.1.1.0/24"]
}

resource "azurestack_virtual_network_gateway_connection" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type                       = "IPsec"
  virtual_network_gateway_id = azurestack_virtual_network_gateway.test.id
  local_network_gateway_id   = azurestack_local_network_gateway.test.id
}
`, data.RandomInteger, data.Locations.Primary, acctest.RandIntRange(2, 253))
}

func (r VirtualNetworkGatewayConnectionResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_virtual_network_gateway_connection" "import" {
  name                       = azurestack_virtual_network_gateway_connection.test.name
  location                   = azurestack_virtual_network_gateway_connection.test.location
  resource_group_name        = azurestack_virtual_network_gateway_connection.test.resource_group_name
  type                       = azurestack_virtual_network_gateway_connection.test.type
  virtual_network_gateway_id = azurestack_virtual_network_gateway_connection.test.virtual_network_gateway_id
  local_network_gateway_id   = azurestack_virtual_network_gateway_connection.test.local_network_gateway_id
  shared_key                 = azurestack_virtual_network_gateway_connection.test.shared_key
}
`, r.sitetosite(data))
}

func (VirtualNetworkGatewayConnectionResource) ipsecpolicy(data acceptance.TestData) string {
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

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Standard"

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

  gateway_address = "168.62.225.%d"
  address_space   = ["10.1.1.0/24"]
}

resource "azurestack_virtual_network_gateway_connection" "test" {
  name                = "acctest-${var.random}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type                       = "IPsec"
  virtual_network_gateway_id = azurestack_virtual_network_gateway.test.id
  local_network_gateway_id   = azurestack_local_network_gateway.test.id

  use_policy_based_traffic_selectors = true
  routing_weight                     = 20

  ipsec_policy {
    dh_group         = "DHGroup14"
    ike_encryption   = "AES256"
    ike_integrity    = "SHA256"
    ipsec_encryption = "AES256"
    ipsec_integrity  = "SHA256"
    pfs_group        = "PFS24"
    sa_datasize      = 102400000
    sa_lifetime      = 27000
  }

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
`, data.RandomInteger, data.Locations.Primary, acctest.RandIntRange(2, 253))
}
