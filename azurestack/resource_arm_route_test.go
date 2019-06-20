package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackRoute_basic(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackRoute_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteExists("azurestack_route.test"),
				),
			},
			{
				ResourceName:      "azurestack_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackRoute_disappears(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackRoute_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteExists("azurestack_route.test"),
					testCheckAzureStackRouteDisappears("azurestack_route.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackRoute_multipleRoutes(t *testing.T) {
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackRoute_basic(ri, location)
	postConfig := testAccAzureStackRoute_multipleRoutes(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteExists("azurestack_route.test"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteExists("azurestack_route.test1"),
				),
			},
		},
	})
}

func testCheckAzureStackRouteExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		name := rs.Primary.Attributes["name"]
		rtName := rs.Primary.Attributes["route_table_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for route: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).routesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, rtName, name)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Route %q (resource group: %q) does not exist", name, resourceGroup)
			}
			return fmt.Errorf("Bad: Get on routesClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackRouteDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		rtName := rs.Primary.Attributes["route_table_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for route: %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).routesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, rtName, name)
		if err != nil {
			return fmt.Errorf("Error deleting Route %q (Route Table %q / Resource Group %q): %+v", name, rtName, resourceGroup, err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for deletion of Route %q (Route Table %q / Resource Group %q): %+v", name, rtName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).routesClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_route" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		rtName := rs.Primary.Attributes["route_table_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, rtName, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Route still exists:\n%#v", resp.RoutePropertiesFormat)
		}
	}

	return nil
}

func testAccAzureStackRoute_basic(rInt int, location string) string {
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

resource "azurestack_route" "test" {
  name                = "acctestroute%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  route_table_name    = "${azurestack_route_table.test.name}"

  address_prefix = "10.1.0.0/16"
  next_hop_type  = "vnetlocal"
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackRoute_multipleRoutes(rInt int, location string) string {
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

resource "azurestack_route" "test1" {
  name                = "acctestroute%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  route_table_name    = "${azurestack_route_table.test.name}"

  address_prefix = "10.2.0.0/16"
  next_hop_type  = "none"
}
`, rInt, location, rInt, rInt)
}
