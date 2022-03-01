package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type NetworkInterfaceResource struct{}

func TestAccNetworkInterface_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
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

func TestAccNetworkInterface_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccNetworkInterface_dnsServers(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.dnsServers(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.dnsServersUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_enableIPForwarding(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// Enabled
			Config: r.enableIPForwarding(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			// Disabled
			Config: r.enableIPForwarding(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			// Enabled
			Config: r.enableIPForwarding(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_multipleIPConfigurations(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multipleIPConfigurations(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_multipleIPConfigurationsSecondaryAsPrimary(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multipleIPConfigurationsSecondaryAsPrimary(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_publicIP(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.publicIP(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.publicIPRemoved(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.publicIP(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_network_interface"),
		},
	})
}

func TestAccNetworkInterface_static(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_tags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.tags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.tagsUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.multipleIPConfigurations(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetworkInterface_updateMultipleParameters(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_network_interface", "test")
	r := NetworkInterfaceResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withMultipleParameters(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateMultipleParameters(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t NetworkInterfaceResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.NetworkInterfaceID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.InterfacesClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", *id, err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (NetworkInterfaceResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.NetworkInterfaceID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.InterfacesClient.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.InterfacesClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return pointer.FromBool(true), nil
}

func (r NetworkInterfaceResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) withMultipleParameters(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                 = "acctestni-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  enable_ip_forwarding = true

  dns_servers = [
    "10.0.0.5",
    "10.0.0.6"
  ]

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }

  tags = {
    env = "Test"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) updateMultipleParameters(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                 = "acctestni-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  enable_ip_forwarding = true

  dns_servers = [
    "10.0.0.5",
    "10.0.0.7"
  ]

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }

  tags = {
    env = "Test2"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) dnsServers(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  dns_servers = [
    "10.0.0.5",
    "10.0.0.6"
  ]

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) dnsServersUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  dns_servers = [
    "10.0.0.6",
    "10.0.0.5"
  ]

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) enableIPForwarding(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                 = "acctestni-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  enable_ip_forwarding = %t

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.template(data), data.RandomInteger, enabled)
}

func (r NetworkInterfaceResource) multipleIPConfigurations(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    primary                       = true
  }

  ip_configuration {
    name                          = "secondary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) multipleIPConfigurationsSecondaryAsPrimary(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }

  ip_configuration {
    name                          = "secondary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    primary                       = true
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) publicIP(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurestack_public_ip.test.id
  }
}
`, r.publicIPTemplate(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) publicIPRemoved(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.publicIPTemplate(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) publicIPTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "import" {
  name                = azurestack_network_interface.test.name
  location            = azurestack_network_interface.test.location
  resource_group_name = azurestack_network_interface.test.resource_group_name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}
`, r.basic(data))
}

func (r NetworkInterfaceResource) static(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Static"
    private_ip_address            = "10.0.2.15"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) tags(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }

  tags = {
    Hello = "World"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r NetworkInterfaceResource) tagsUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }

  tags = {
    Hello     = "World"
    Elephants = "Five"
  }
}
`, r.template(data), data.RandomInteger)
}

func (NetworkInterfaceResource) template(data acceptance.TestData) string {
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
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
