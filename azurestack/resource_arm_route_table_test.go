package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackRouteTable_basic(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackRouteTable_basic(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
				),
			},
			{
				ResourceName:      "azurestack_route_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackRouteTable_complete(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackRouteTable_complete(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists("azurestack_route_table.test"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
			{
				ResourceName:      "azurestack_route_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackRouteTable_update(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackRouteTable_basic(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists("azurestack_route_table.test"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
				),
			},
			{
				Config: testAccAzureStackRouteTable_complete(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists("azurestack_route_table.test"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
		},
	})
}

func TestAccAzureStackRouteTable_singleRoute(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackRouteTable_singleRoute(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists("azurestack_route_table.test"),
				),
			},
			{
				ResourceName:      "azurestack_route_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackRouteTable_removeRoute(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()
	config := testAccAzureStackRouteTable_singleRoute(ri, testLocation())
	updatedConfig := testAccAzureStackRouteTable_singleRouteRemoved(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
				),
			},
		},
	})
}

func TestAccAzureStackRouteTable_disappears(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()
	config := testAccAzureStackRouteTable_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					testCheckAzureStackRouteTableDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackRouteTable_withTags(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()
	preConfig := testAccAzureStackRouteTable_withTags(ri, testLocation())
	postConfig := testAccAzureStackRouteTable_withTagsUpdate(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost_center", "MSFT"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "staging"),
				),
			},
		},
	})
}

func TestAccAzureStackRouteTable_multipleRoutes(t *testing.T) {
	resourceName := "azurestack_route_table.test"
	ri := acctest.RandInt()
	preConfig := testAccAzureStackRouteTable_singleRoute(ri, testLocation())
	postConfig := testAccAzureStackRouteTable_multipleRoutes(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "route.0.name", "route1"),
					resource.TestCheckResourceAttr(resourceName, "route.0.address_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "route.0.next_hop_type", "VnetLocal"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "route.0.name", "route1"),
					resource.TestCheckResourceAttr(resourceName, "route.0.address_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "route.0.next_hop_type", "VnetLocal"),
					resource.TestCheckResourceAttr(resourceName, "route.1.name", "route2"),
					resource.TestCheckResourceAttr(resourceName, "route.1.address_prefix", "10.2.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "route.1.next_hop_type", "VnetLocal"),
				),
			},
			{
				ResourceName:      "azurestack_route_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackRouteTable_withTagsSubnet(t *testing.T) {
	ri := acctest.RandInt()
	configSetup := testAccAzureStackRouteTable_withTagsSubnet(ri, testLocation())
	configTest := testAccAzureStackRouteTable_withAddTagsSubnet(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: configSetup,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists("azurestack_route_table.test"),
					testCheckAzureStackSubnetExists("azurestack_subnet.subnet1"),
					resource.TestCheckResourceAttrSet("azurestack_subnet.subnet1", "route_table_id"),
				),
			},
			{
				Config: configTest,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists("azurestack_route_table.test"),
					testCheckAzureStackSubnetExists("azurestack_subnet.subnet1"),
					resource.TestCheckResourceAttrSet("azurestack_subnet.subnet1", "route_table_id"),
				),
			},
		},
	})
}

func testCheckAzureStackRouteTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for route table: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).routeTablesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, name, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Route Table %q (resource group: %q) does not exist", name, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on routeTablesClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackRouteTableDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for route table: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).routeTablesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, name)
		if err != nil {
			if !response.WasNotFound(future.Response()) {
				return fmt.Errorf("Error deleting Route Table %q (Resource Group %q): %+v", name, resourceGroup, err)
			}
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for deletion of Route Table %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackRouteTableDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).routeTablesClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_route_table" {
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

		return fmt.Errorf("Route Table still exists:\n%#v", resp.RouteTablePropertiesFormat)
	}

	return nil
}

func testAccAzureStackRouteTable_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_complete(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "acctestRoute"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_singleRoute(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_singleRouteRemoved(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route = []
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_multipleRoutes(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }

  route {
    name           = "route2"
    address_prefix = "10.2.0.0/16"
    next_hop_type  = "vnetlocal"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_withTagsUpdate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackRouteTable_withTagsSubnet(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  address_space       = ["10.0.0.0/16"]

  tags = {
    environment = "staging"
  }
}

resource "azurestack_subnet" "subnet1" {
  name                 = "subnet1"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
  route_table_id       = "${azurestack_route_table.test.id}"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackRouteTable_withAddTagsSubnet(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"

  tags = {
    environment = "staging"
    cloud       = "Azure"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  address_space       = ["10.0.0.0/16"]

  tags = {
    environment = "staging"
    cloud       = "Azure"
  }
}

resource "azurestack_subnet" "subnet1" {
  name                 = "subnet1"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
  route_table_id       = "${azurestack_route_table.test.id}"
}

resource "azurestack_route_table" "test" {
  name                = "acctestrt%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }

  tags = {
    environment = "staging"
    cloud       = "Azure"
  }
}
`, rInt, location, rInt, rInt)
}
