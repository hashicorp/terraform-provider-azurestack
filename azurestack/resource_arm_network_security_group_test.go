package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/utils"
)

func TestAccAzureStackNetworkSecurityGroup_basic(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_basic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
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

func TestAccAzureStackNetworkSecurityGroup_singleRule(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_singleRule(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
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

func TestAccAzureStackNetworkSecurityGroup_update(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_singleRule(rInt, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
				),
			},
			{
				Config: testAccAzureStackNetworkSecurityGroup_basic(rInt, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackNetworkSecurityGroup_disappears(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_basic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
					testCheckAzureStackNetworkSecurityGroupDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackNetworkSecurityGroup_withTags(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_withTags(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost_center", "MSFT"),
				),
			},

			{
				Config: testAccAzureStackNetworkSecurityGroup_withTagsUpdate(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
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

func TestAccAzureStackNetworkSecurityGroup_addingExtraRules(t *testing.T) {
	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_singleRule(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_rule.#", "1"),
				),
			},

			{
				Config: testAccAzureStackNetworkSecurityGroup_anotherRule(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_rule.#", "2"),
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

// Not supported by the profile by now
func TestAccAzureStackNetworkSecurityGroup_augmented(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_augmented(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_rule.#", "1"),
				),
			},
		},
	})
}

// application security group not in scope
func TestAccAzureStackNetworkSecurityGroup_applicationSecurityGroup(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_network_security_group.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityGroup_applicationSecurityGroup(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_rule.#", "1"),
				),
			},
		},
	})
}

func testCheckAzureStackNetworkSecurityGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		sgName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for network security group: %q", sgName)
		}

		client := testAccProvider.Meta().(*ArmClient).secGroupClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, resourceGroup, sgName, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Network Security Group %q (resource group: %q) does not exist", name, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on secGroupClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackNetworkSecurityGroupDisappears(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		sgName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for network security group: %q", sgName)
		}

		client := testAccProvider.Meta().(*ArmClient).secGroupClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		future, err := client.Delete(ctx, resourceGroup, sgName)
		if err != nil {
			return fmt.Errorf("Error deleting NSG %q (Resource Group %q): %+v", sgName, resourceGroup, err)
		}
		err = future.WaitForCompletion(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error deleting NSG %q (Resource Group %q): %+v", sgName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackNetworkSecurityGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).secGroupClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_network_security_group" {
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

		return fmt.Errorf("Network Security Group still exists:\n%#v", resp.SecurityGroupPropertiesFormat)
	}

	return nil
}

func testAccAzureStackNetworkSecurityGroup_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityGroup_singleRule(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "TCP"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

`, rInt, location)
}

func testAccAzureStackNetworkSecurityGroup_anotherRule(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "testDeny"
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Deny"
    protocol                   = "Udp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityGroup_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityGroup_withTagsUpdate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags {
    environment = "staging"
  }
}

`, rInt, location)
}

func testAccAzureStackNetworkSecurityGroup_augmented(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                         = "test123"
    priority                     = 100
    direction                    = "Inbound"
    access                       = "Allow"
    protocol                     = "Tcp"
    source_port_ranges           = [ "10000-40000" ]
    destination_port_ranges      = [ "80", "443", "8080", "8190" ]
    source_address_prefixes      = [ "10.0.0.0/8", "192.168.0.0/16" ]
    destination_address_prefixes = [ "172.16.0.0/20", "8.8.8.8" ]
  }
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityGroup_applicationSecurityGroup(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_application_security_group" "first" {
  name                = "acctest-first%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_application_security_group" "second" {
  name                = "acctest-second%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                                       = "test123"
    priority                                   = 100
    direction                                  = "Inbound"
    access                                     = "Allow"
    protocol                                   = "Tcp"
    source_application_security_group_ids      = ["${azurestack_application_security_group.first.id}"]
    destination_application_security_group_ids = ["${azurestack_application_security_group.second.id}"]
    source_port_ranges                         = [ "10000-40000" ]
    destination_port_ranges                    = [ "80", "443", "8080", "8190" ]
  }
}
`, rInt, location, rInt, rInt, rInt)
}
