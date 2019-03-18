package azurestack

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2015-06-15/network"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceAzureStackLoadBalancerRuleNameLabel_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "-word",
			ErrCount: 1,
		},
		{
			Value:    "testing-",
			ErrCount: 1,
		},
		{
			Value:    "test#test",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(81, "abcdedfed"),
			ErrCount: 1,
		},
		{
			Value:    "test.rule",
			ErrCount: 0,
		},
		{
			Value:    "test_rule",
			ErrCount: 0,
		},
		{
			Value:    "test-rule",
			ErrCount: 0,
		},
		{
			Value:    "TestRule",
			ErrCount: 0,
		},
		{
			Value:    "Test123Rule",
			ErrCount: 0,
		},
		{
			Value:    "TestRule",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateArmLoadBalancerRuleName(tc.Value, "azurestack_lb_rule")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM LoadBalancer Rule Name Label to trigger a validation error")
		}
	}
}

func TestAccAzureStackLoadBalancerRule_basic(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	lbRuleID := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/loadBalancingRules/%s",
		subscriptionID, ri, ri, lbRuleName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerRule_basic(ri, lbRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
					resource.TestCheckResourceAttr(
						"azurestack_lb_rule.test", "id", lbRuleID),
				),
			},
			{
				ResourceName:      "azurestack_lb_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackLoadBalancerRule_removal(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerRule_basic(ri, lbRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerRule_removal(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleNotExists(lbRuleName, &lb),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/9424
func TestAccAzureStackLoadBalancerRule_inconsistentReads(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	backendPoolName := fmt.Sprintf("LbPool-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	probeName := fmt.Sprintf("LbProbe-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerRule_inconsistentRead(ri, backendPoolName, probeName, lbRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerBackEndAddressPoolExists(backendPoolName, &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerRule_update(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	lbRule2Name := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	lbRuleID := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/loadBalancingRules/%s",
		subscriptionID, ri, ri, lbRuleName)

	lbRule2ID := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/loadBalancingRules/%s",
		subscriptionID, ri, ri, lbRule2Name)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerRule_multipleRules(ri, lbRuleName, lbRule2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRule2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test", "id", lbRuleID),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test2", "id", lbRule2ID),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test2", "frontend_port", "3390"),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test2", "backend_port", "3390"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerRule_multipleRulesUpdate(ri, lbRuleName, lbRule2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRule2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test", "id", lbRuleID),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test2", "id", lbRule2ID),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test2", "frontend_port", "3391"),
					resource.TestCheckResourceAttr("azurestack_lb_rule.test2", "backend_port", "3391"),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerRule_reapply(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	deleteRuleState := func(s *terraform.State) error {
		return s.Remove("azurestack_lb_rule.test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerRule_basic(ri, lbRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
					deleteRuleState,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAzureStackLoadBalancerRule_basic(ri, lbRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerRule_disappears(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerRule_basic(ri, lbRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerRuleExists(lbRuleName, &lb),
					testCheckAzureStackLoadBalancerRuleDisappears(lbRuleName, &lb),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureStackLoadBalancerRuleExists(lbRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerRuleByName(lb, lbRuleName)
		if !exists {
			return fmt.Errorf("A LoadBalancer Rule with name %q cannot be found.", lbRuleName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerRuleNotExists(lbRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerRuleByName(lb, lbRuleName)
		if exists {
			return fmt.Errorf("A LoadBalancer Rule with name %q has been found.", lbRuleName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerRuleDisappears(ruleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		_, i, exists := findLoadBalancerRuleByName(lb, ruleName)
		if !exists {
			return fmt.Errorf("A Rule with name %q cannot be found.", ruleName)
		}

		currentRules := *lb.LoadBalancerPropertiesFormat.LoadBalancingRules
		rules := append(currentRules[:i], currentRules[i+1:]...)
		lb.LoadBalancerPropertiesFormat.LoadBalancingRules = &rules

		id, err := parseAzureResourceID(*lb.ID)
		if err != nil {
			return err
		}

		future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, *lb.Name, *lb)
		if err != nil {
			return fmt.Errorf("Error Creating/Updating LoadBalancer %q (Resource Group %q): %+v", *lb.Name, id.ResourceGroup, err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for completion of LoadBalancer %q (Resource Group %q): %+v", *lb.Name, id.ResourceGroup, err)
		}

		_, err = client.Get(ctx, id.ResourceGroup, *lb.Name, "")
		return err
	}
}

func testAccAzureStackLoadBalancerRule_basic(rInt int, lbRuleName string, location string) string {
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

resource "azurestack_lb_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, lbRuleName, rInt)
}

func testAccAzureStackLoadBalancerRule_removal(rInt int, location string) string {
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

// https://github.com/hashicorp/terraform/issues/9424
func testAccAzureStackLoadBalancerRule_inconsistentRead(rInt int, backendPoolName, probeName, lbRuleName string, location string) string {
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

resource "azurestack_lb_backend_address_pool" "teset" {
  name                = "%s"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
}

resource "azurestack_lb_probe" "test" {
  name                = "%s"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  protocol            = "Tcp"
  port                = 443
}

resource "azurestack_lb_rule" "test" {
  name                           = "%s"
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, backendPoolName, probeName, lbRuleName, rInt)
}

func testAccAzureStackLoadBalancerRule_multipleRules(rInt int, lbRuleName, lbRule2Name string, location string) string {
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

resource "azurestack_lb_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Udp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurestack_lb_rule" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Udp"
  frontend_port                  = 3390
  backend_port                   = 3390
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, lbRuleName, rInt, lbRule2Name, rInt)
}

func testAccAzureStackLoadBalancerRule_multipleRulesUpdate(rInt int, lbRuleName, lbRule2Name string, location string) string {
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

resource "azurestack_lb_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Udp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurestack_lb_rule" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "%s"
  protocol                       = "Udp"
  frontend_port                  = 3391
  backend_port                   = 3391
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, lbRuleName, rInt, lbRule2Name, rInt)
}
