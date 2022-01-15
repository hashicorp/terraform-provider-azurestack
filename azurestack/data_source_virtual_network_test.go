package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArmVirtualNetwork_basic(t *testing.T) {
	dataSourceName := "data.azurestack_virtual_network.test"
	ri := acctest.RandInt()

	name := fmt.Sprintf("acctestvnet-%d", ri)
	config := testAccDataSourceArmVirtualNetwork_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "dns_servers.0", "10.0.0.4"),
					resource.TestCheckResourceAttr(dataSourceName, "address_spaces.0", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(dataSourceName, "subnets.0", "subnet1"),
				),
			},
		},
	})
}

func testAccDataSourceArmVirtualNetwork_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvnet-%d"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  dns_servers         = ["10.0.0.4"]

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }
}

data "azurestack_virtual_network" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  name                = "${azurestack_virtual_network.test.name}"
}
`, rInt, location, rInt)
}
