package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/utils"
)

func TestAccAzureStackNetworkSecurityRule_basic(t *testing.T) {
	resourceName := "azurestack_network_security_rule.test"
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityRule_basic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityRuleExists(resourceName),
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

func TestAccAzureStackNetworkSecurityRule_disappears(t *testing.T) {
	resourceGroup := "azurestack_network_security_rule.test"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityRule_basic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityRuleExists(resourceGroup),
					testCheckAzureStackNetworkSecurityRuleDisappears(resourceGroup),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackNetworkSecurityRule_addingRules(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityRule_updateBasic(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityRuleExists("azurestack_network_security_rule.test1"),
				),
			},

			{
				Config: testAccAzureStackNetworkSecurityRule_updateExtraRule(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityRuleExists("azurestack_network_security_rule.test2"),
				),
			},
		},
	})
}

func TestAccAzureStackNetworkSecurityRule_augmented(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityRule_augmented(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityRuleExists("azurestack_network_security_rule.test1"),
				),
			},
		},
	})
}

// azurestack_application_security_group not in scope, skipping
func TestAccAzureStackNetworkSecurityRule_applicationSecurityGroups(t *testing.T) {

	t.Skip()

	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackNetworkSecurityRule_applicationSecurityGroups(rInt, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackNetworkSecurityRuleExists("azurestack_network_security_rule.test1"),
				),
			},
		},
	})
}

func testCheckAzureStackNetworkSecurityRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		sgName := rs.Primary.Attributes["network_security_group_name"]
		sgrName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for network security rule: %q", sgName)
		}

		client := testAccProvider.Meta().(*ArmClient).secRuleClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, sgName, sgrName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Network Security Rule %q (resource group: %q) (network security group: %q) does not exist", sgrName, sgName, resourceGroup)
			}
			return fmt.Errorf("Error retrieving Network Security Rule %q (NSG %q / Resource Group %q): %+v", sgrName, sgName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackNetworkSecurityRuleDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		sgName := rs.Primary.Attributes["network_security_group_name"]
		sgrName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for network security rule: %s", sgName)
		}

		client := testAccProvider.Meta().(*ArmClient).secRuleClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		future, err := client.Delete(ctx, resourceGroup, sgName, sgrName)
		if err != nil {
			if !response.WasNotFound(future.Response()) {
				return fmt.Errorf("Error deleting Network Security Rule %q (NSG %q / Resource Group %q): %+v", sgrName, sgName, resourceGroup, err)
			}
		}

		err = future.WaitForCompletion(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for the deletion of Network Security Rule %q (NSG %q / Resource Group %q): %+v", sgrName, sgName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackNetworkSecurityRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).secRuleClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {

		if rs.Type != "azurestack_network_security_rule" {
			continue
		}

		sgName := rs.Primary.Attributes["network_security_group_name"]
		sgrName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, sgName, sgrName)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Network Security Rule still exists:\n%#v", resp.SecurityRulePropertiesFormat)
		}
	}

	return nil
}

func testAccAzureStackNetworkSecurityRule_basic(rInt int, location string) string {
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

resource "azurestack_network_security_rule" "test" {
  name                        = "test123"
  priority                    = 100
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
`, rInt, location)
}

func testAccAzureStackNetworkSecurityRule_updateBasic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test1" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test1" {
  name                = "acceptanceTestSecurityGroup2"
  location            = "${azurestack_resource_group.test1.location}"
  resource_group_name = "${azurestack_resource_group.test1.name}"
}

resource "azurestack_network_security_rule" "test1" {
  name                        = "test123"
  priority                    = 100
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurestack_resource_group.test1.name}"
  network_security_group_name = "${azurestack_network_security_group.test1.name}"
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityRule_updateExtraRule(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test1" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test1" {
  name                = "acceptanceTestSecurityGroup2"
  location            = "${azurestack_resource_group.test1.location}"
  resource_group_name = "${azurestack_resource_group.test1.name}"
}

resource "azurestack_network_security_rule" "test1" {
  name                        = "test123"
  priority                    = 100
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurestack_resource_group.test1.name}"
  network_security_group_name = "${azurestack_network_security_group.test1.name}"
}

resource "azurestack_network_security_rule" "test2" {
  name                        = "testing456"
  priority                    = 101
  direction                   = "Inbound"
  access                      = "Deny"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = "${azurestack_resource_group.test1.name}"
  network_security_group_name = "${azurestack_network_security_group.test1.name}"
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityRule_augmented(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test1" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test1" {
  name                = "acceptanceTestSecurityGroup2"
  location            = "${azurestack_resource_group.test1.location}"
  resource_group_name = "${azurestack_resource_group.test1.name}"
}

resource "azurestack_network_security_rule" "test1" {
  name                         = "test123"
  priority                     = 100
  direction                    = "Outbound"
  access                       = "Allow"
  protocol                     = "Tcp"
	source_port_range            = "*"
	destination_port_range       = "*"
  source_address_prefix        = "10.0.0.0/8"
  destination_address_prefix   = "172.16.0.0/20"
  resource_group_name          = "${azurestack_resource_group.test1.name}"
  network_security_group_name  = "${azurestack_network_security_group.test1.name}"
}
`, rInt, location)
}

func testAccAzureStackNetworkSecurityRule_applicationSecurityGroups(rInt int, location string) string {
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
}

resource "azurestack_network_security_rule" "test1" {
  name                                       = "test123"
  resource_group_name                        = "${azurestack_resource_group.test.name}"
  network_security_group_name                = "${azurestack_network_security_group.test.name}"
  priority                                   = 100
  direction                                  = "Outbound"
  access                                     = "Allow"
  protocol                                   = "Tcp"
  source_application_security_group_ids      = ["${azurestack_application_security_group.first.id}"]
  destination_application_security_group_ids = ["${azurestack_application_security_group.second.id}"]
  source_port_ranges                         = [ "10000-40000" ]
  destination_port_ranges                    = [ "80", "443", "8080", "8190" ]
}
`, rInt, location, rInt, rInt, rInt)
}
