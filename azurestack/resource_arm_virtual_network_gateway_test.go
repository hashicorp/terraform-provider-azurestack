package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackVirtualNetworkGateway_basic(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurestack_virtual_network_gateway.test"
	config := testAccAzureStackVirtualNetworkGateway_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sku", "Basic"),
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

func TestAccAzureStackVirtualNetworkGateway_lowerCaseSubnetName(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurestack_virtual_network_gateway.test"
	config := testAccAzureStackVirtualNetworkGateway_lowerCaseSubnetName(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sku", "Basic"),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualNetworkGateway_standard(t *testing.T) {
	resourceName := "azurestack_virtual_network_gateway.test"
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualNetworkGateway_sku(ri, testLocation(), "Standard")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sku", "Standard"),
				),
			},
		},
	})
}

func TestAccAzureStackVirtualNetworkGateway_vpnClientConfig(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := "azurestack_virtual_network_gateway.test"
	config := testAccAzureStackVirtualNetworkGateway_vpnClientConfig(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpn_client_configuration.0.address_space.0", "10.2.0.0/24"),
					// resource.TestCheckResourceAttr(resourceName, "vpn_client_configuration.0.radius_server_address", "1.2.3.4"),
					// resource.TestCheckResourceAttr(resourceName, "vpn_client_configuration.0.vpn_client_protocols.#", "2"),
				),
			},
		},
	})
}

func testCheckAzureStackVirtualNetworkGatewayExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name, resourceGroup, err := getArmResourceNameAndGroup(s, name)
		if err != nil {
			return err
		}

		client := testAccProvider.Meta().(*ArmClient).vnetGatewayClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on vnetGatewayClient: %+v", err)
		}

		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Bad: Virtual Network Gateway %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureStackVirtualNetworkGatewayDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).vnetGatewayClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_virtual_network_gateway" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return nil
		}

		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Virtual Network Gateway still exists:\n%#v", resp.VirtualNetworkGatewayPropertiesFormat)
		}
	}

	return nil
}

func testAccAzureStackVirtualNetworkGateway_basic(rInt int, location string) string {
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
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackVirtualNetworkGateway_lowerCaseSubnetName(rInt int, location string) string {
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
  name                 = "gatewaySubnet"
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
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackVirtualNetworkGateway_vpnClientConfig(rInt int, location string) string {
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
  depends_on          = ["azurestack_public_ip.test"]
  name                = "acctestvng-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  #sku = "VpnGw1"

  ip_configuration {
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test.id}"
  }
  vpn_client_configuration {
    address_space = ["10.2.0.0/24"]

    #vpn_client_protocols, radius_server_address and radius_server_certificate are not supported yet.
    #vpn_client_protocols = ["SSTP", "IkeV2"]

    #radius_server_address = "1.2.3.4"
    #  radius_server_secret = "1234"
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackVirtualNetworkGateway_sku(rInt int, location string, sku string) string {
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
  sku      = "%s"

  ip_configuration {
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test.id}"
  }
}
`, rInt, location, rInt, rInt, rInt, sku)
}
