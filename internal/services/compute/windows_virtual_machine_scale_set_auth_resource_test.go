package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccWindowsVirtualMachineScaleSet_authPassword(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_windows_virtual_machine_scale_set", "test")
	r := WindowsVirtualMachineScaleSetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.authPassword(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("admin_password"),
	})
}

func (r WindowsVirtualMachineScaleSetResource) authPassword(data acceptance.TestData) string {
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
