package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackVirtualNetworkGatewayConnection_sitetosite(t *testing.T) {
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualNetworkGatewayConnection_sitetosite(ri, testLocation())

	resource.Test(t, resource.TestCase{
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

//Vnet to VnetConnection is not supported yet.
func TestAccAzureStackVirtualNetworkGatewayConnection_vnettonet(t *testing.T) {
	t.Skip()
	firstResourceName := "azurestack_virtual_network_gateway_connection.test_1"
	secondResourceName := "azurestack_virtual_network_gateway_connection.test_2"

	ri := acctest.RandInt()
	ri2 := acctest.RandInt()
	sharedKey := "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
	config := testAccAzureStackVirtualNetworkGatewayConnection_vnettovnet(ri, ri2, sharedKey, testLocation(), testAltLocation())
	fmt.Printf("%+v\n", config)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayConnectionExists(firstResourceName),
					testCheckAzureStackVirtualNetworkGatewayConnectionExists(secondResourceName),
					resource.TestCheckResourceAttr(firstResourceName, "shared_key", sharedKey),
					resource.TestCheckResourceAttr(secondResourceName, "shared_key", sharedKey),
				),
			},
		},
	})
}

//Vnet to VnetConnection is not supported yet.
func TestAccAzureStackVirtualNetworkGatewayConnection_ipsecpolicy(t *testing.T) {
	t.Skip()
	ri := acctest.RandInt()
	config := testAccAzureStackVirtualNetworkGatewayConnection_ipsecpolicy(ri, testLocation())

	resource.Test(t, resource.TestCase{
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
		},
	})
}

