package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccWindowsVirtualMachineScaleSet_disksOSDiskCaching(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksOSDiskCaching(data, "None"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.disksOSDiskCaching(data, "ReadOnly"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.disksOSDiskCaching(data, "ReadWrite"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_disksOSDiskCustomSize(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// unset
			Config: r.authPassword(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.disksOSDiskCustomSize(data, 128),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// resize a second time to confirm https://github.com/Azure/azure-rest-api-specs/issues/1906
			Config: r.disksOSDiskCustomSize(data, 256),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_disksOSDiskEphemeral(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksOSDiskEphemeral(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_disksOSDiskStorageAccountTypeStandardLRS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksOSDiskStorageAccountType(data, "Standard_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_disksOSDiskStorageAccountTypePremiumLRS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksOSDiskStorageAccountType(data, "Premium_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_disksOSDiskWriteAcceleratorEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksOSDiskWriteAcceleratorEnabled(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func (r WindowsVirtualMachineScaleSetResource) disksOSDiskCaching(data acceptance.TestData, caching string) string {
	return fmt.Sprintf(`
%s

resource "azurestack_windows_virtual_machine_scale_set" "test" {
  name                = local.vm_name
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2012-Datacenter-smalldisk"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "%s"
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data), caching)
}

func (r WindowsVirtualMachineScaleSetResource) disksOSDiskCustomSize(data acceptance.TestData, diskSize int) string {
	return fmt.Sprintf(`
%s

resource "azurestack_windows_virtual_machine_scale_set" "test" {
  name                = local.vm_name
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2012-Datacenter-smalldisk"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
    disk_size_gb         = %d
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data), diskSize)
}

func (r WindowsVirtualMachineScaleSetResource) disksOSDiskEphemeral(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_windows_virtual_machine_scale_set" "test" {
  name                = local.vm_name
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F8s_v2" # has to be this large for ephemeral disks on Windows
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2012-Datacenter-smalldisk"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadOnly"
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) disksOSDiskStorageAccountType(data acceptance.TestData, storageAccountType string) string {
	return fmt.Sprintf(`
%s

resource "azurestack_windows_virtual_machine_scale_set" "test" {
  name                = local.vm_name
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2s_v2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2012-Datacenter-smalldisk"
    version   = "latest"
  }

  os_disk {
    storage_account_type = %q
    caching              = "ReadWrite"
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data), storageAccountType)
}

func (r WindowsVirtualMachineScaleSetResource) disksOSDiskWriteAcceleratorEnabled(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "azurestack_windows_virtual_machine_scale_set" "test" {
  name                = local.vm_name
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_DS14_v2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2012-Datacenter-smalldisk"
    version   = "latest"
  }

  os_disk {
    storage_account_type      = "Premium_LRS"
    caching                   = "None"
    write_accelerator_enabled = %t
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data), enabled)
}
