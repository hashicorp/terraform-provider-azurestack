package azurestack

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackSubnet_basic(t *testing.T) {

	resourceName := "azurestack_subnet.test"
	ri := acctest.RandInt()
	config := testAccAzureStackSubnet_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists(resourceName),
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

func TestAccAzureStackSubnet_routeTableUpdate(t *testing.T) {

	resourceName := "azurestack_subnet.test"
	ri := acctest.RandInt()
	location := testLocation()
	initConfig := testAccAzureStackSubnet_routeTable(ri, location)
	updatedConfig := testAccAzureStackSubnet_updatedRouteTable(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: initConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists(resourceName),
				),
			},

			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetRouteTableExists("azurestack_subnet.test", fmt.Sprintf("acctest-%d", ri)),
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

func TestAccAzureStackSubnet_routeTableRemove(t *testing.T) {

	resourceName := "azurestack_subnet.test"
	ri := acctest.RandInt()
	location := testLocation()
	initConfig := testAccAzureStackSubnet_routeTable(ri, location)
	updatedConfig := testAccAzureStackSubnet_routeTableUnlinked(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: initConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
				),
			},

			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route_table_id", ""),
				),
			},
		},
	})
}

func TestAccAzureStackSubnet_removeNetworkSecurityGroup(t *testing.T) {
	resourceName := "azurestack_subnet.test"
	ri := acctest.RandInt()
	location := testLocation()
	initConfig := testAccAzureStackSubnet_networkSecurityGroup(ri, location)
	updatedConfig := testAccAzureStackSubnet_networkSecurityGroupDetached(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: initConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "network_security_group_id"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_security_group_id", ""),
				),
			},
		},
	})
}

func TestAccAzureStackSubnet_bug7986(t *testing.T) {

	ri := acctest.RandInt()
	initConfig := testAccAzureStackSubnet_bug7986(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: initConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists("azurestack_subnet.first"),
					testCheckAzureStackSubnetExists("azurestack_subnet.second"),
				),
			},
		},
	})
}

func TestAccAzureStackSubnet_bug15204(t *testing.T) {

	ri := acctest.RandInt()
	initConfig := testAccAzureStackSubnet_bug15204(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: initConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists("azurestack_subnet.test"),
				),
			},
		},
	})
}

func TestAccAzureStackSubnet_disappears(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackSubnet_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists("azurestack_subnet.test"),
					testCheckAzureStackSubnetDisappears("azurestack_subnet.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureStackSubnetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		log.Printf("[INFO] Checking Subnet addition.")

		name := rs.Primary.Attributes["name"]
		vnetName := rs.Primary.Attributes["virtual_network_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for subnet: %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).subnetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, vnetName, name, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Subnet %q (resource group: %q) does not exist", name, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackSubnetRouteTableExists(subnetName string, routeTableId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[subnetName]
		if !ok {
			return fmt.Errorf("Not found: %s", subnetName)
		}

		log.Printf("[INFO] Checking Subnet update.")

		name := rs.Primary.Attributes["name"]
		vnetName := rs.Primary.Attributes["virtual_network_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for subnet: %s", name)
		}

		networksClient := testAccProvider.Meta().(*ArmClient).vnetClient
		subnetsClient := testAccProvider.Meta().(*ArmClient).subnetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		vnetResp, vnetErr := networksClient.Get(ctx, resourceGroup, vnetName, "")
		if vnetErr != nil {
			return fmt.Errorf("Bad: Get on vnetClient: %+v", vnetErr)
		}

		if vnetResp.Subnets == nil {
			return fmt.Errorf("Bad: Vnet %q (resource group: %q) does not have subnets after update", vnetName, resourceGroup)
		}

		resp, err := subnetsClient.Get(ctx, resourceGroup, vnetName, name, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Subnet %q (resource group: %q) does not exist", subnetName, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on subnetClient: %+v", err)
		}

		if resp.RouteTable == nil {
			return fmt.Errorf("Bad: Subnet %q (resource group: %q) does not contain route tables after update", subnetName, resourceGroup)
		}

		if !strings.Contains(*resp.RouteTable.ID, routeTableId) {
			return fmt.Errorf("Bad: Subnet %q (resource group: %q) does not have route table %q", subnetName, resourceGroup, routeTableId)
		}

		return nil
	}
}

func testCheckAzureStackSubnetDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		vnetName := rs.Primary.Attributes["virtual_network_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for subnet: %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).subnetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		future, err := client.Delete(ctx, resourceGroup, vnetName, name)
		if err != nil {
			if !response.WasNotFound(future.Response()) {
				return fmt.Errorf("Error deleting Subnet %q (Network %q / Resource Group %q): %+v", name, vnetName, resourceGroup, err)
			}
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for completion of Subnet %q (Network %q / Resource Group %q): %+v", name, vnetName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackSubnetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).subnetClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_subnet" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		vnetName := rs.Primary.Attributes["virtual_network_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, vnetName, name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Subnet still exists:\n%#v", resp.SubnetPropertiesFormat)
			}
			return nil
		}
	}

	return nil
}

// Not supported for 2017-03-09 profile
func TestAccAzureStackSubnet_serviceEndpoints(t *testing.T) {

	t.Skip()

	ri := acctest.RandInt()
	config := testAccAzureStackSubnet_serviceEndpoints(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackSubnetExists("azurestack_subnet.test"),
				),
			},
		},
	})
}

func testAccAzureStackSubnet_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackSubnet_routeTable(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
  route_table_id       = "${azurestack_route_table.test.id}"
}

resource "azurestack_route_table" "test" {
  name                = "acctest-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name                   = "acctest-%d"
    address_prefix         = "10.100.0.0/14"
    next_hop_type          = "VirtualAppliance"
    next_hop_in_ip_address = "10.10.1.1"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_routeTableUnlinked(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_route_table" "test" {
  name                = "acctest-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name                   = "acctest-%d"
    address_prefix         = "10.100.0.0/14"
    next_hop_type          = "VirtualAppliance"
    next_hop_in_ip_address = "10.10.1.1"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_updatedRouteTable(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"

  tags = {
    environment = "Testing"
  }
}

resource "azurestack_network_security_group" "test_secgroup" {
  name                = "acctest-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  security_rule {
    name                       = "acctest-%d"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = {
    environment = "Testing"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "Testing"
  }
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
  route_table_id       = "${azurestack_route_table.test.id}"
}

resource "azurestack_route_table" "test" {
  name                = "acctest-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name                   = "acctest-%d"
    address_prefix         = "10.100.0.0/14"
    next_hop_type          = "VirtualAppliance"
    next_hop_in_ip_address = "10.10.1.1"
  }

  tags = {
    environment = "Testing"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_networkSecurityGroup(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg%d"
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

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                      = "acctest%d-private"
  resource_group_name       = "${azurestack_resource_group.test.name}"
  virtual_network_name      = "${azurestack_virtual_network.test.name}"
  address_prefix            = "10.0.0.0/24"
  network_security_group_id = "${azurestack_network_security_group.test.id}"
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_networkSecurityGroupDetached(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg%d"
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

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctest%d-private"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.0.0/24"
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_bug7986(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest%d-rg"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_route_table" "first" {
  name                = "acctest%d-private-1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "acctest%d-private-1"
    address_prefix = "0.0.0.0/0"
    next_hop_type  = "None"
  }
}

resource "azurestack_subnet" "first" {
  name                 = "acctest%d-private-1"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.0.0/24"
  route_table_id       = "${azurestack_route_table.first.id}"
}

resource "azurestack_route_table" "second" {
  name                = "acctest%d-private-2"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "acctest%d-private-2"
    address_prefix = "0.0.0.0/0"
    next_hop_type  = "None"
  }
}

resource "azurestack_subnet" "second" {
  name                 = "acctest%d-private-2"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
  route_table_id       = "${azurestack_route_table.second.id}"
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_bug15204(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctest-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
  address_space       = ["10.85.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_network_security_group" "test" {
  name                = "acctestnsg-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                      = "acctestsubnet-%d"
  resource_group_name       = "${azurestack_resource_group.test.name}"
  virtual_network_name      = "${azurestack_virtual_network.test.name}"
  address_prefix            = "10.85.9.0/24"
  route_table_id            = "${azurestack_route_table.test.id}"
  network_security_group_id = "${azurestack_network_security_group.test.id}"
}
`, rInt, location, rInt, rInt, rInt, rInt)
}

func testAccAzureStackSubnet_serviceEndpoints(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctestsubnet%d"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
  service_endpoints    = ["Microsoft.Sql", "Microsoft.Storage"]
}
`, rInt, location, rInt, rInt)
}
