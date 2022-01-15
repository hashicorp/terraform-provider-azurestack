package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAzureStackNetworkSecurityGroup_basic(t *testing.T) {
	dataSourceName := "data.azurestack_network_security_group.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccDataSourceAzureStackNetworkSecurityGroupBasic(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "location"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureStackNetworkSecurityGroup_rules(t *testing.T) {
	dataSourceName := "data.azurestack_network_security_group.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccDataSourceAzureStackNetworkSecurityGroupWithRules(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "location"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.name", "test123"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.priority", "100"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.direction", "Inbound"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.access", "Allow"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.protocol", "Tcp"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.source_port_range", "*"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.destination_port_range", "*"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.source_address_prefix", "*"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.0.destination_address_prefix", "*"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureStackNetworkSecurityGroup_tags(t *testing.T) {
	dataSourceName := "data.azurestack_network_security_group.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccDataSourceAzureStackNetworkSecurityGroupTags(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackNetworkSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "location"),
					resource.TestCheckResourceAttr(dataSourceName, "security_rule.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.environment", "staging"),
				),
			},
		},
	})
}

func testAccDataSourceAzureStackNetworkSecurityGroupBasic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

data "azurestack_network_security_group" "test" {
  name                = "${azurestack_network_security_group.test.name}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location, rInt)
}

func testAccDataSourceAzureStackNetworkSecurityGroupWithRules(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
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
}

data "azurestack_network_security_group" "test" {
  name                = "${azurestack_network_security_group.test.name}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location, rInt)
}

func testAccDataSourceAzureStackNetworkSecurityGroupTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "staging"
  }
}

data "azurestack_network_security_group" "test" {
  name                = "${azurestack_network_security_group.test.name}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location, rInt)
}
