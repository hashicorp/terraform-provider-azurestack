package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackNetworkInterface_basic(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_basic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_disappears(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_basic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					testCheckAzureStackNetworkInterfaceDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_setNetworkSecurityGroupId(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	location := testLocation()
	config := testAccAzureStackNetworkInterface_basic(rInt, location)
	updatedConfig := testAccAzureStackNetworkInterface_basicWithNetworkSecurityGroup(rInt, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "network_security_group_id"),
				),
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_removeNetworkSecurityGroupId(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	location := testLocation()
	config := testAccAzureStackNetworkInterface_basicWithNetworkSecurityGroup(rInt, location)
	updatedConfig := testAccAzureStackNetworkInterface_basic(rInt, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_security_group_id", ""),
				),
			},
		},
	})
}

// Not supported by the account
// Microsoft.Network/AllowMultipleIpConfigurationsPerNic
func TestAccAzureStackNetworkInterface_multipleSubnets(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	location := testLocation()
	config := testAccAzureStackNetworkInterface_multipleSubnets(rInt, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.#", "2"),
				),
			},
		},
	})
}

// Code="SubscriptionNotRegisteredForFeature" Message="Subscription XXXX-XXXX-XXXX-XXXX
// isnot registered for feature Microsoft.Network/AllowMultipleIpConfigurationsPerNic
// required to carry out the requested operation.

func TestAccAzureStackNetworkInterface_multipleSubnetsPrimary(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	location := testLocation()
	config := testAccAzureStackNetworkInterface_multipleSubnets(rInt, location)
	updatedConfig := testAccAzureStackNetworkInterface_multipleSubnetsUpdatedPrimary(rInt, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.0.primary", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.0.name", "testconfiguration1"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.1.primary", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.1.name", "testconfiguration2"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.0.primary", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.0.name", "testconfiguration2"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.1.primary", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.1.name", "testconfiguration1"),
				),
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_enableIPForwarding(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_ipForwarding(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_ip_forwarding", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_accelerated_networking"},
			},
		},
	})
}

// enableAcceleratedNetworking not in the profile
func TestAccAzureStackNetworkInterface_enableAcceleratedNetworking(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_acceleratedNetworking(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_accelerated_networking"},
			},
		},
	})
}

// Will skip this until I add LB

func TestAccAzureStackNetworkInterface_multipleLoadBalancers(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_multipleLoadBalancers(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					testCheckAzureStackNetworkInterfaceExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_applicationGateway(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_applicationGatewayBackendPool(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists("azurestack_network_interface.test"),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.0.application_gateway_backend_address_pools_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackNetworkInterface_withTags(t *testing.T) {
	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_withTags(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost_center", "MSFT"),
				),
			},
			{
				Config: testAccAzureStackNetworkInterface_withTagsUpdate(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "staging"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// missing public ip
func TestAccAzureStackNetworkInterface_bug7986(t *testing.T) {

	t.Skip()

	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_bug7986(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists("azurestack_network_interface.test1"),
					testCheckAzureStackNetworkInterfaceExists("azurestack_network_interface.test2"),
				),
			},
		},
	})
}

// app security group is not supported by the profile
func TestAccAzureStackNetworkInterface_applicationSecurityGroups(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_applicationSecurityGroup(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_configuration.0.application_security_group_ids.#", "1"),
				),
			},
		},
	})
}

// Not enough configuration to run this.
func TestAccAzureStackNetworkInterface_internalFQDN(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_interface.test"
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkInterface_internalFQDN(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkInterfaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "internal_fqdn", fmt.Sprintf("acctestnic-%d.example.com", rInt)),
				),
			},
		},
	})
}

func testCheckAzureStackNetworkInterfaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for availability set: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).ifaceClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, name, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Network Interface %q (resource group: %q) does not exist", name, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on ifaceClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackNetworkInterfaceDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for availability set: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).ifaceClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Error deleting Network Interface %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for the deletion of Network Interface %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackNetworkInterfaceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).ifaceClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_network_interface" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Network Interface still exists:\n%#v", resp.InterfacePropertiesFormat)
	}

	return nil
}

