package azurestack

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-10-01/network"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureStackLoadBalancerNatPool_basic(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natPoolName := fmt.Sprintf("NatPool-%d", ri)

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	natPoolId := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/inboundNatPools/%s",
		subscriptionID, ri, ri, natPoolName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatPool_basic(ri, natPoolName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPoolName, &lb),
					resource.TestCheckResourceAttr(
						"azurestack_lb_nat_pool.test", "id", natPoolId),
				),
			},
			{
				ResourceName:      "azurestack_lb_nat_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatPool_removal(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natPoolName := fmt.Sprintf("NatPool-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatPool_basic(ri, natPoolName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPoolName, &lb),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerNatPool_removal(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatPoolNotExists(natPoolName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatPool_update(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natPoolName := fmt.Sprintf("NatPool-%d", ri)
	natPool2Name := fmt.Sprintf("NatPool-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatPool_multiplePools(ri, natPoolName, natPool2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPoolName, &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPool2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_nat_pool.test2", "backend_port", "3390"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerNatPool_multiplePoolsUpdate(ri, natPoolName, natPool2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPoolName, &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPool2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_nat_pool.test2", "backend_port", "3391"),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatPool_disappears(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natPoolName := fmt.Sprintf("NatPool-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatPool_basic(ri, natPoolName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatPoolExists(natPoolName, &lb),
					testCheckAzureStackLoadBalancerNatPoolDisappears(natPoolName, &lb),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureStackLoadBalancerNatPoolExists(natPoolName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerNatPoolByName(lb, natPoolName)
		if !exists {
			return fmt.Errorf("A NAT Pool with name %q cannot be found.", natPoolName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerNatPoolNotExists(natPoolName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerNatPoolByName(lb, natPoolName)
		if exists {
			return fmt.Errorf("A NAT Pool with name %q has been found.", natPoolName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerNatPoolDisappears(natPoolName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		_, i, exists := findLoadBalancerNatPoolByName(lb, natPoolName)
		if !exists {
			return fmt.Errorf("A Nat Pool with name %q cannot be found.", natPoolName)
		}

		currentPools := *lb.LoadBalancerPropertiesFormat.InboundNatPools
		pools := append(currentPools[:i], currentPools[i+1:]...)
		lb.LoadBalancerPropertiesFormat.InboundNatPools = &pools

		id, err := parseAzureResourceID(*lb.ID)
		if err != nil {
			return err
		}

		future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, *lb.Name, *lb)
		if err != nil {
			return fmt.Errorf("Error Creating/Updating LoadBalancer %+v", err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for the completion of LoadBalancer %+v", err)
		}

		_, err = client.Get(ctx, id.ResourceGroup, *lb.Name, "")
		return err
	}
}

func testAccAzureStackLoadBalancerNatPool_basic(rInt int, natPoolName string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_lb_nat_pool" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natPoolName, rInt)
}

func testAccAzureStackLoadBalancerNatPool_removal(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureStackLoadBalancerNatPool_multiplePools(rInt int, natPoolName, natPool2Name string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_lb_nat_pool" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurestack_lb_nat_pool" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port_start            = 82
  frontend_port_end              = 83
  backend_port                   = 3390
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natPoolName, rInt, natPool2Name, rInt)
}

func testAccAzureStackLoadBalancerNatPool_multiplePoolsUpdate(rInt int, natPoolName, natPool2Name string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_lb_nat_pool" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurestack_lb_nat_pool" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port_start            = 82
  frontend_port_end              = 83
  backend_port                   = 3391
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natPoolName, rInt, natPool2Name, rInt)
}
