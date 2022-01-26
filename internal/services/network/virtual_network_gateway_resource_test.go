package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type VirtualNetworkGatewayResource struct{}

func TestAccVirtualNetworkGateway_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku").HasValue("Basic"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetworkGateway_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_virtual_network_gateway"),
		},
	})
}

func TestAccVirtualNetworkGateway_lowerCaseSubnetName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.lowerCaseSubnetName(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku").HasValue("Basic"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetworkGateway_vpnGw1(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vpnGw1(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_activeActive(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.activeActive(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_standard(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sku(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku").HasValue("Standard"),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_vpnGw2(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sku(data, "VpnGw2"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku").HasValue("VpnGw2"),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_vpnGw3(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.sku(data, "VpnGw3"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku").HasValue("VpnGw3"),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_vpnClientConfig(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vpnClientConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("vpn_client_configuration.0.radius_server_address").HasValue("1.2.3.4"),
				check.That(data.ResourceName).Key("vpn_client_configuration.0.vpn_client_protocols.#").HasValue("2"),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_vpnClientConfigOpenVPN(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vpnClientConfigOpenVPN(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("vpn_client_configuration.0.vpn_client_protocols.#").HasValue("1"),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_expressRoute(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.expressRoute(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("type").HasValue("ExpressRoute"),
				check.That(data.ResourceName).Key("bgp_settings.#").HasValue("0"),
			),
		},
	})
}

func TestAccVirtualNetworkGateway_privateIpAddressEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_gateway", "test")
	r := VirtualNetworkGatewayResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.privateIpAddressEnabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.privateIpAddressEnabledUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t VirtualNetworkGatewayResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	gatewayName := state.Attributes["name"]
	resourceGroup := state.Attributes["resource_group_name"]

	resp, err := clients.Network.VnetGatewayClient.Get(ctx, resourceGroup, gatewayName)
	if err != nil {
		return nil, fmt.Errorf("reading Virtual Network Gateway (%s): %+v", state.ID, err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (VirtualNetworkGatewayResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctestpip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r VirtualNetworkGatewayResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_virtual_network_gateway" "import" {
  name                = azurestack_virtual_network_gateway.test.name
  location            = azurestack_virtual_network_gateway.test.location
  resource_group_name = azurestack_virtual_network_gateway.test.resource_group_name
  type                = azurestack_virtual_network_gateway.test.type
  vpn_type            = azurestack_virtual_network_gateway.test.vpn_type
  sku                 = azurestack_virtual_network_gateway.test.sku

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, r.basic(data))
}

func (VirtualNetworkGatewayResource) lowerCaseSubnetName(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test" {
  name                 = "gatewaySubnet"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) vpnGw1(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctestpip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "VpnGw1"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) activeActive(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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

resource "azurestack_public_ip" "first" {
  name                = "acctestpip1-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_public_ip" "second" {
  name = "acctestpip2-%d"

  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  depends_on = [
    azurestack_public_ip.first,
    azurestack_public_ip.second,
  ]
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "VpnGw1"

  active_active = true
  enable_bgp    = true

  ip_configuration {
    name                 = "gw-ip1"
    public_ip_address_id = azurestack_public_ip.first.id

    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }

  ip_configuration {
    name                          = "gw-ip2"
    public_ip_address_id          = azurestack_public_ip.second.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }

  bgp_settings {
    asn = "65010"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) vpnClientConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctestpip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  depends_on          = [azurestack_public_ip.test]
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "VpnGw1"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }

  vpn_client_configuration {
    address_space        = ["10.2.0.0/24"]
    vpn_client_protocols = ["SSTP", "IkeV2"]

    radius_server_address = "1.2.3.4"
    radius_server_secret  = "1234"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) vpnClientConfigOpenVPN(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctestpip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  depends_on          = [azurestack_public_ip.test]
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "VpnGw1"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }

  vpn_client_configuration {
    address_space        = ["10.2.0.0/24"]
    vpn_client_protocols = ["OpenVPN"]
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) sku(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctestpip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "%s"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, sku)
}

func (VirtualNetworkGatewayResource) expressRoute(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctestpip1-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctestvng-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type     = "ExpressRoute"
  vpn_type = "PolicyBased"
  sku      = "Standard"

  ip_configuration {
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) privateIpAddressEnabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctest-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctest-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type                       = "Vpn"
  vpn_type                   = "RouteBased"
  sku                        = "VpnGw1AZ"
  private_ip_address_enabled = true

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualNetworkGatewayResource) privateIpAddressEnabledUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
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
  name                = "acctest-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctest-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  type                       = "Vpn"
  vpn_type                   = "RouteBased"
  sku                        = "VpnGw1AZ"
  private_ip_address_enabled = false

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = azurestack_public_ip.test.id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
