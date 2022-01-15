package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAzureStackRouteTable_basic(t *testing.T) {
	dataSourceName := "data.azurestack_route_table.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccDataSourceAzureStackRouteTable_basic(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "route.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureStackRouteTable_singleRoute(t *testing.T) {
	dataSourceName := "data.azurestack_route_table.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccDataSourceAzureStackRouteTable_singleRoute(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "route.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "route.0.name", "route1"),
					resource.TestCheckResourceAttr(dataSourceName, "route.0.address_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(dataSourceName, "route.0.next_hop_type", "VnetLocal"),
				),
			},
		},
	})
}

func TestAccDataSourceAzureStackRouteTable_multipleRoutes(t *testing.T) {
	dataSourceName := "data.azurestack_route_table.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccDataSourceAzureStackRouteTable_multipleRoutes(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackRouteTableExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "route.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "route.0.name", "route1"),
					resource.TestCheckResourceAttr(dataSourceName, "route.0.address_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(dataSourceName, "route.0.next_hop_type", "VnetLocal"),
					resource.TestCheckResourceAttr(dataSourceName, "route.1.name", "route2"),
					resource.TestCheckResourceAttr(dataSourceName, "route.1.address_prefix", "10.2.0.0/16"),
					resource.TestCheckResourceAttr(dataSourceName, "route.1.next_hop_type", "VnetLocal"),
				),
			},
		},
	})
}

func testAccDataSourceAzureStackRouteTable_basic(rInt int, location string) string {
	resource := testAccAzureStackRouteTable_basic(rInt, location)
	return fmt.Sprintf(`
%s

data "azurestack_route_table" "test" {
  name                = "${azurestack_route_table.test.name}"
  resource_group_name = "${azurestack_route_table.test.resource_group_name}"
}
`, resource)
}

func testAccDataSourceAzureStackRouteTable_singleRoute(rInt int, location string) string {
	resource := testAccAzureStackRouteTable_singleRoute(rInt, location)
	return fmt.Sprintf(`
%s

data "azurestack_route_table" "test" {
  name                = "${azurestack_route_table.test.name}"
  resource_group_name = "${azurestack_route_table.test.resource_group_name}"
}
`, resource)
}

func testAccDataSourceAzureStackRouteTable_multipleRoutes(rInt int, location string) string {
	resource := testAccAzureStackRouteTable_multipleRoutes(rInt, location)
	return fmt.Sprintf(`
%s

data "azurestack_route_table" "test" {
  name                = "${azurestack_route_table.test.name}"
  resource_group_name = "${azurestack_route_table.test.resource_group_name}"
}
`, resource)
}
