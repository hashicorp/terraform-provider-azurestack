package compute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type ImageResource struct{}

func TestImageReource_standaloneImage(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_image", "test")
	r := ImageResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// need to create a vm and then reference it in the image creation
			Config: r.setupUnmanagedDisks(data, "LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				data.CheckWithClientForResource(r.virtualMachineExists, "azurestack_virtual_machine.testsource"),
			),
		},
		{
			Config: r.standaloneImageProvision(data, "LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r ImageResource) standaloneImageProvision(data acceptance.TestData, storageType string) string {
	template := r.setupUnmanagedDisks(data, storageType)
	return fmt.Sprintf(`
%s

resource "azurestack_image" "test" {
  name                = "accteste"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  zone_resilient      = false

  os_disk {
    os_type  = "Linux"
    os_state = "Generalized"
    blob_uri = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
    size_gb  = 30
    caching  = "None"
  }
}
`, template)
}

func (ImageResource) virtualMachineExists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	id, err := parse.VirtualMachineID(state.ID)
	if err != nil {
		return err
	}

	resp, err := client.Compute.VMClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s does not exist", *id)
		}

		return fmt.Errorf("Bad: Get on client: %+v", err)
	}

	return nil
}

func (ImageResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ImageID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.ImageClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Image %q", id)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (r ImageResource) setupUnmanagedDisks(data acceptance.TestData, storageType string) string {
	template := r.template(data)
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_network_interface" "testsource" {
  name                = "acctnicsource-${local.number}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  ip_configuration {
    name                          = "testconfigurationsource"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurestack_public_ip.test.id
  }
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa${local.random_string}"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "%s"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "blob"
}

resource "azurestack_virtual_machine" "testsource" {
  name                  = "testsource"
  location              = azurestack_resource_group.test.location
  resource_group_name   = azurestack_resource_group.test.name
  network_interface_ids = [azurestack_network_interface.testsource.id]
  vm_size               = "Standard_F2s_v2"
  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name          = "myosdisk1"
    vhd_uri       = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
    caching       = "ReadWrite"
    create_option = "FromImage"
  }

  os_profile {
    computer_name  = "mdimagetestsource"
    admin_username = local.admin_username
    admin_password = local.admin_password
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}
`, template, storageType)
}

func (ImageResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
locals {
  number         = "%d"
  location       = %q
  random_string  = %q
  admin_username = "testadmin%d"
  admin_password = "Password1234!%d"
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-${local.number}"
  location = local.location
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-${local.number}"
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

resource "azurestack_public_ip" "test" {
  name                = "acctpip-${local.number}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, data.RandomInteger)
}