//Vnet to VnetConnection is not supported yet.
func TestAccAzureStackVirtualNetworkGatewayConnection_updatingSharedKey(t *testing.T) {
	t.Skip()
	firstResourceName := "azurestack_virtual_network_gateway_connection.test_1"
	secondResourceName := "azurestack_virtual_network_gateway_connection.test_2"

	ri := acctest.RandInt()
	ri2 := acctest.RandInt()
	loc1 := testLocation()
	loc2 := testAltLocation()

	firstSharedKey := "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
	secondSharedKey := "4-r33ly-53cr37-1p53c-5h4r3d-k3y"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackVirtualNetworkGatewayConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackVirtualNetworkGatewayConnection_vnettovnet(ri, ri2, firstSharedKey, loc1, loc2),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayConnectionExists(firstResourceName),
					testCheckAzureStackVirtualNetworkGatewayConnectionExists(secondResourceName),
					resource.TestCheckResourceAttr(firstResourceName, "shared_key", firstSharedKey),
					resource.TestCheckResourceAttr(secondResourceName, "shared_key", firstSharedKey),
				),
			},
			{
				Config: testAccAzureStackVirtualNetworkGatewayConnection_vnettovnet(ri, ri2, secondSharedKey, loc1, loc2),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualNetworkGatewayConnectionExists(firstResourceName),
					testCheckAzureStackVirtualNetworkGatewayConnectionExists(secondResourceName),
					resource.TestCheckResourceAttr(firstResourceName, "shared_key", secondSharedKey),
					resource.TestCheckResourceAttr(secondResourceName, "shared_key", secondSharedKey),
				),
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

		if utils.ResponseWasNotFound(resp.Response) {
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

		if utils.ResponseWasNotFound(resp.Response) {
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

func testAccAzureStackVirtualNetworkGatewayConnection_vnettovnet(rInt, rInt2 int, sharedKey, location, altLocation string) string {
	return fmt.Sprintf(`
variable "random1" {
  default = "%d"
}

variable "random2" {
  default = "%d"
}

variable "shared_key" {
  default = "%s"
}

resource "azurestack_resource_group" "test_1" {
  name     = "acctestRG-${var.random1}"
  location = "%s"
}

resource "azurestack_virtual_network" "test_1" {
  name                = "acctestvn-${var.random1}"
  location            = "${azurestack_resource_group.test_1.location}"
  resource_group_name = "${azurestack_resource_group.test_1.name}"
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test_1" {
  name                 = "GatewaySubnet"
  resource_group_name  = "${azurestack_resource_group.test_1.name}"
  virtual_network_name = "${azurestack_virtual_network.test_1.name}"
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test_1" {
  name                         = "acctest-${var.random1}"
  location                     = "${azurestack_resource_group.test_1.location}"
  resource_group_name          = "${azurestack_resource_group.test_1.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test_1" {
  name                = "acctest-${var.random1}"
  location            = "${azurestack_resource_group.test_1.location}"
  resource_group_name = "${azurestack_resource_group.test_1.name}"

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = "${azurestack_public_ip.test_1.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test_1.id}"
  }
}

resource "azurestack_virtual_network_gateway_connection" "test_1" {
  name                = "acctest-${var.random1}"
  location            = "${azurestack_resource_group.test_1.location}"
  resource_group_name = "${azurestack_resource_group.test_1.name}"

  type                            = "Vnet2Vnet"
  virtual_network_gateway_id      = "${azurestack_virtual_network_gateway.test_1.id}"
  peer_virtual_network_gateway_id = "${azurestack_virtual_network_gateway.test_2.id}"

  shared_key = "${var.shared_key}"
}

resource "azurestack_resource_group" "test_2" {
  name     = "acctestRG-${var.random2}"
  location = "%s"
}

resource "azurestack_virtual_network" "test_2" {
  name                = "acctest-${var.random2}"
  location            = "${azurestack_resource_group.test_2.location}"
  resource_group_name = "${azurestack_resource_group.test_2.name}"
  address_space       = ["10.1.0.0/16"]
}

resource "azurestack_subnet" "test_2" {
  name                 = "GatewaySubnet"
  resource_group_name  = "${azurestack_resource_group.test_2.name}"
  virtual_network_name = "${azurestack_virtual_network.test_2.name}"
  address_prefix       = "10.1.1.0/24"
}

resource "azurestack_public_ip" "test_2" {
  name                         = "acctest-${var.random2}"
  location                     = "${azurestack_resource_group.test_2.location}"
  resource_group_name          = "${azurestack_resource_group.test_2.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test_2" {
  name                = "acctest-${var.random2}"
  location            = "${azurestack_resource_group.test_2.location}"
  resource_group_name = "${azurestack_resource_group.test_2.name}"

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    name                          = "vnetGatewayConfig"
    public_ip_address_id          = "${azurestack_public_ip.test_2.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test_2.id}"
  }
}

resource "azurestack_virtual_network_gateway_connection" "test_2" {
  name                = "acctest-${var.random2}"
  location            = "${azurestack_resource_group.test_2.location}"
  resource_group_name = "${azurestack_resource_group.test_2.name}"

  type                            = "Vnet2Vnet"
  virtual_network_gateway_id      = "${azurestack_virtual_network_gateway.test_2.id}"
  peer_virtual_network_gateway_id = "${azurestack_virtual_network_gateway.test_1.id}"

  shared_key = "${var.shared_key}"
}
`, rInt, rInt2, sharedKey, location, altLocation)
}

func testAccAzureStackVirtualNetworkGatewayConnection_ipsecpolicy(rInt int, location string) string {
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

  #sku = "VpnGw1"
  sku = "Basic"

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

  use_policy_based_traffic_selectors = true
  routing_weight                     = 20

  ipsec_policy {
    dh_group         = "DHGroup14"
    ike_encryption   = "AES256"
    ike_integrity    = "SHA256"
    ipsec_encryption = "AES256"
    ipsec_integrity  = "SHA256"
    pfs_group        = "PFS2048"
    sa_datasize      = 102400000
    sa_lifetime      = 27000
  }

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
`, rInt, location)
}
