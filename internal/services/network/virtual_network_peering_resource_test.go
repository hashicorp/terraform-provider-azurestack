// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

type VirtualNetworkPeeringResource struct{}

func TestAccVirtualNetworkPeering_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_peering", "test1")
	r := VirtualNetworkPeeringResource{}
	secondResourceName := "azurestack_virtual_network_peering.test2"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(secondResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("allow_virtual_network_access").HasValue("true"),
				acceptance.TestCheckResourceAttr(secondResourceName, "allow_virtual_network_access", "true"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualNetworkPeering_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_peering", "test1")
	r := VirtualNetworkPeeringResource{}
	secondResourceName := "azurestack_virtual_network_peering.test2"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(secondResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccVirtualNetworkPeering_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_peering", "test1")
	r := VirtualNetworkPeeringResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccVirtualNetworkPeering_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_network_peering", "test1")
	r := VirtualNetworkPeeringResource{}
	secondResourceName := "azurestack_virtual_network_peering.test2"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(secondResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("allow_virtual_network_access").HasValue("true"),
				acceptance.TestCheckResourceAttr(secondResourceName, "allow_virtual_network_access", "true"),
				check.That(data.ResourceName).Key("allow_forwarded_traffic").HasValue("false"),
				acceptance.TestCheckResourceAttr(secondResourceName, "allow_forwarded_traffic", "false"),
			),
		},
		data.ImportStep(),
		{
			Config: r.basicUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(secondResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("allow_virtual_network_access").HasValue("true"),
				acceptance.TestCheckResourceAttr(secondResourceName, "allow_virtual_network_access", "true"),
				check.That(data.ResourceName).Key("allow_forwarded_traffic").HasValue("true"),
				acceptance.TestCheckResourceAttr(secondResourceName, "allow_forwarded_traffic", "true"),
			),
		},
		data.ImportStep(),
	})
}

func (t VirtualNetworkPeeringResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualNetworkPeeringID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := clients.Network.VnetPeeringsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", *id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r VirtualNetworkPeeringResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualNetworkPeeringID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.VnetPeeringsClient.Delete(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting on virtual network peering: %+v", err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.VnetPeeringsClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return utils.Bool(true), nil
}

func (VirtualNetworkPeeringResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test1" {
  name                = "acctestvirtnet-1-%d"
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.1.0/24"]
  location            = azurestack_resource_group.test.location
}

resource "azurestack_virtual_network" "test2" {
  name                = "acctestvirtnet-2-%d"
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.2.0/24"]
  location            = azurestack_resource_group.test.location
}

resource "azurestack_virtual_network_peering" "test1" {
  name                         = "acctestpeer-1-%d"
  resource_group_name          = azurestack_resource_group.test.name
  virtual_network_name         = azurestack_virtual_network.test1.name
  remote_virtual_network_id    = azurestack_virtual_network.test2.id
  allow_virtual_network_access = true
}

resource "azurestack_virtual_network_peering" "test2" {
  name                         = "acctestpeer-2-%d"
  resource_group_name          = azurestack_resource_group.test.name
  virtual_network_name         = azurestack_virtual_network.test2.name
  remote_virtual_network_id    = azurestack_virtual_network.test1.id
  allow_virtual_network_access = true
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r VirtualNetworkPeeringResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurestack_virtual_network_peering" "import" {
  name                         = azurestack_virtual_network_peering.test1.name
  resource_group_name          = azurestack_virtual_network_peering.test1.resource_group_name
  virtual_network_name         = azurestack_virtual_network_peering.test1.virtual_network_name
  remote_virtual_network_id    = azurestack_virtual_network_peering.test1.remote_virtual_network_id
  allow_virtual_network_access = azurestack_virtual_network_peering.test1.allow_virtual_network_access
}
`, r.basic(data))
}

func (VirtualNetworkPeeringResource) basicUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test1" {
  name                = "acctestvirtnet-1-%d"
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.1.0/24"]
  location            = azurestack_resource_group.test.location
}

resource "azurestack_virtual_network" "test2" {
  name                = "acctestvirtnet-2-%d"
  resource_group_name = azurestack_resource_group.test.name
  address_space       = ["10.0.2.0/24"]
  location            = azurestack_resource_group.test.location
}

resource "azurestack_virtual_network_peering" "test1" {
  name                         = "acctestpeer-1-%d"
  resource_group_name          = azurestack_resource_group.test.name
  virtual_network_name         = azurestack_virtual_network.test1.name
  remote_virtual_network_id    = azurestack_virtual_network.test2.id
  allow_forwarded_traffic      = true
  allow_virtual_network_access = true
}

resource "azurestack_virtual_network_peering" "test2" {
  name                         = "acctestpeer-2-%d"
  resource_group_name          = azurestack_resource_group.test.name
  virtual_network_name         = azurestack_virtual_network.test2.name
  remote_virtual_network_id    = azurestack_virtual_network.test1.id
  allow_forwarded_traffic      = true
  allow_virtual_network_access = true
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
