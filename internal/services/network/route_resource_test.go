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

type RouteResource struct{}

func TestAccRoute_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_route", "test")
	r := RouteResource{}

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

func TestAccRoute_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_route", "test")
	r := RouteResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_route"),
		},
	})
}

func TestAccRoute_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_route", "test")
	r := RouteResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("next_hop_type").HasValue("VnetLocal"),
				check.That(data.ResourceName).Key("next_hop_in_ip_address").HasValue(""),
			),
		},
		{
			Config: r.basicAppliance(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("next_hop_type").HasValue("VirtualAppliance"),
				check.That(data.ResourceName).Key("next_hop_in_ip_address").HasValue("192.168.0.1"),
			),
		},
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("next_hop_type").HasValue("VnetLocal"),
				check.That(data.ResourceName).Key("next_hop_in_ip_address").HasValue(""),
			),
		},
	})
}

func TestAccRoute_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_route", "test")
	r := RouteResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccRoute_multipleRoutes(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_route", "test")
	r := RouteResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},

		{
			Config: r.multipleRoutes(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func (t RouteResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.RouteID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.RoutesClient.Get(ctx, id.ResourceGroup, id.RouteTableName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading Route (%s): %+v", *id, err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (r RouteResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.RouteID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.RoutesClient.Delete(ctx, id.ResourceGroup, id.RouteTableName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting on routesClient: %+v", err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.RoutesClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for deletion of Route %q: %+v", id, err)
	}

	return pointer.FromBool(true), nil
}

func (RouteResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_route" "test" {
  name                = "acctestroute%d"
  resource_group_name = azurestack_resource_group.test.name
  route_table_name    = azurestack_route_table.test.name

  address_prefix = "10.1.0.0/16"
  next_hop_type  = "vnetlocal"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (r RouteResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurestack_route" "import" {
  name                = azurestack_route.test.name
  resource_group_name = azurestack_route.test.resource_group_name
  route_table_name    = azurestack_route.test.route_table_name

  address_prefix = azurestack_route.test.address_prefix
  next_hop_type  = azurestack_route.test.next_hop_type
}
`, r.basic(data))
}

func (RouteResource) basicAppliance(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_route" "test" {
  name                = "acctestroute%d"
  resource_group_name = azurestack_resource_group.test.name
  route_table_name    = azurestack_route_table.test.name

  address_prefix         = "10.1.0.0/16"
  next_hop_type          = "VirtualAppliance"
  next_hop_in_ip_address = "192.168.0.1"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (RouteResource) multipleRoutes(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_route" "test" {
  name                = "acctestroute%d"
  resource_group_name = azurestack_resource_group.test.name
  route_table_name    = azurestack_route_table.test.name

  address_prefix = "10.1.0.0/16"
  next_hop_type  = "vnetlocal"
}

resource "azurestack_route" "test1" {
  name                = "acctestroute%d1"
  resource_group_name = azurestack_resource_group.test.name
  route_table_name    = azurestack_route_table.test.name

  address_prefix = "10.2.0.0/16"
  next_hop_type  = "none"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
