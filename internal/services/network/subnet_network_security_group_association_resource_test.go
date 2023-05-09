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

type SubnetNetworkSecurityGroupAssociationResource struct{}

func TestAccSubnetNetworkSecurityGroupAssociation_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet_network_security_group_association", "test")
	r := SubnetNetworkSecurityGroupAssociationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnetNetworkSecurityGroupAssociation_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet_network_security_group_association", "test")
	r := SubnetNetworkSecurityGroupAssociationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_subnet_network_security_group_association"),
		},
	})
}

func TestAccSubnetNetworkSecurityGroupAssociation_updateSubnet(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet_network_security_group_association", "test")
	r := SubnetNetworkSecurityGroupAssociationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateSubnet(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubnetNetworkSecurityGroupAssociation_deleted(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_subnet_network_security_group_association", "test")
	r := SubnetNetworkSecurityGroupAssociationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		// intentional not using a DisappearsStep as this is a Virtual Resource
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				data.CheckWithClient(r.destroy),
				data.CheckWithClientForResource(SubnetResource{}.hasNoNetworkSecurityGroup, "azurestack_subnet.test"),
			),
			ExpectNonEmptyPlan: true,
		},
	})
}

func (SubnetNetworkSecurityGroupAssociationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SubnetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading %s: %+v", *id, err)
	}

	props := resp.SubnetPropertiesFormat
	if props == nil || props.NetworkSecurityGroup == nil {
		return nil, fmt.Errorf("properties was nil for %s", *id)
	}

	return utils.Bool(props.NetworkSecurityGroup.ID != nil), nil
}

func (SubnetNetworkSecurityGroupAssociationResource) destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	subnetId := state.Attributes["subnet_id"]
	id, err := parse.SubnetID(subnetId)
	if err != nil {
		return err
	}

	read, err := client.Network.SubnetsClient.Get(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, "")
	if err != nil {
		if !utils.ResponseWasNotFound(read.Response) {
			return fmt.Errorf("retrieving Subnet %q (Network %q / Resource Group %q): %+v", id.Name, id.VirtualNetworkName, id.ResourceGroup, err)
		}
	}

	read.SubnetPropertiesFormat.NetworkSecurityGroup = nil

	future, err := client.Network.SubnetsClient.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualNetworkName, id.Name, read)
	if err != nil {
		return fmt.Errorf("updating Subnet %q (Network %q / Resource Group %q): %+v", id.Name, id.VirtualNetworkName, id.ResourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Network.SubnetsClient.Client); err != nil {
		return fmt.Errorf("waiting for completion of Subnet %q (Network %q / Resource Group %q): %+v", id.Name, id.VirtualNetworkName, id.ResourceGroup, err)
	}

	return nil
}

func (r SubnetNetworkSecurityGroupAssociationResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_subnet_network_security_group_association" "test" {
  subnet_id                 = azurestack_subnet.test.id
  network_security_group_id = azurestack_network_security_group.test.id
}
`, r.template(data))
}

func (r SubnetNetworkSecurityGroupAssociationResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet_network_security_group_association" "internal" {
  subnet_id                 = azurestack_subnet_network_security_group_association.test.subnet_id
  network_security_group_id = azurestack_subnet_network_security_group_association.test.network_security_group_id
}
`, r.basic(data))
}

func (r SubnetNetworkSecurityGroupAssociationResource) updateSubnet(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_subnet_network_security_group_association" "test" {
  subnet_id                 = azurestack_subnet.test.id
  network_security_group_id = azurestack_network_security_group.test.id
}
`, r.template(data))
}

func (SubnetNetworkSecurityGroupAssociationResource) template(data acceptance.TestData) string {
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

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
