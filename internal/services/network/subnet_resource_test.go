// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type SubnetResource struct{}

func TestAccSubnet_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet", "test")
	r := SubnetResource{}

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

func TestAccSubnet_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_subnet"),
		},
	})
}

func TestAccSubnet_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccSubnet_updateAddressPrefix(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet", "test")
	r := SubnetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updatedAddressPrefix(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t SubnetResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading Subnet (%s): %+v", id, err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (SubnetResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.SubnetsClient.Delete(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting Subnet %q: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.SubnetsClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for deletion of Subnet %q: %+v", id, err)
	}

	return pointer.FromBool(true), nil
}

func (SubnetResource) hasNoNetworkSecurityGroup(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return err
	}

	resp, err := client.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Bad: Subnet %q (Virtual Network %q / Resource Group: %q) does not exist", id.Name, id.VirtualNetworkName, id.ResourceGroup)
		}

		return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
	}

	props := resp.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("Properties was nil for Subnet %q (Virtual Network %q / Resource Group: %q)", id.Name, id.VirtualNetworkName, id.ResourceGroup)
	}

	if props.NetworkSecurityGroup != nil && ((props.NetworkSecurityGroup.ID == nil) || (props.NetworkSecurityGroup.ID != nil && *props.NetworkSecurityGroup.ID == "")) {
		return fmt.Errorf("No Network Security Group should exist for Subnet %q (Virtual Network %q / Resource Group: %q) but got %q", id.Name, id.VirtualNetworkName, id.ResourceGroup, *props.NetworkSecurityGroup.ID)
	}

	return nil
}

func (SubnetResource) hasNoRouteTable(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return err
	}

	resp, err := client.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Bad: Subnet %q (Virtual Network %q / Resource Group: %q) does not exist", id.Name, id.VirtualNetworkName, id.ResourceGroup)
		}

		return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
	}

	props := resp.SubnetPropertiesFormat
	if props == nil {
		return fmt.Errorf("Properties was nil for Subnet %q (Virtual Network %q / Resource Group: %q)", id.Name, id.VirtualNetworkName, id.ResourceGroup)
	}

	if props.RouteTable != nil && ((props.RouteTable.ID == nil) || (props.RouteTable.ID != nil && *props.RouteTable.ID == "")) {
		return fmt.Errorf("No Route Table should exist for Subnet %q (Virtual Network %q / Resource Group: %q) but got %q", id.Name, id.VirtualNetworkName, id.ResourceGroup, *props.RouteTable.ID)
	}

	return nil
}

func (r SubnetResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_subnet" "test2" {
  name                 = "internal2"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.3.0/24"
}
`, r.template(data))
}

func (r SubnetResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "import" {
  name                 = azurestack_subnet.test.name
  resource_group_name  = azurestack_subnet.test.resource_group_name
  virtual_network_name = azurestack_subnet.test.virtual_network_name
  address_prefix       = azurestack_subnet.test.address_prefix
}
`, r.basic(data))
}

func (r SubnetResource) updatedAddressPrefix(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.3.0/24"
}
`, r.template(data))
}

func (SubnetResource) template(data acceptance.TestData) string {
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
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
