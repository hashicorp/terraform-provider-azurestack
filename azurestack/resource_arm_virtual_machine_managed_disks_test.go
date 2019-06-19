package azurestack

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_explicit(t *testing.T) {
	var vm compute.VirtualMachine
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_explicit(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_implicit(t *testing.T) {
	var vm compute.VirtualMachine
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_implicit(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_attach(t *testing.T) {
	var vm compute.VirtualMachine
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_attach(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_withDataDisk_managedDisk_explicit(t *testing.T) {
	var vm compute.VirtualMachine

	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_withDataDisk_managedDisk_explicit(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_withDataDisk_managedDisk_implicit(t *testing.T) {
	var vm compute.VirtualMachine

	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_withDataDisk_managedDisk_implicit(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_deleteManagedDiskOptOut(t *testing.T) {
	var vm compute.VirtualMachine
	var osd string
	var dtd string
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackVirtualMachine_withDataDisk_managedDisk_implicit(ri, location)
	postConfig := testAccAzureStackVirtualMachine_basicLinuxMachineDeleteVM_managedDisk(ri, location)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
					testLookupAzureStackVirtualMachineManagedDiskID(&vm, "myosdisk1", &osd),
					testLookupAzureStackVirtualMachineManagedDiskID(&vm, "mydatadisk1", &dtd),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineManagedDiskExists(&osd, true),
					testCheckAzureStackVirtualMachineManagedDiskExists(&dtd, true),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_deleteManagedDiskOptIn(t *testing.T) {
	var vm compute.VirtualMachine
	var osd string
	var dtd string
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_DestroyDisksBefore(ri, location)
	postConfig := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_DestroyDisksAfter(ri, location)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
					testLookupAzureStackVirtualMachineManagedDiskID(&vm, "myosdisk1", &osd),
					testLookupAzureStackVirtualMachineManagedDiskID(&vm, "mydatadisk1", &dtd),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineManagedDiskExists(&osd, false),
					testCheckAzureStackVirtualMachineManagedDiskExists(&dtd, false),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_osDiskTypeConflict(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_osDiskTypeConflict(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("conflicts with storage_os_disk.0.managed_disk_type"),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_dataDiskTypeConflict(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_dataDiskTypeConflict(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Conflict between `vhd_uri`"),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_bugAzureStack33(t *testing.T) {
	ri := acctest.RandInt()
	rs := acctest.RandString(7)
	config := testAccAzureStackVirtualMachine_bugAzureStack33(ri, rs, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_attachSecondDataDiskWithAttachOption(t *testing.T) {
	var afterCreate, afterUpdate compute.VirtualMachine
	resourceName := "azurestack_virtual_machine.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_empty(ri, location)
	postConfig := testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_attach(ri, location)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists(resourceName, &afterCreate),
					resource.TestCheckResourceAttr(resourceName, "storage_data_disk.0.create_option", "Empty"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists(resourceName, &afterUpdate),
					testAccCheckVirtualMachineRecreated(t, &afterCreate, &afterUpdate),
					resource.TestCheckResourceAttr(resourceName, "storage_data_disk.0.create_option", "Empty"),
					resource.TestCheckResourceAttr(resourceName, "storage_data_disk.1.create_option", "Attach"),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_linuxNoConfig(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_linuxNoConfig(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Error: either a `os_profile_linux_config` or a `os_profile_windows_config` must be specified."),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_windowsNoConfig(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_windowsNoConfig(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Error: either a `os_profile_linux_config` or a `os_profile_windows_config` must be specified."),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_multipleNICs(t *testing.T) {
	resourceName := "azurestack_virtual_machine.test"
	ri := acctest.RandInt()
	rs := acctest.RandString(5)
	subscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")
	prefix := fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/networkInterfaces", subscriptionId, ri)
	firstNicName := fmt.Sprintf("%s/acctni1-%d", prefix, ri)
	secondNicName := fmt.Sprintf("%s/acctni2-%d", prefix, ri)

	config := testAccAzureStackVirtualMachine_multipleNICs(ri, rs, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.0", firstNicName),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.1", secondNicName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachine_hasDiskInfoWhenStopped(t *testing.T) {
	var vm compute.VirtualMachine
	resourceName := "azurestack_virtual_machine.test"
	rInt := acctest.RandInt()
	config := testAccAzureStackVirtualMachine_hasDiskInfoWhenStopped(rInt, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists(resourceName, &vm),
					resource.TestCheckResourceAttr(resourceName, "storage_os_disk.0.managed_disk_type", "Standard_LRS"),
					resource.TestCheckResourceAttr(resourceName, "storage_data_disk.0.disk_size_gb", "64"),
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAndStopAzureStackVirtualMachine(&vm),
					resource.TestCheckResourceAttr(resourceName, "storage_os_disk.0.managed_disk_type", "Standard_LRS"),
					resource.TestCheckResourceAttr(resourceName, "storage_data_disk.0.disk_size_gb", "64"),
				),
			},
		},
	})
}

func testCheckAndStopAzureStackVirtualMachine(vm *compute.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vmID, err := parseAzureResourceID(*vm.ID)
		if err != nil {
			return fmt.Errorf("Unable to parse virtual machine ID %s, %+v", *vm.ID, err)
		}

		name := vmID.Path["virtualMachines"]
		resourceGroup := vmID.ResourceGroup

		client := testAccProvider.Meta().(*ArmClient).vmClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Deallocate(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Failed stopping virtual machine %q: %+v", resourceGroup, err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("Failed long polling for the stop of virtual machine %q: %+v", resourceGroup, err)
		}

		return nil
	}
}

func testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_explicit(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osd-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    disk_size_gb      = "50"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_implicit(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name          = "osd-%d"
    caching       = "ReadWrite"
    create_option = "FromImage"
    disk_size_gb  = "50"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_attach(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctmd2-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osd-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    disk_size_gb      = "50"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "acctmd-%d"
    create_option     = "Empty"
    disk_size_gb      = "1"
    lun               = 0
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name            = "${azurestack_managed_disk.test.name}"
    create_option   = "Attach"
    disk_size_gb    = "1"
    lun             = 1
    managed_disk_id = "${azurestack_managed_disk.test.id}"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_empty(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  delete_os_disk_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osd-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    disk_size_gb      = "50"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "acctmd-%d"
    create_option     = "Empty"
    disk_size_gb      = "1"
    managed_disk_type = "Standard_LRS"
    lun               = 0
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_DestroyDisksBefore(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name          = "myosdisk1"
    caching       = "ReadWrite"
    create_option = "FromImage"
  }

  delete_os_disk_on_termination = true

  storage_data_disk {
    name          = "mydatadisk1"
    disk_size_gb  = "1"
    create_option = "Empty"
    lun           = 0
  }

  delete_data_disks_on_termination = true

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_basicLinuxMachine_managedDisk_DestroyDisksAfter(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_basicLinuxMachineDeleteVM_managedDisk(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_withDataDisk_managedDisk_explicit(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osd-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "dtd-%d"
    disk_size_gb      = "1"
    create_option     = "Empty"
    caching           = "ReadWrite"
    lun               = 0
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_withDataDisk_managedDisk_implicit(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name          = "myosdisk1"
    caching       = "ReadWrite"
    create_option = "FromImage"
  }

  storage_data_disk {
    name          = "mydatadisk1"
    disk_size_gb  = "1"
    create_option = "Empty"
    caching       = "ReadWrite"
    lun           = 0
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_osDiskTypeConflict(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osd-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    disk_size_gb      = "10"
    managed_disk_type = "Standard_LRS"
    vhd_uri           = "should_cause_conflict"
  }

  storage_data_disk {
    name              = "mydatadisk1"
    caching           = "ReadWrite"
    create_option     = "Empty"
    disk_size_gb      = "45"
    managed_disk_type = "Standard_LRS"
    lun               = "0"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_dataDiskTypeConflict(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osd-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    disk_size_gb      = "10"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "mydatadisk1"
    caching           = "ReadWrite"
    create_option     = "Empty"
    disk_size_gb      = "45"
    managed_disk_type = "Standard_LRS"
    lun               = "0"
  }

  storage_data_disk {
    name              = "mydatadisk1"
    vhd_uri           = "should_cause_conflict"
    caching           = "ReadWrite"
    create_option     = "Empty"
    disk_size_gb      = "45"
    managed_disk_type = "Standard_LRS"
    lun               = "1"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_bugAzureStack33(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm%s"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_F1"

  storage_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "acctvm%s"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_windows_config {}

  tags = {
    environment = "Production"
    cost-center = "Ops"
  }
}
`, rInt, location, rInt, rInt, rInt, rString, rString)
}

func testCheckAzureStackVirtualMachineManagedDiskExists(managedDiskID *string, shouldExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		d, err := testGetAzureStackVirtualMachineManagedDisk(managedDiskID)
		if err != nil {
			return fmt.Errorf("Error trying to retrieve Managed Disk %s, %+v", *managedDiskID, err)
		}
		if d.StatusCode == http.StatusNotFound && shouldExist {
			return fmt.Errorf("Unable to find Managed Disk %s", *managedDiskID)
		}
		if d.StatusCode != http.StatusNotFound && !shouldExist {
			return fmt.Errorf("Found unexpected Managed Disk %s", *managedDiskID)
		}

		return nil
	}
}

func testLookupAzureStackVirtualMachineManagedDiskID(vm *compute.VirtualMachine, diskName string, managedDiskID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if osd := vm.StorageProfile.OsDisk; osd != nil {
			if strings.EqualFold(*osd.Name, diskName) {
				if osd.ManagedDisk != nil {
					id, err := findAzureStackVirtualMachineManagedDiskID(osd.ManagedDisk)
					if err != nil {
						return fmt.Errorf("Unable to parse Managed Disk ID for OS Disk %s, %+v", diskName, err)
					}
					*managedDiskID = id
					return nil
				}
			}
		}

		for _, dataDisk := range *vm.StorageProfile.DataDisks {
			if strings.EqualFold(*dataDisk.Name, diskName) {
				if dataDisk.ManagedDisk != nil {
					id, err := findAzureStackVirtualMachineManagedDiskID(dataDisk.ManagedDisk)
					if err != nil {
						return fmt.Errorf("Unable to parse Managed Disk ID for Data Disk %s, %+v", diskName, err)
					}
					*managedDiskID = id
					return nil
				}
			}
		}

		return fmt.Errorf("Unable to locate disk %s on vm %s", diskName, *vm.Name)
	}
}

func findAzureStackVirtualMachineManagedDiskID(md *compute.ManagedDiskParameters) (string, error) {
	if _, err := parseAzureResourceID(*md.ID); err != nil {
		return "", err
	}
	return *md.ID, nil
}

func testGetAzureStackVirtualMachineManagedDisk(managedDiskID *string) (*compute.Disk, error) {
	armID, err := parseAzureResourceID(*managedDiskID)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse Managed Disk ID %s, %+v", *managedDiskID, err)
	}
	name := armID.Path["disks"]
	resourceGroup := armID.ResourceGroup
	client := testAccProvider.Meta().(*ArmClient).diskClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext
	d, err := client.Get(ctx, resourceGroup, name)
	//check status first since sdk client returns error if not 200
	if d.Response.StatusCode == http.StatusNotFound {
		return &d, nil
	}
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func testAccAzureStackVirtualMachine_linuxNoConfig(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_F1"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "acctvm%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_windowsNoConfig(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_F1"

  storage_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "acctvm%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackVirtualMachine_multipleNICs(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "first" {
  name                = "acctni1-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_network_interface" "second" {
  name                = "acctni2-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                         = "acctvm%s"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  network_interface_ids        = ["${azurestack_network_interface.first.id}", "${azurestack_network_interface.second.id}"]
  primary_network_interface_id = "${azurestack_network_interface.first.id}"
  vm_size                      = "Standard_F1"

  storage_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "acctvm%s"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_windows_config {}
}
`, rInt, location, rInt, rInt, rInt, rInt, rString, rString)
}

func testAccAzureStackVirtualMachine_hasDiskInfoWhenStopped(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest-rg-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctestvm-%d"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_DS1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "acctest-osdisk-%d"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name          = "acctest-datadisk-%d"
    caching       = "ReadWrite"
    create_option = "Empty"
    lun           = 0
    disk_size_gb  = 64
  }

  os_profile {
    computer_name  = "acctest-machine-%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}
