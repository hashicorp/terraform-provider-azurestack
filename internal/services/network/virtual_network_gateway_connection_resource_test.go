package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
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

func TestAccVirtualNetworkGatewayConnection_vnettonet(t *testing.T) {
	data1 := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test_1")
	data2 := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test_2")
	r := VirtualNetworkGatewayConnectionResource{}

	sharedKey := "4-v3ry-53cr37-1p53c-5h4r3d-k3y"

	data1.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vnettovnet(data1, data2.RandomInteger, sharedKey),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data1.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttr(data1.ResourceName, "shared_key", sharedKey),
				acceptance.TestCheckResourceAttr(data2.ResourceName, "shared_key", sharedKey),
			),
		},
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

func TestAccVirtualNetworkGatewayConnection_updatingSharedKey(t *testing.T) {
	data1 := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test_1")
	data2 := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway_connection", "test_2")
	r := VirtualNetworkGatewayConnectionResource{}

	firstSharedKey := "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
	secondSharedKey := "4-r33ly-53cr37-1p53c-5h4r3d-k3y"

	data1.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vnettovnet(data1, data2.RandomInteger, firstSharedKey),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data1.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttr(data1.ResourceName, "shared_key", firstSharedKey),
				acceptance.TestCheckResourceAttr(data2.ResourceName, "shared_key", firstSharedKey),
			),
		},
		{
			Config: r.vnettovnet(data1, data2.RandomInteger, secondSharedKey),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data1.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttr(data1.ResourceName, "shared_key", secondSharedKey),
				acceptance.TestCheckResourceAttr(data2.ResourceName, "shared_key", secondSharedKey),
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

	return utils.Bool(resp.ID != nil), nil
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

  gateway_address = "168.62.225.23"
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
`, data.RandomInteger, data.Locations.Primary)
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

  gateway_address = "168.62.225.23"
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
`, data.RandomInteger, data.Locations.Primary)
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

func (VirtualNetworkGatewayConnectionResource) vnettovnet(data acceptance.TestData, rInt2 int, sharedKey string) string {
	return fmt.Sprintf(`
variable "random1" {
  default = "%d"
}

variable "random2" {
  default = "%d"
}

variable "shared_key" {
  default = "%s"
}

resource "azurestack_resource_group" "test_1" {
  name     = "acctestRG-${var.random1}"
  location = "%s"
}

resource "azurestack_virtual_network" "test_1" {
  name                = "acctestvn-${var.random1}"
  location            = azurestack_resource_group.test_1.location
  resource_group_name = azurestack_resource_group.test_1.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test_1" {
  name                 = "GatewaySubnet"
  resource_group_name  = azurestack_resource_group.test_1.name
  virtual_network_name = azurestack_virtual_network.test_1.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test_1" {
  name                = "acctest-${var.random1}"
  location            = azurestack_resource_group.test_1.location
  resource_group_name = azurestack_resource_group.test_1.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test_1" {
  name                = "acctest-${var.random1}"
  location            = azurestack_resource_group.test_1.location
  resource_group_name = azurestack_resource_group.test_1.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = azurestack_public_ip.test_1.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test_1.id
  }
}

resource "azurestack_virtual_network_gateway_connection" "test_1" {
  name                = "acctest-${var.random1}"
  location            = azurestack_resource_group.test_1.location
  resource_group_name = azurestack_resource_group.test_1.name

  type                            = "Vnet2Vnet"
  virtual_network_gateway_id      = azurestack_virtual_network_gateway.test_1.id
  peer_virtual_network_gateway_id = azurestack_virtual_network_gateway.test_2.id

  shared_key = var.shared_key
}

resource "azurestack_resource_group" "test_2" {
  name     = "acctestRG-${var.random2}"
  location = "%s"
}

resource "azurestack_virtual_network" "test_2" {
  name                = "acctest-${var.random2}"
  location            = azurestack_resource_group.test_2.location
  resource_group_name = azurestack_resource_group.test_2.name
  address_space       = ["10.1.0.0/16"]
}

resource "azurestack_subnet" "test_2" {
  name                 = "GatewaySubnet"
  resource_group_name  = azurestack_resource_group.test_2.name
  virtual_network_name = azurestack_virtual_network.test_2.name
  address_prefix       = "10.1.1.0/24"
}

resource "azurestack_public_ip" "test_2" {
  name                = "acctest-${var.random2}"
  location            = azurestack_resource_group.test_2.location
  resource_group_name = azurestack_resource_group.test_2.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test_2" {
  name                = "acctest-${var.random2}"
  location            = azurestack_resource_group.test_2.location
  resource_group_name = azurestack_resource_group.test_2.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = azurestack_public_ip.test_2.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test_2.id
  }
}

resource "azurestack_virtual_network_gateway_connection" "test_2" {
  name                = "acctest-${var.random2}"
  location            = azurestack_resource_group.test_2.location
  resource_group_name = azurestack_resource_group.test_2.name

  type                            = "Vnet2Vnet"
  virtual_network_gateway_id      = azurestack_virtual_network_gateway.test_2.id
  peer_virtual_network_gateway_id = azurestack_virtual_network_gateway.test_1.id

  shared_key = var.shared_key
}
`, data.RandomInteger, rInt2, sharedKey, data.Locations.Primary, data.Locations.Secondary)
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
  sku      = "VpnGw1"

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

  gateway_address = "168.62.225.23"
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
    ike_encryption   = "GCMAES256"
    ike_integrity    = "SHA256"
    ipsec_encryption = "AES256"
    ipsec_integrity  = "SHA256"
    pfs_group        = "PFS14"
    sa_datasize      = 102400000
    sa_lifetime      = 27000
  }

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
`, data.RandomInteger, data.Locations.Primary)
}
