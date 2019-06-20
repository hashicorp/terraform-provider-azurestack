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

func TestAccAzureStackLoadBalancerNatRule_basic(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	natRuleId := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/inboundNatRules/%s",
		subscriptionID, ri, ri, natRuleName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
					resource.TestCheckResourceAttr(
						"azurestack_lb_nat_rule.test", "id", natRuleId),
				),
			},
			{
				ResourceName:      "azurestack_lb_nat_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatRule_removal(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerNatRule_removal(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleNotExists(natRuleName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatRule_update(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)
	natRule2Name := fmt.Sprintf("NatRule-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatRule_multipleRules(ri, natRuleName, natRule2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRule2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_nat_rule.test2", "frontend_port", "3390"),
					resource.TestCheckResourceAttr("azurestack_lb_nat_rule.test2", "backend_port", "3390"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerNatRule_multipleRulesUpdate(ri, natRuleName, natRule2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRule2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_nat_rule.test2", "frontend_port", "3391"),
					resource.TestCheckResourceAttr("azurestack_lb_nat_rule.test2", "backend_port", "3391"),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatRule_disappears(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
					testCheckAzureStackLoadBalancerNatRuleDisappears(natRuleName, &lb),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatRule_enableFloatingIP(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatRule_enableFloatingIP(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerNatRule_disableFloatingIP(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerNatRule_basic(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerNatRule_enableFloatingIP(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerNatRule_basic(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
		},
	})
}

func testCheckAzureStackLoadBalancerNatRuleExists(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerNatRuleByName(lb, natRuleName)
		if !exists {
			return fmt.Errorf("A NAT Rule with name %q cannot be found.", natRuleName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerNatRuleNotExists(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerNatRuleByName(lb, natRuleName)
		if exists {
			return fmt.Errorf("A NAT Rule with name %q has been found.", natRuleName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerNatRuleDisappears(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		_, i, exists := findLoadBalancerNatRuleByName(lb, natRuleName)
		if !exists {
			return fmt.Errorf("A Nat Rule with name %q cannot be found.", natRuleName)
		}

		currentRules := *lb.LoadBalancerPropertiesFormat.InboundNatRules
		rules := append(currentRules[:i], currentRules[i+1:]...)
		lb.LoadBalancerPropertiesFormat.InboundNatRules = &rules

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
			return fmt.Errorf("Error waiting for the completion of LoadBalancer %q (Resource Group %q): %+v", *lb.Name, id.ResourceGroup, err)
		}

		_, err = client.Get(ctx, id.ResourceGroup, *lb.Name, "")
		return err
	}
}

func testAccAzureStackLoadBalancerNatRule_basic(rInt int, natRuleName string, location string) string {
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

resource "azurestack_lb_nat_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt)
}

func testAccAzureStackLoadBalancerNatRule_removal(rInt int, location string) string {
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

func testAccAzureStackLoadBalancerNatRule_multipleRules(rInt int, natRuleName, natRule2Name string, location string) string {
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

resource "azurestack_lb_nat_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurestack_lb_nat_rule" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3390
  backend_port                   = 3390
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt, natRule2Name, rInt)
}

func testAccAzureStackLoadBalancerNatRule_multipleRulesUpdate(rInt int, natRuleName, natRule2Name string, location string) string {
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

resource "azurestack_lb_nat_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurestack_lb_nat_rule" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3391
  backend_port                   = 3391
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt, natRule2Name, rInt)
}

func testAccAzureStackLoadBalancerNatRule_enableFloatingIP(rInt int, natRuleName string, location string) string {
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

resource "azurestack_lb_nat_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt)
}
