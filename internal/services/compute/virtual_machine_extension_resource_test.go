package compute_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type VirtualMachineExtensionResource struct{}

func TestAccVirtualMachineExtension_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_machine_extension", "test")
	r := VirtualMachineExtensionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestMatchResourceAttr(data.ResourceName, "settings", regexp.MustCompile("hostname")),
			),
		},
		data.ImportStep("protected_settings"),
		{
			Config: r.basicUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestMatchResourceAttr(data.ResourceName, "settings", regexp.MustCompile("whoami")),
			),
		},
		data.ImportStep("protected_settings"),
	})
}

func TestAccVirtualMachineExtension_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_machine_extension", "test")
	r := VirtualMachineExtensionResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_virtual_machine_extension"),
		},
	})
}

func TestAccVirtualMachineExtension_concurrent(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_virtual_machine_extension", "test")
	r := VirtualMachineExtensionResource{}
	secondResourceName := "azurestack_virtual_machine_extension.test2"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.concurrent(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestMatchResourceAttr(data.ResourceName, "settings", regexp.MustCompile("hostname")),
				acceptance.TestMatchResourceAttr(secondResourceName, "settings", regexp.MustCompile("whoami")),
			),
		},
	})
}

func (t VirtualMachineExtensionResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VirtualMachineExtensionID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.VMExtensionClient.Get(ctx, id.ResourceGroup, id.VirtualMachineName, id.ExtensionName, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Virtual Machine Extension %q", id.String())
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (VirtualMachineExtensionResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%d"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = azurestack_resource_group.test.location
  resource_group_name   = azurestack_resource_group.test.name
  network_interface_ids = [azurestack_network_interface.test.id]
  vm_size               = "Standard_F2"

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
    computer_name  = "hostname%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurestack_virtual_machine_extension" "test" {
  name                 = "acctvme-%d"
  virtual_machine_id   = azurestack_virtual_machine.test.id
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
	{
		"commandToExecute": "hostname"
	}
SETTINGS


  tags = {
    environment = "Production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r VirtualMachineExtensionResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_virtual_machine_extension" "import" {
  name                 = azurestack_virtual_machine_extension.test.name
  virtual_machine_id   = azurestack_virtual_machine_extension.test.virtual_machine_id
  publisher            = azurestack_virtual_machine_extension.test.publisher
  type                 = azurestack_virtual_machine_extension.test.type
  type_handler_version = azurestack_virtual_machine_extension.test.type_handler_version
  settings             = azurestack_virtual_machine_extension.test.settings
  tags                 = azurestack_virtual_machine_extension.test.tags
}
`, r.basic(data))
}

func (VirtualMachineExtensionResource) basicUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%d"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = azurestack_resource_group.test.location
  resource_group_name   = azurestack_resource_group.test.name
  network_interface_ids = [azurestack_network_interface.test.id]
  vm_size               = "Standard_F2"

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
    computer_name  = "hostname%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurestack_virtual_machine_extension" "test" {
  name                 = "acctvme-%d"
  virtual_machine_id   = azurestack_virtual_machine.test.id
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
	{
		"commandToExecute": "whoami"
	}
SETTINGS


  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualMachineExtensionResource) concurrent(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%d"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = azurestack_resource_group.test.location
  resource_group_name   = azurestack_resource_group.test.name
  network_interface_ids = [azurestack_network_interface.test.id]
  vm_size               = "Standard_F2"

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
    computer_name  = "hostname%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurestack_virtual_machine_extension" "test" {
  name                 = "acctvme-%d"
  virtual_machine_id   = azurestack_virtual_machine.test.id
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
	{
		"commandToExecute": "hostname"
	}
SETTINGS

}

resource "azurestack_virtual_machine_extension" "test2" {
  name                 = "acctvme-%d-2"
  virtual_machine_id   = azurestack_virtual_machine.test.id
  publisher            = "Microsoft.OSTCExtensions"
  type                 = "CustomScriptForLinux"
  type_handler_version = "1.5"

  settings = <<SETTINGS
	{
		"commandToExecute": "whoami"
	}
SETTINGS

}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
