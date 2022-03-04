package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccLinuxVirtualMachine_diskOSBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSCachingType(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSCachingType(data, "None"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.diskOSCachingType(data, "ReadOnly"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.diskOSCachingType(data, "ReadWrite"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSCustomName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSCustomName(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSCustomSize(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSCustomSize(data, 30),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSCustomSizeExpanded(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSCustomSize(data, 30),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.diskOSCustomSize(data, 60),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSEphemeral(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSEphemeral(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSStorageTypeStandardLRS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSStorageAccountType(data, "Standard_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSStorageTypePremiumLRS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSStorageAccountType(data, "Premium_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSStorageTypeUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.diskOSStorageAccountType(data, "Standard_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.diskOSStorageAccountType(data, "Premium_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.diskOSStorageAccountType(data, "Standard_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLinuxVirtualMachine_diskOSWriteAcceleratorEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine", "test")
	r := LinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// Enabled
			Config: r.diskOSWriteAcceleratorEnabled(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			// Disabled
			Config: r.diskOSWriteAcceleratorEnabled(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			// Enabled
			Config: r.diskOSWriteAcceleratorEnabled(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r LinuxVirtualMachineResource) diskOSBasic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r LinuxVirtualMachineResource) diskOSCachingType(data acceptance.TestData, cachingType string) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    caching              = "%s"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger, cachingType)
}

func (r LinuxVirtualMachineResource) diskOSCustomName(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    name                 = "osdisk1"
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r LinuxVirtualMachineResource) diskOSCustomSize(data acceptance.TestData, size int) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    name                 = "osdisk1"
    caching              = "ReadWrite"
    disk_size_gb         = %d
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger, size)
}

func (r LinuxVirtualMachineResource) diskOSEphemeral(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    caching              = "ReadOnly"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger)
}

func (r LinuxVirtualMachineResource) diskOSStorageAccountType(data acceptance.TestData, accountType string) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2s_v2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "%s"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger, accountType)
}

func (r LinuxVirtualMachineResource) diskOSWriteAcceleratorEnabled(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine" "test" {
  name                = "acctestVM-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  size                = "Standard_F2s_v2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  admin_ssh_key {
    username   = "adminuser"
    public_key = local.first_public_key
  }

  os_disk {
    storage_account_type      = "Standard_LRS"
    caching                   = "None"
    write_accelerator_enabled = %t
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, r.template(data), data.RandomInteger, enabled)
}
