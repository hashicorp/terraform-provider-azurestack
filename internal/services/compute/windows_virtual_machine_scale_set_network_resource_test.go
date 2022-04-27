package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccWindowsVirtualMachineScaleSet_networkDNSServers(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkDNSServers(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			Config: r.networkDNSServersUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkIPForwarding(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// enabled
			Config: r.networkIPForwarding(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// disabled
			Config: r.networkPrivate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// enabled
			Config: r.networkIPForwarding(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkLoadBalancer(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkLoadBalancer(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkMultipleIPConfigurations(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkMultipleIPConfigurations(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkMultipleNICs(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkMultipleNICs(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkMultipleNICsMultipleIPConfigurations(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkMultipleNICsMultipleIPConfigurations(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkMultipleNICsWithDifferentDNSServers(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkMultipleNICsWithDifferentDNSServers(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkNetworkSecurityGroup(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkNetworkSecurityGroup(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkNetworkSecurityGroupUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// without
			Config: r.networkPrivate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// add one
			Config: r.networkNetworkSecurityGroup(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
		{
			// change it
			Config: r.networkNetworkSecurityGroupUpdated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func TestAccWindowsVirtualMachineScaleSet_networkPrivate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.networkPrivate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func (r WindowsVirtualMachineScaleSetResource) networkDNSServers(data acceptance.TestData) string {
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
  }

  network_interface {
    name        = "example"
    primary     = true
    dns_servers = ["8.8.8.8"]

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkDNSServersUpdated(data acceptance.TestData) string {
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
  }

  network_interface {
    name        = "example"
    primary     = true
    dns_servers = ["1.1.1.1", "8.8.8.8"]

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkIPForwarding(data acceptance.TestData) string {
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
  }

  network_interface {
    name                 = "example"
    primary              = true
    enable_ip_forwarding = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkLoadBalancer(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "internal"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "test"
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
}

resource "azurestack_lb_nat_pool" "test" {
  name                           = "test"
  resource_group_name            = azurestack_resource_group.test.name
  loadbalancer_id                = azurestack_lb.test.id
  frontend_ip_configuration_name = "internal"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 8080
}

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
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name                                   = "internal"
      primary                                = true
      subnet_id                              = azurestack_subnet.test.id
      load_balancer_backend_address_pool_ids = [azurestack_lb_backend_address_pool.test.id]
      load_balancer_inbound_nat_rules_ids    = [azurestack_lb_nat_pool.test.id]
    }
  }
}
`, r.template(data), data.RandomInteger, data.RandomInteger)
}

func (r WindowsVirtualMachineScaleSetResource) networkMultipleIPConfigurations(data acceptance.TestData) string {
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
  }

  network_interface {
    name    = "internal"
    primary = true

    ip_configuration {
      name      = "primary"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }

    ip_configuration {
      name      = "secondary"
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkMultipleNICs(data acceptance.TestData) string {
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
  }

  network_interface {
    name    = "primary"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }

  network_interface {
    name = "secondary"

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkMultipleNICsMultipleIPConfigurations(data acceptance.TestData) string {
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
  }

  network_interface {
    name    = "primary"
    primary = true

    ip_configuration {
      name      = "first"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }

    ip_configuration {
      name      = "second"
      subnet_id = azurestack_subnet.test.id
    }
  }

  network_interface {
    name = "secondary"

    ip_configuration {
      name      = "third"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }

    ip_configuration {
      name      = "fourth"
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkMultipleNICsWithDifferentDNSServers(data acceptance.TestData) string {
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
  }

  network_interface {
    name        = "primary"
    primary     = true
    dns_servers = ["8.8.8.8"]

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }

  network_interface {
    name        = "secondary"
    dns_servers = ["1.1.1.1"]

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data))
}

func (r WindowsVirtualMachineScaleSetResource) networkNetworkSecurityGroup(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

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
  }

  network_interface {
    name                      = "example"
    primary                   = true
    network_security_group_id = azurestack_network_security_group.test.id

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data), data.RandomInteger)
}

func (r WindowsVirtualMachineScaleSetResource) networkNetworkSecurityGroupUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_network_security_group" "other" {
  name                = "acctestnsg2-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

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
  }

  network_interface {
    name                      = "example"
    primary                   = true
    network_security_group_id = azurestack_network_security_group.other.id

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = azurestack_subnet.test.id
    }
  }
}
`, r.template(data), data.RandomInteger, data.RandomInteger)
}

func (r WindowsVirtualMachineScaleSetResource) networkPrivate(data acceptance.TestData) string {
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
