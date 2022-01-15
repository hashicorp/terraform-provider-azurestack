package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAzureStackDataSourceVirtualNetworkGateway_basic(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackDataSourceVirtualNetworkGateway_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayExists("data.azurestack_virtual_network_gateway.test"),
				),
			},
		},
	})
}

func testAccAzureStackDataSourceVirtualNetworkGateway_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test" {
  name                 = "GatewaySubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctestvng-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test.id}"
  }
}

data "azurestack_virtual_network_gateway" "test" {
  name                = "${azurestack_virtual_network_gateway.test.name}"
  resource_group_name = "${azurestack_virtual_network_gateway.test.resource_group_name}"
}
`, rInt, location, rInt, rInt, rInt)
}
