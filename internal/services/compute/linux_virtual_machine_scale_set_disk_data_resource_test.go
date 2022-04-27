package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskBasic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskCaching(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskCaching(data, "None"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.disksDataDiskCaching(data, "ReadOnly"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.disksDataDiskCaching(data, "ReadWrite"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskResizing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// 30GB
			Config: r.disksDataDiskResize(data, 30),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// 60GB
			Config: r.disksDataDiskResize(data, 60),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskMultiple(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskMultiple(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskRemove(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.authPassword(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskScaling(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// no disks
			Config: r.authPassword(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// one disk
			Config: r.disksDataDiskBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// two disks
			Config: r.disksDataDiskMultiple(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// no disks
			Config: r.authPassword(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskStorageAccountTypeStandardLRS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskStorageAccountType(data, "Standard_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskStorageAccountTypePremiumLRS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskStorageAccountType(data, "Premium_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccLinuxVirtualMachineScaleSet_disksDataDiskWriteAcceleratorEnabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_linux_virtual_machine_scale_set", "test")
	r := LinuxVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.disksDataDiskWriteAcceleratorEnabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func (r LinuxVirtualMachineScaleSetResource) disksDataDiskBasic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  data_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
    disk_size_gb         = 10
    lun                  = 10
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
`, r.template(data), data.RandomInteger)
}

func (r LinuxVirtualMachineScaleSetResource) disksDataDiskCaching(data acceptance.TestData, caching string) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  data_disk {
    storage_account_type = "Standard_LRS"
    caching              = %q
    disk_size_gb         = 10
    lun                  = 10
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
`, r.template(data), data.RandomInteger, caching)
}

func (r LinuxVirtualMachineScaleSetResource) disksDataDiskResize(data acceptance.TestData, diskSizeGb int) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  data_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
    lun                  = 10
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
`, r.template(data), data.RandomInteger, diskSizeGb)
}

func (r LinuxVirtualMachineScaleSetResource) disksDataDiskMultiple(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  data_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
    disk_size_gb         = 10
    lun                  = 10
  }

  data_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
    disk_size_gb         = 10
    lun                  = 20
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
`, r.template(data), data.RandomInteger)
}

func (r LinuxVirtualMachineScaleSetResource) disksDataDiskStorageAccountType(data acceptance.TestData, storageAccountType string) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2s_v2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  data_disk {
    storage_account_type = %q
    caching              = "None"
    disk_size_gb         = 10
    lun                  = 10
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
`, r.template(data), data.RandomInteger, storageAccountType)
}

func (r LinuxVirtualMachineScaleSetResource) disksDataDiskWriteAcceleratorEnabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_DS14_v2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Premium_LRS"
    caching              = "None"
  }

  data_disk {
    storage_account_type      = "Premium_LRS"
    caching                   = "None"
    disk_size_gb              = 10
    lun                       = 10
    write_accelerator_enabled = true
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
`, r.template(data), data.RandomInteger)
}
