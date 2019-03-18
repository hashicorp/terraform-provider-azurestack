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

func TestAccAzureStackLoadBalancerProbe_basic(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	probeID := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/probes/%s",
		subscriptionID, ri, ri, probeName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_basic(ri, probeName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					resource.TestCheckResourceAttr(
						"azurestack_lb_probe.test", "id", probeID),
				),
			},
			{
				ResourceName:      "azurestack_lb_probe.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackLoadBalancerProbe_removal(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)
	location := testLocation()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_basic(ri, probeName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerProbe_removal(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeNotExists(probeName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerProbe_update(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)
	probe2Name := fmt.Sprintf("probe-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_multipleProbes(ri, probeName, probe2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					testCheckAzureStackLoadBalancerProbeExists(probe2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_probe.test2", "port", "80"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerProbe_multipleProbesUpdate(ri, probeName, probe2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					testCheckAzureStackLoadBalancerProbeExists(probe2Name, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_probe.test2", "port", "8080"),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerProbe_updateProtocol(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_updateProtocolBefore(ri, probeName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_probe.test", "protocol", "Http"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancerProbe_updateProtocolAfter(ri, probeName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					resource.TestCheckResourceAttr("azurestack_lb_probe.test", "protocol", "Tcp"),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerProbe_reapply(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)

	deleteProbeState := func(s *terraform.State) error {
		return s.Remove("azurestack_lb_probe.test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_basic(ri, probeName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					deleteProbeState,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAzureStackLoadBalancerProbe_basic(ri, probeName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancerProbe_disappears(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	probeName := fmt.Sprintf("probe-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancerProbe_basic(ri, probeName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
					testCheckAzureStackLoadBalancerProbeExists(probeName, &lb),
					testCheckAzureStackLoadBalancerProbeDisappears(probeName, &lb),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureStackLoadBalancerProbeExists(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerProbeByName(lb, natRuleName)
		if !exists {
			return fmt.Errorf("A Probe with name %q cannot be found.", natRuleName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerProbeNotExists(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, _, exists := findLoadBalancerProbeByName(lb, natRuleName)
		if exists {
			return fmt.Errorf("A Probe with name %q has been found.", natRuleName)
		}

		return nil
	}
}

func testCheckAzureStackLoadBalancerProbeDisappears(addressPoolName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		_, i, exists := findLoadBalancerProbeByName(lb, addressPoolName)
		if !exists {
			return fmt.Errorf("A Probe with name %q cannot be found.", addressPoolName)
		}

		currentProbes := *lb.LoadBalancerPropertiesFormat.Probes
		probes := append(currentProbes[:i], currentProbes[i+1:]...)
		lb.LoadBalancerPropertiesFormat.Probes = &probes

		id, err := parseAzureResourceID(*lb.ID)
		if err != nil {
			return err
		}

		future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, *lb.Name, *lb)
		if err != nil {
			return fmt.Errorf("Error Creating/Updating LoadBalancer: %+v", err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for completion for LoadBalancer: %+v", err)
		}

		_, err = client.Get(ctx, id.ResourceGroup, *lb.Name, "")
		return err
	}
}

func testAccAzureStackLoadBalancerProbe_basic(rInt int, probeName string, location string) string {
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

resource "azurestack_lb_probe" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  port                = 22
}
`, rInt, location, rInt, rInt, rInt, probeName)
}

func testAccAzureStackLoadBalancerProbe_removal(rInt int, location string) string {
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

func testAccAzureStackLoadBalancerProbe_multipleProbes(rInt int, probeName, probe2Name string, location string) string {
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

resource "azurestack_lb_probe" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  port                = 22
}

resource "azurestack_lb_probe" "test2" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  port                = 80
}
`, rInt, location, rInt, rInt, rInt, probeName, probe2Name)
}

func testAccAzureStackLoadBalancerProbe_multipleProbesUpdate(rInt int, probeName, probe2Name string, location string) string {
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

resource "azurestack_lb_probe" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  port                = 22
}

resource "azurestack_lb_probe" "test2" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  port                = 8080
}
`, rInt, location, rInt, rInt, rInt, probeName, probe2Name)
}

func testAccAzureStackLoadBalancerProbe_updateProtocolBefore(rInt int, probeName string, location string) string {
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

resource "azurestack_lb_probe" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  protocol            = "Http"
  request_path        = "/"
  port                = 80
}
`, rInt, location, rInt, rInt, rInt, probeName)
}

func testAccAzureStackLoadBalancerProbe_updateProtocolAfter(rInt int, probeName string, location string) string {
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

resource "azurestack_lb_probe" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "%s"
  protocol            = "Tcp"
  port                = 80
}
`, rInt, location, rInt, rInt, rInt, probeName)
}
