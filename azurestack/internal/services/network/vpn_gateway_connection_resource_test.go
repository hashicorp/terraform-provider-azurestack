package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type VPNGatewayConnectionResource struct {
}

func TestAccVpnGatewayConnection_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vpn_gateway_connection", "test")
	r := VPNGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVpnGatewayConnection_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vpn_gateway_connection", "test")
	r := VPNGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVpnGatewayConnection_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vpn_gateway_connection", "test")
	r := VPNGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVpnGatewayConnection_customRouteTable(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vpn_gateway_connection", "test")
	r := VPNGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.customRouteTable(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.customRouteTableUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVpnGatewayConnection_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_vpn_gateway_connection", "test")
	r := VPNGatewayConnectionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (t VPNGatewayConnectionResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VpnConnectionID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.VpnConnectionsClient.Get(ctx, id.ResourceGroup, id.VpnGatewayName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading VPN Gateway Connnection (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r VPNGatewayConnectionResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vpn_gateway_connection" "test" {
  name               = "acctest-VpnGwConn-%[2]d"
  vpn_gateway_id     = azurerm_vpn_gateway.test.id
  remote_vpn_site_id = azurerm_vpn_site.test.id
  vpn_link {
    name             = "link1"
    vpn_site_link_id = azurerm_vpn_site.test.link[0].id
  }
  vpn_link {
    name             = "link2"
    vpn_site_link_id = azurerm_vpn_site.test.link[1].id
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VPNGatewayConnectionResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vpn_gateway_connection" "test" {
  name               = "acctest-VpnGwConn-%[2]d"
  vpn_gateway_id     = azurerm_vpn_gateway.test.id
  remote_vpn_site_id = azurerm_vpn_site.test.id
  vpn_link {
    name             = "link1"
    vpn_site_link_id = azurerm_vpn_site.test.link[0].id
  }
  vpn_link {
    name             = "link2"
    vpn_site_link_id = azurerm_vpn_site.test.link[1].id
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VPNGatewayConnectionResource) customRouteTable(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_hub_route_table" "test" {
  name           = "acctest-RouteTable-%[2]d"
  virtual_hub_id = azurerm_virtual_hub.test.id
}

resource "azurerm_vpn_gateway_connection" "test" {
  name               = "acctest-VpnGwConn-%[2]d"
  vpn_gateway_id     = azurerm_vpn_gateway.test.id
  remote_vpn_site_id = azurerm_vpn_site.test.id
  routing {
    associated_route_table  = azurerm_virtual_hub_route_table.test.id
    propagated_route_tables = [azurerm_virtual_hub_route_table.test.id]
  }
  vpn_link {
    name             = "link1"
    vpn_site_link_id = azurerm_vpn_site.test.link[0].id
    ipsec_policy {
      sa_lifetime_sec          = 300
      sa_data_size_kb          = 1024
      encryption_algorithm     = "AES256"
      integrity_algorithm      = "SHA256"
      ike_encryption_algorithm = "AES128"
      ike_integrity_algorithm  = "SHA256"
      dh_group                 = "DHGroup14"
      pfs_group                = "PFS14"
    }
    bandwidth_mbps                        = 30
    protocol                              = "IKEv2"
    ratelimit_enabled                     = true
    route_weight                          = 2
    shared_key                            = "secret"
    local_azure_ip_address_enabled        = true
    policy_based_traffic_selector_enabled = true
  }

  vpn_link {
    name             = "link3"
    vpn_site_link_id = azurerm_vpn_site.test.link[1].id
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VPNGatewayConnectionResource) customRouteTableUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_hub_route_table" "test" {
  name           = "acctest-RouteTable-%[2]d"
  virtual_hub_id = azurerm_virtual_hub.test.id
}

resource "azurerm_virtual_hub_route_table" "test2" {
  name           = "acctest-RouteTable-%[2]d-2"
  virtual_hub_id = azurerm_virtual_hub.test.id
}

resource "azurerm_vpn_gateway_connection" "test" {
  name               = "acctest-VpnGwConn-%[2]d"
  vpn_gateway_id     = azurerm_vpn_gateway.test.id
  remote_vpn_site_id = azurerm_vpn_site.test.id
  routing {
    associated_route_table  = azurerm_virtual_hub_route_table.test2.id
    propagated_route_tables = [azurerm_virtual_hub_route_table.test2.id]
  }
  vpn_link {
    name             = "link1"
    vpn_site_link_id = azurerm_vpn_site.test.link[0].id
    ipsec_policy {
      sa_lifetime_sec          = 300
      sa_data_size_kb          = 1024
      encryption_algorithm     = "AES256"
      integrity_algorithm      = "SHA256"
      ike_encryption_algorithm = "AES128"
      ike_integrity_algorithm  = "SHA256"
      dh_group                 = "DHGroup14"
      pfs_group                = "PFS14"
    }
    bandwidth_mbps                        = 30
    protocol                              = "IKEv2"
    ratelimit_enabled                     = true
    route_weight                          = 2
    shared_key                            = "secret"
    local_azure_ip_address_enabled        = true
    policy_based_traffic_selector_enabled = true
  }

  vpn_link {
    name             = "link3"
    vpn_site_link_id = azurerm_vpn_site.test.link[1].id
  }
}
`, r.template(data), data.RandomInteger)
}

func (r VPNGatewayConnectionResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_vpn_gateway_connection" "import" {
  name               = azurerm_vpn_gateway_connection.test.name
  vpn_gateway_id     = azurerm_vpn_gateway_connection.test.vpn_gateway_id
  remote_vpn_site_id = azurerm_vpn_gateway_connection.test.remote_vpn_site_id
  dynamic "vpn_link" {
    for_each = azurerm_vpn_gateway_connection.test.vpn_link
    iterator = v
    content {
      name             = v.value["name"]
      vpn_site_link_id = v.value["vpn_site_link_id"]
    }
  }
}
`, r.basic(data))
}

func (VPNGatewayConnectionResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-vpn-%[1]d"
  location = "%[2]s"
}


resource "azurerm_virtual_wan" "test" {
  name                = "acctest-vwan-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_virtual_hub" "test" {
  name                = "acctest-vhub-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  virtual_wan_id      = azurerm_virtual_wan.test.id
  address_prefix      = "10.0.0.0/24"
}

resource "azurerm_vpn_gateway" "test" {
  name                = "acctest-vpngw-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  virtual_hub_id      = azurerm_virtual_hub.test.id
}

resource "azurerm_vpn_site" "test" {
  name                = "acctest-vpnsite-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  virtual_wan_id      = azurerm_virtual_wan.test.id
  link {
    name       = "link1"
    ip_address = "10.0.0.1"
  }
  link {
    name       = "link2"
    ip_address = "10.0.0.2"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}
