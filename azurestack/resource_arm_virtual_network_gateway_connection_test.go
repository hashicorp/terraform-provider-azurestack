package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hashicorp/terraform-provider-azurestack/azurestack/helpers/response"
)

func TestAccAzureStackVirtualNetworkGatewayConnection_sitetosite(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualNetworkGatewayConnection_sitetosite(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayConnectionExists("azurestack_virtual_network_gateway_connection.test"),
				),
			},
			{
				ResourceName:      "azurestack_virtual_network_gateway_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckAzureStackVirtualNetworkGatewayConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name, resourceGroup, err := getArmResourceNameAndGroup(s, name)
		if err != nil {
			return err
		}

		client := testAccProvider.Meta().(*ArmClient).vnetGatewayConnectionsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on vnetGatewayConnectionsClient: %+v", err)
		}

		if response.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Bad: Virtual Network Gateway Connection %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureStackVirtualNetworkGatewayConnectionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).vnetGatewayConnectionsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_virtual_network_gateway_connection" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return nil
		}

		if response.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Virtual Network Gateway Connection still exists: %#v", resp.VirtualNetworkGatewayConnectionPropertiesFormat)
		}
	}

	return nil
}

func testAccAzureStackVirtualNetworkGatewayConnection_sitetosite(rInt int, location string) string {
	return fmt.Sprintf(`
variable "random" {
  default = "%d"
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-${var.random}"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvn-${var.random}"
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
  name                         = "acctest-${var.random}"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "acctest-${var.random}"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test.id}"
  }
}

resource "azurestack_local_network_gateway" "test" {
  name                = "acctest-${var.random}"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  gateway_address = "168.62.225.23"
  address_space   = ["10.1.1.0/24"]
}

resource "azurestack_virtual_network_gateway_connection" "test" {
  name                = "acctest-${var.random}"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  type                       = "IPsec"
  virtual_network_gateway_id = "${azurestack_virtual_network_gateway.test.id}"
  local_network_gateway_id   = "${azurestack_local_network_gateway.test.id}"

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
`, rInt, location)
}
