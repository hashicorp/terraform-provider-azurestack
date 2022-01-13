package azurestack

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAzureStackVirtualMachineScaleSet_basic(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basic(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"os_profile.0.admin_password", "priority", "single_placement_group"},
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicPublicIP(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basicPublicIP(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetIsPrimary(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_networkSecurityGroup(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_networkSecurityGroup(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicWindows(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basicWindows(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_linuxUpdated(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccAzureStackVirtualMachineScaleSet_linux(ri, location)
	updatedConfig := testAccAzureStackVirtualMachineScaleSet_linuxUpdated(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicLinux_managedDisk(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basicLinux_managedDisk(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicWindows_managedDisk(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basicWindows_managedDisk(ri, testLocation(), "Standard_D1_v2")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicWindows_managedDisk_resize(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	preConfig := testAccAzureStackVirtualMachineScaleSet_basicWindows_managedDisk(ri, testLocation(), "Standard_D1_v2")
	postConfig := testAccAzureStackVirtualMachineScaleSet_basicWindows_managedDisk(ri, testLocation(), "Standard_D2_v2")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicLinux_managedDiskNoName(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basicLinux_managedDiskNoName(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_basicLinux_disappears(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_basic(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_planManagedDisk(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_planManagedDisk(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_loadBalancer(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSetLoadBalancerTemplate(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetHasLoadbalancer(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_loadBalancerManagedDataDisks(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSetLoadBalancerTemplateManagedDataDisks(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetHasDataDisks(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_overprovision(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSetOverProvisionTemplate(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_priority(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSetPriorityTemplate(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "priority", "Low"),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_extension(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSetExtensionTemplate(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetExtension(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_extensionUpdate(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccAzureStackVirtualMachineScaleSetExtensionTemplate(ri, location)
	updatedConfig := testAccAzureStackVirtualMachineScaleSetExtensionTemplateUpdated(ri, location)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetExtension(resourceName),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetExtension(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_multipleExtensions(t *testing.T) {
	resourceName := "azurestack_virtual_machine_scale_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSetMultipleExtensionsTemplate(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineScaleSetExists(resourceName),
					testCheckAzureStackVirtualMachineScaleSetExtension(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualMachineScaleSet_osDiskTypeConflict(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualMachineScaleSet_osDiskTypeConflict(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualMachineScaleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Conflict between `vhd_containers`"),
			},
		},
	})
}

func testGetAzureStackVirtualMachineScaleSet(s *terraform.State, resourceName string) (result *compute.VirtualMachineScaleSet, err error) {
	// Ensure we have enough information in state to look up in API
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", resourceName)
	}

	// Name of the actual scale set
	name := rs.Primary.Attributes["name"]

	resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
	if !hasResourceGroup {
		return nil, fmt.Errorf("Bad: no resource group found in state for virtual machine: scale set %s", name)
	}

	client := testAccProvider.Meta().(*ArmClient).vmScaleSetClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	vmss, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return nil, fmt.Errorf("Bad: Get on vmScaleSetClient: %+v", err)
	}

	if vmss.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Bad: VirtualMachineScaleSet %q (resource group: %q) does not exist", name, resourceGroup)
	}

	return &vmss, err
}

// nolint:unparam
func testCheckAzureStackVirtualMachineScaleSetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetAzureStackVirtualMachineScaleSet(s, name)
		return err
	}
}

func testCheckAzureStackVirtualMachineScaleSetDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for virtual machine: scale set %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).vmScaleSetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Delete on vmScaleSetClient: %+v", err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Bad: Delete on vmScaleSetClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackVirtualMachineScaleSetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).vmScaleSetClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_virtual_machine_scale_set" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Virtual Machine Scale Set still exists:\n%#v", resp.VirtualMachineScaleSetProperties)
		}
	}

	return nil
}

func testCheckAzureStackVirtualMachineScaleSetHasLoadbalancer(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resp, err := testGetAzureStackVirtualMachineScaleSet(s, name)
		if err != nil {
			return err
		}

		n := resp.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		if n == nil || len(*n) == 0 {
			return fmt.Errorf("Bad: Could not get network interface configurations for scale set %v", name)
		}

		ip := (*n)[0].IPConfigurations
		if ip == nil || len(*ip) == 0 {
			return fmt.Errorf("Bad: Could not get ip configurations for scale set %v", name)
		}

		pools := (*ip)[0].LoadBalancerBackendAddressPools
		if pools == nil || len(*pools) == 0 {
			return fmt.Errorf("Bad: Load balancer backend pools is empty for scale set %v", name)
		}

		return nil
	}
}

func testCheckAzureStackVirtualMachineScaleSetIsPrimary(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resp, err := testGetAzureStackVirtualMachineScaleSet(s, name)
		if err != nil {
			return err
		}

		n := resp.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
		if n == nil || len(*n) == 0 {
			return fmt.Errorf("Bad: Could not get network interface configurations for scale set %v", name)
		}

		ip := (*n)[0].IPConfigurations
		if ip == nil || len(*ip) == 0 {
			return fmt.Errorf("Bad: Could not get ip configurations for scale set %v", name)
		}

		// primary := *(*ip)[0].Primary
		// if primary != boolean {
		// 	return fmt.Errorf("Bad: Primary set incorrectly for scale set %v\n Wanted: %+v Received: %+v", name, boolean, primary)
		// }

		return nil
	}
}

// nolint:unparam
func testCheckAzureStackVirtualMachineScaleSetExtension(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resp, err := testGetAzureStackVirtualMachineScaleSet(s, name)
		if err != nil {
			return err
		}

		n := resp.VirtualMachineProfile.ExtensionProfile.Extensions
		if n == nil || len(*n) == 0 {
			return fmt.Errorf("Bad: Could not get extensions for scale set %v", name)
		}

		return nil
	}
}

func testCheckAzureStackVirtualMachineScaleSetHasDataDisks(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for virtual machine: scale set %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).vmScaleSetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on vmScaleSetClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: VirtualMachineScaleSet %q (resource group: %q) does not exist", name, resourceGroup)
		}

		// storageProfile := resp.VirtualMachineProfile.StorageProfile.DataDisks
		// if storageProfile == nil || len(*storageProfile) == 0 {
		// 	return fmt.Errorf("Bad: Could not get data disks configurations for scale set %v", name)
		// }

		return nil
	}
}

func testAccAzureStackVirtualMachineScaleSet_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "osDiskProfile"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_basicPublicIP(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
      primary   = true

      public_ip_address_configuration {
        name              = "TestPublicIPConfiguration"
        domain_name_label = "test-domain-label-%[1]d"
        idle_timeout      = 4
      }
    }
  }

  storage_profile_os_disk {
    name           = "osDiskProfile"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_networkSecurityGroup(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name                      = "TestNetworkProfile-%[1]d"
    primary                   = true
    network_security_group_id = "${azurestack_network_security_group.test.id}"

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
      primary   = true

      public_ip_address_configuration {
        name              = "TestPublicIPConfiguration"
        domain_name_label = "test-domain-label-%[1]d"
        idle_timeout      = 4
      }
    }
  }

  storage_profile_os_disk {
    name           = "osDiskProfile"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_basicWindows(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  os_profile_windows_config {
    enable_automatic_upgrades = false
    provision_vm_agent        = true

    winrm {
      protocol = "http"
    }
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "osDiskProfile"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter-Server-Core"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_linux(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"
  address_space       = ["10.0.0.0/8"]
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsn-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "acctestsc-%[1]d"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpip-%[1]d"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  location                     = "${azurestack_resource_group.test.location}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"

  frontend_ip_configuration {
    name                 = "ip-address"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "acctestbap-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"
  upgrade_policy_mode = "Automatic"

  sku {
    name     = "Standard_A0"
    tier     = "Standard"
    capacity = "1"
  }

  os_profile {
    computer_name_prefix = "prefix"
    admin_username       = "ubuntu"
    admin_password       = "password"
    custom_data          = "custom data!"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/ubuntu/.ssh/authorized_keys"
      key_data = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCsTcryUl51Q2VSEHqDRNmceUFo55ZtcIwxl2QITbN1RREti5ml/VTytC0yeBOvnZA4x4CFpdw/lCDPk0yrH9Ei5vVkXmOrExdTlT3qI7YaAzj1tUVlBd4S6LX1F7y6VLActvdHuDDuXZXzCDd/97420jrDfWZqJMlUK/EmCE5ParCeHIRIvmBxcEnGfFIsw8xQZl0HphxWOtJil8qsUWSdMyCiJYYQpMoMliO99X40AUc4/AlsyPyT5ddbKk08YrZ+rKDVHF7o29rh4vi5MmHkVgVQHKiKybWlHq+b71gIAUQk9wrJxD+dqt4igrmDSpIjfjwnd+l5UIn5fJSO5DYV4YT/4hwK7OKmuo7OFHD0WyY5YnkYEMtFgzemnRBdE8ulcT60DQpVgRMXFWHvhyCWy0L6sgj1QWDZlLpvsIvNfHsyhKFMG1frLnMt/nP0+YCcfg+v1JYeCKjeoJxB8DWcRBsjzItY0CGmzP8UYZiYKl/2u+2TgFS5r7NWH11bxoUzjKdaa1NLw+ieA8GlBFfCbfWe6YVB9ggUte4VtYFMZGxOjS2bAiYtfgTKFJv+XqORAwExG6+G2eDxIDyo80/OA9IG7Xv/jwQr7D6KDjDuULFcN/iTxuttoKrHeYz1hf5ZQlBdllwJHYx6fK2g8kha6r2JIQKocvsAXiiONqSfw== hello@world.com"
    }
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name                                   = "TestIPConfiguration"
      subnet_id                              = "${azurestack_subnet.test.id}"
      load_balancer_backend_address_pool_ids = ["${azurestack_lb_backend_address_pool.test.id}"]
    }
  }

  storage_profile_os_disk {
    name           = "osDiskProfile"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    os_type        = "linux"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_linuxUpdated(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"
  address_space       = ["10.0.0.0/8"]
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsn-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "acctestsc-%[1]d"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpip-%[1]d"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  location                     = "${azurestack_resource_group.test.location}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"

  frontend_ip_configuration {
    name                 = "ip-address"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "acctestbap-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctestvmss-%[1]d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"
  upgrade_policy_mode = "Automatic"

  sku {
    name     = "Standard_A0"
    tier     = "Standard"
    capacity = "1"
  }

  os_profile {
    computer_name_prefix = "prefix"
    admin_username       = "ubuntu"
    admin_password       = "password"
    custom_data          = "custom data!"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/ubuntu/.ssh/authorized_keys"
      key_data = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCsTcryUl51Q2VSEHqDRNmceUFo55ZtcIwxl2QITbN1RREti5ml/VTytC0yeBOvnZA4x4CFpdw/lCDPk0yrH9Ei5vVkXmOrExdTlT3qI7YaAzj1tUVlBd4S6LX1F7y6VLActvdHuDDuXZXzCDd/97420jrDfWZqJMlUK/EmCE5ParCeHIRIvmBxcEnGfFIsw8xQZl0HphxWOtJil8qsUWSdMyCiJYYQpMoMliO99X40AUc4/AlsyPyT5ddbKk08YrZ+rKDVHF7o29rh4vi5MmHkVgVQHKiKybWlHq+b71gIAUQk9wrJxD+dqt4igrmDSpIjfjwnd+l5UIn5fJSO5DYV4YT/4hwK7OKmuo7OFHD0WyY5YnkYEMtFgzemnRBdE8ulcT60DQpVgRMXFWHvhyCWy0L6sgj1QWDZlLpvsIvNfHsyhKFMG1frLnMt/nP0+YCcfg+v1JYeCKjeoJxB8DWcRBsjzItY0CGmzP8UYZiYKl/2u+2TgFS5r7NWH11bxoUzjKdaa1NLw+ieA8GlBFfCbfWe6YVB9ggUte4VtYFMZGxOjS2bAiYtfgTKFJv+XqORAwExG6+G2eDxIDyo80/OA9IG7Xv/jwQr7D6KDjDuULFcN/iTxuttoKrHeYz1hf5ZQlBdllwJHYx6fK2g8kha6r2JIQKocvsAXiiONqSfw== hello@world.com"
    }
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name                                   = "TestIPConfiguration"
      subnet_id                              = "${azurestack_subnet.test.id}"
      load_balancer_backend_address_pool_ids = ["${azurestack_lb_backend_address_pool.test.id}"]
    }
  }

  storage_profile_os_disk {
    name           = "osDiskProfile"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    os_type        = "linux"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  tags = {
    ThisIs = "a test"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_basicLinux_managedDisk(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name              = ""
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_basicWindows_managedDisk(rInt int, location string, vmSize string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "%[3]s"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  os_profile_windows_config {
    enable_automatic_upgrades = false
    provision_vm_agent        = true

    additional_unattend_config {
      pass         = "oobeSystem"
      component    = "Microsoft-Windows-Shell-Setup"
      setting_name = "AutoLogon"
      content      = "<AutoLogon><Username>myadmin</Username><Password><Value>Passwword1234</Value></Password><Enabled>true</Enabled><LogonCount>1</LogonCount></AutoLogon>"
    }

    additional_unattend_config {
      pass         = "oobeSystem"
      component    = "Microsoft-Windows-Shell-Setup"
      setting_name = "FirstLogonCommands"
      content      = "<FirstLogonCommands><SynchronousCommand><CommandLine>shutdown /r /t 0 /c \"initial reboot\"</CommandLine><Description>reboot</Description><Order>1</Order></SynchronousCommand></FirstLogonCommands>"
    }
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name              = ""
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_profile_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter-Server-Core"
    version   = "latest"
  }
}
`, rInt, location, vmSize)
}

func testAccAzureStackVirtualMachineScaleSet_basicLinux_managedDiskNoName(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetLoadBalancerTemplate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                          = "default"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "test"
  resource_group_name = "${azurestack_resource_group.test.name}"

  #location            = "${azurestack_resource_group.test.location}"
  loadbalancer_id = "${azurestack_lb.test.id}"
}

resource "azurestack_lb_nat_pool" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  name                           = "ssh"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  protocol                       = "Tcp"
  frontend_port_start            = 50000
  frontend_port_end              = 50119
  backend_port                   = 22
  frontend_ip_configuration_name = "default"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name                                   = "TestIPConfiguration"
      subnet_id                              = "${azurestack_subnet.test.id}"
      load_balancer_backend_address_pool_ids = ["${azurestack_lb_backend_address_pool.test.id}"]
      load_balancer_inbound_nat_rules_ids    = ["${azurestack_lb_nat_pool.test.id}"]
    }
  }

  storage_profile_os_disk {
    name           = "os-disk"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetOverProvisionTemplate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"
  overprovision       = false

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "os-disk"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetPriorityTemplate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"
  overprovision       = false
  priority            = "Low"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "os-disk"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetExtensionTemplate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"
  overprovision       = false

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/myadmin/.ssh/authorized_keys"
      key_data = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCsTcryUl51Q2VSEHqDRNmceUFo55ZtcIwxl2QITbN1RREti5ml/VTytC0yeBOvnZA4x4CFpdw/lCDPk0yrH9Ei5vVkXmOrExdTlT3qI7YaAzj1tUVlBd4S6LX1F7y6VLActvdHuDDuXZXzCDd/97420jrDfWZqJMlUK/EmCE5ParCeHIRIvmBxcEnGfFIsw8xQZl0HphxWOtJil8qsUWSdMyCiJYYQpMoMliO99X40AUc4/AlsyPyT5ddbKk08YrZ+rKDVHF7o29rh4vi5MmHkVgVQHKiKybWlHq+b71gIAUQk9wrJxD+dqt4igrmDSpIjfjwnd+l5UIn5fJSO5DYV4YT/4hwK7OKmuo7OFHD0WyY5YnkYEMtFgzemnRBdE8ulcT60DQpVgRMXFWHvhyCWy0L6sgj1QWDZlLpvsIvNfHsyhKFMG1frLnMt/nP0+YCcfg+v1JYeCKjeoJxB8DWcRBsjzItY0CGmzP8UYZiYKl/2u+2TgFS5r7NWH11bxoUzjKdaa1NLw+ieA8GlBFfCbfWe6YVB9ggUte4VtYFMZGxOjS2bAiYtfgTKFJv+XqORAwExG6+G2eDxIDyo80/OA9IG7Xv/jwQr7D6KDjDuULFcN/iTxuttoKrHeYz1hf5ZQlBdllwJHYx6fK2g8kha6r2JIQKocvsAXiiONqSfw== hello@world.com"
    }
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "os-disk"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  extension {
    name                       = "CustomScript"
    publisher                  = "Microsoft.Azure.Extensions"
    type                       = "CustomScript"
    type_handler_version       = "2.0"
    auto_upgrade_minor_version = true

    settings = <<SETTINGS
		{
			"commandToExecute": "echo $HOSTNAME"
		}
SETTINGS

    protected_settings = <<SETTINGS
		{
			"storageAccountName": "${azurestack_storage_account.test.name}",
			"storageAccountKey": "${azurestack_storage_account.test.primary_access_key}"
		}
SETTINGS
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetExtensionTemplateUpdated(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"
  overprovision       = false

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      path     = "/home/myadmin/.ssh/authorized_keys"
      key_data = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCsTcryUl51Q2VSEHqDRNmceUFo55ZtcIwxl2QITbN1RREti5ml/VTytC0yeBOvnZA4x4CFpdw/lCDPk0yrH9Ei5vVkXmOrExdTlT3qI7YaAzj1tUVlBd4S6LX1F7y6VLActvdHuDDuXZXzCDd/97420jrDfWZqJMlUK/EmCE5ParCeHIRIvmBxcEnGfFIsw8xQZl0HphxWOtJil8qsUWSdMyCiJYYQpMoMliO99X40AUc4/AlsyPyT5ddbKk08YrZ+rKDVHF7o29rh4vi5MmHkVgVQHKiKybWlHq+b71gIAUQk9wrJxD+dqt4igrmDSpIjfjwnd+l5UIn5fJSO5DYV4YT/4hwK7OKmuo7OFHD0WyY5YnkYEMtFgzemnRBdE8ulcT60DQpVgRMXFWHvhyCWy0L6sgj1QWDZlLpvsIvNfHsyhKFMG1frLnMt/nP0+YCcfg+v1JYeCKjeoJxB8DWcRBsjzItY0CGmzP8UYZiYKl/2u+2TgFS5r7NWH11bxoUzjKdaa1NLw+ieA8GlBFfCbfWe6YVB9ggUte4VtYFMZGxOjS2bAiYtfgTKFJv+XqORAwExG6+G2eDxIDyo80/OA9IG7Xv/jwQr7D6KDjDuULFcN/iTxuttoKrHeYz1hf5ZQlBdllwJHYx6fK2g8kha6r2JIQKocvsAXiiONqSfw== hello@world.com"
    }
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "os-disk"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  extension {
    name                       = "CustomScript"
    publisher                  = "Microsoft.Azure.Extensions"
    type                       = "CustomScript"
    type_handler_version       = "2.0"
    auto_upgrade_minor_version = true

    settings = <<SETTINGS
		{
			"commandToExecute": "echo $HOSTNAME",
			"timestamp": 12345679955
		}
SETTINGS

    protected_settings = <<SETTINGS
		{
			"storageAccountName": "${azurestack_storage_account.test.name}",
			"storageAccountKey": "${azurestack_storage_account.test.primary_access_key}"
		}
SETTINGS
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetMultipleExtensionsTemplate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%[1]d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"
  overprovision       = false

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name           = "os-disk"
    caching        = "ReadWrite"
    create_option  = "FromImage"
    vhd_containers = ["${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  extension {
    name                       = "CustomScript"
    publisher                  = "Microsoft.Azure.Extensions"
    type                       = "CustomScript"
    type_handler_version       = "2.0"
    auto_upgrade_minor_version = true

    settings = <<SETTINGS
		{
			"commandToExecute": "echo $HOSTNAME"
		}
SETTINGS

    protected_settings = <<SETTINGS
		{
			"storageAccountName": "${azurestack_storage_account.test.name}",
			"storageAccountKey": "${azurestack_storage_account.test.primary_access_key}"
		}
SETTINGS
  }

  extension {
    name                       = "Docker"
    publisher                  = "Microsoft.Azure.Extensions"
    type                       = "DockerExtension"
    type_handler_version       = "1.0"
    auto_upgrade_minor_version = true
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_osDiskTypeConflict(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  storage_profile_os_disk {
    name              = ""
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    vhd_containers    = ["should_cause_conflict"]
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSetLoadBalancerTemplateManagedDataDisks(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                          = "default"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "test"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_A0"
    tier     = "Standard"
    capacity = 1
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile"
    primary = true

    ip_configuration {
      name                                   = "TestIPConfiguration"
      subnet_id                              = "${azurestack_subnet.test.id}"
      load_balancer_backend_address_pool_ids = ["${azurestack_lb_backend_address_pool.test.id}"]
    }
  }

  storage_profile_os_disk {
    name              = ""
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_profile_data_disk {
    lun               = 0
    caching           = "ReadWrite"
    create_option     = "Empty"
    disk_size_gb      = 10
    managed_disk_type = "Standard_LRS"
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}

func testAccAzureStackVirtualMachineScaleSet_planManagedDisk(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%[1]d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_virtual_machine_scale_set" "test" {
  name                = "acctvmss-%[1]d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  upgrade_policy_mode = "Manual"

  sku {
    name     = "Standard_D1_v2"
    tier     = "Standard"
    capacity = 2
  }

  os_profile {
    computer_name_prefix = "testvm-%[1]d"
    admin_username       = "myadmin"
    admin_password       = "Passwword1234"
  }

  network_profile {
    name    = "TestNetworkProfile-%[1]d"
    primary = true

    ip_configuration {
      name      = "TestIPConfiguration"
      subnet_id = "${azurestack_subnet.test.id}"
    }
  }

  plan {
    name      = "basic-byol"
    publisher = "kemptech"
    product   = "vlm-azure"
  }

  storage_profile_os_disk {
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_profile_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, rInt, location)
}