func testAccAzureStackNetworkInterface_basic(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_basicWithNetworkSecurityGroup(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_network_interface" "test" {
  name                      = "acctestni-%d"
  location                  = "${azurestack_resource_group.test.location}"
  resource_group_name       = "${azurestack_resource_group.test.name}"
  network_security_group_id = "${azurestack_network_security_group.test.id}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackNetworkInterface_multipleSubnets(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
    primary                       = true
  }

  ip_configuration {
    name                          = "testconfiguration2"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_multipleSubnetsUpdatedPrimary(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration2"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
    primary                       = true
  }

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_ipForwarding(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                 = "acctestni-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  enable_ip_forwarding = true

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_acceleratedNetworking(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                          = "acctestni-%d"
  location                      = "${azurestack_resource_group.test.location}"
  resource_group_name           = "${azurestack_resource_group.test.name}"
  enable_ip_forwarding          = false
  enable_accelerated_networking = true

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_withTags(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_withTagsUpdate(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestni-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackNetworkInterface_multipleLoadBalancers(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_public_ip" "testext" {
  name                         = "acctestpip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "testext" {
  name                = "acctestlb-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "publicipext"
    public_ip_address_id = "${azurestack_public_ip.testext.id}"
  }
}

resource "azurestack_lb_backend_address_pool" "testext" {
  name                = "testbackendpoolext"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.testext.id}"
}

resource "azurestack_lb_nat_rule" "testext" {
  name                           = "testnatruleext"
  location                       = "${azurestack_resource_group.test.location}"
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.testext.id}"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3390
  frontend_ip_configuration_name = "publicipext"
}

resource "azurestack_public_ip" "testint" {
  name                         = "testpublicipint"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "testint" {
  name                = "testlbint"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                          = "publicipint"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_lb_backend_address_pool" "testint" {
  name                = "testbackendpoolint"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.testint.id}"
}

resource "azurestack_lb_nat_rule" "testint" {
  name                           = "testnatruleint"
  location                       = "${azurestack_resource_group.test.location}"
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.testint.id}"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3391
  frontend_ip_configuration_name = "publicipint"
}

resource "azurestack_network_interface" "test1" {
  name                 = "acctestnic1-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  enable_ip_forwarding = true

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"

    load_balancer_backend_address_pools_ids = [
      "${azurestack_lb_backend_address_pool.testext.id}",
      "${azurestack_lb_backend_address_pool.testint.id}",
    ]
  }
}

resource "azurestack_network_interface" "test2" {
  name                 = "acctestnic2-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  enable_ip_forwarding = true

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"

    load_balancer_inbound_nat_rules_ids = [
      "${azurestack_lb_nat_rule.testext.id}",
      "${azurestack_lb_nat_rule.testint.id}",
    ]
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackNetworkInterface_applicationGatewayBackendPool(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest-vnet-%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  address_space       = ["10.254.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
}

resource "azurestack_subnet" "gateway" {
  name                 = "subnet-gateway-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.254.0.0/24"
}

resource "azurestack_subnet" "test" {
  name                 = "subnet-%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.254.1.0/24"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctest-pubip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "dynamic"
}

resource "azurestack_application_gateway" "test" {
  name                = "acctestgw-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  sku {
    name     = "Standard_Medium"
    tier     = "Standard"
    capacity = 1
  }

  gateway_ip_configuration {
    name      = "gw-ip-config1"
    subnet_id = "${azurestack_subnet.gateway.id}"
  }

  frontend_port {
    name = "port-8080"
    port = 8080
  }

  frontend_ip_configuration {
    name                 = "ip-config-public"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }

  backend_address_pool {
    name = "pool-1"
  }

  backend_http_settings {
    name                  = "backend-http-1"
    port                  = 8080
    protocol              = "Http"
    cookie_based_affinity = "Enabled"
    request_timeout       = 30
  }

  http_listener {
    name                           = "listener-1"
    frontend_ip_configuration_name = "ip-config-public"
    frontend_port_name             = "port-8080"
    protocol                       = "Http"
  }

  request_routing_rule {
    name                       = "rule-basic-1"
    rule_type                  = "Basic"
    http_listener_name         = "listener-1"
    backend_address_pool_name  = "pool-1"
    backend_http_settings_name = "backend-http-1"
  }

  tags = {
    environment = "tf01"
  }
}

resource "azurestack_network_interface" "test" {
  name                 = "acctestnic-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  enable_ip_forwarding = true

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"

    application_gateway_backend_address_pools_ids = [
      "${azurestack_application_gateway.test.backend_address_pool.0.id}",
    ]
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackNetworkInterface_bug7986(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctest-%d-nsg"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "Production"
  }
}

resource "azurestack_network_security_rule" "test1" {
  name                        = "test1"
  priority                    = 101
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurestack_resource_group.test.name}"
  network_security_group_name = "${azurestack_network_security_group.test.name}"
}

resource "azurestack_network_security_rule" "test2" {
  name                        = "test2"
  priority                    = 102
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurestack_resource_group.test.name}"
  network_security_group_name = "${azurestack_network_security_group.test.name}"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctest-%d-pip"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "Dynamic"

  tags = {
    environment = "Production"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest-%d-vn"
  address_space       = ["10.0.0.0/16"]
  resource_group_name = "${azurestack_resource_group.test.name}"
  location            = "${azurestack_resource_group.test.location}"
}

resource "azurestack_subnet" "test" {
  name                 = "first"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test1" {
  name                = "acctest-%d-nic1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }

  tags = {
    environment = "staging"
  }
}

resource "azurestack_network_interface" "test2" {
  name                = "acctest-%d-nic2"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackNetworkInterface_applicationSecurityGroup(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_application_security_group" "test" {
  name                = "acctest-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestnic-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                           = "testconfiguration1"
    subnet_id                      = "${azurestack_subnet.test.id}"
    private_ip_address_allocation  = "dynamic"
    application_security_group_ids = ["${azurestack_application_security_group.test.id}"]
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackNetworkInterface_internalFQDN(rInt int, location string) string {
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
  name                 = "testsubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctestnic-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  internal_fqdn       = "acctestnic-%d.example.com"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}
`, rInt, location, rInt, rInt, rInt)
}
