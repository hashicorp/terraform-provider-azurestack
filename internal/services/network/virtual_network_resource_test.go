package network_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type VirtualNetworkResource struct{}

func TestAccVirtualNetwork_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet.#").HasValue("1"),
				check.That(data.ResourceName).Key("subnet.0.id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetwork_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

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

func TestAccVirtualNetwork_updateFlowTimeoutInMinutes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateFlowTimeoutInMinutes(data, 5),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateFlowTimeoutInMinutes(data, 6),
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

func TestAccVirtualNetwork_basicUpdated(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet.#").HasValue("1"),
				check.That(data.ResourceName).Key("subnet.0.id").Exists(),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet.#").HasValue("2"),
				check.That(data.ResourceName).Key("subnet.0.id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetwork_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_virtual_network"),
		},
	})
}

func TestAccVirtualNetwork_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccVirtualNetwork_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet.#").HasValue("1"),
				check.That(data.ResourceName).Key("subnet.0.id").Exists(),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("Production"),
				check.That(data.ResourceName).Key("tags.cost_center").HasValue("MSFT"),
			),
		},
		{
			Config: r.withTagsUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet.#").HasValue("1"),
				check.That(data.ResourceName).Key("subnet.0.id").Exists(),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("staging"),
			),
		},
	})
}

func TestAccVirtualNetwork_deleteSubnet(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network", "test")
	r := VirtualNetworkResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.noSubnet(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet.#").HasValue("0"),
			),
		},
		data.ImportStep(),
	})
}

func (t VirtualNetworkResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualNetworkID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.VnetClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", *id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r VirtualNetworkResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualNetworkID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.VnetClient.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting on Virtual Network: %+v", err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.VnetClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for deletion of Virtual Network %q: %+v", id, err)
	}

	return utils.Bool(true), nil
}

func (VirtualNetworkResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (VirtualNetworkResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16", "10.10.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  dns_servers         = ["10.7.7.2", "10.7.7.7", "10.7.7.1", ]

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }

  subnet {
    name           = "subnet2"
    address_prefix = "10.10.1.0/24"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r VirtualNetworkResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_virtual_network" "import" {
  name                = azurestack_virtual_network.test.name
  location            = azurestack_virtual_network.test.location
  resource_group_name = azurestack_virtual_network.test.resource_group_name
  address_space       = ["10.0.0.0/16"]

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }
}
`, r.basic(data))
}

func (VirtualNetworkResource) withTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (VirtualNetworkResource) withTagsUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (VirtualNetworkResource) noSubnet(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  subnet              = []
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (VirtualNetworkResource) updateFlowTimeoutInMinutes(data acceptance.TestData, flowTimeout int) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                    = "acctestvirtnet%d"
  address_space           = ["10.0.0.0/16"]
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  flow_timeout_in_minutes = %d
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, flowTimeout)
}
