package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2015-06-15/network"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceAzureStackLoadBalancerPrivateIpAddressAllocation_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "Random",
			ErrCount: 1,
		},
		{
			Value:    "Static",
			ErrCount: 0,
		},
		{
			Value:    "Dynamic",
			ErrCount: 0,
		},
		{
			Value:    "STATIC",
			ErrCount: 0,
		},
		{
			Value:    "static",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateLoadBalancerPrivateIpAddressAllocation(tc.Value, "azurestack_lb")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM LoadBalancer private_ip_address_allocation to trigger a validation error")
		}
	}
}

func TestAccAzureStackLoadBalancer_basic(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancer_basic(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancer_standard(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancer_standard(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists("azurestack_lb.test", &lb),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancer_frontEndConfig(t *testing.T) {
	var lb network.LoadBalancer
	resourceName := "azurestack_lb.test"
	ri := acctest.RandInt()
	location := testLocation()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancer_frontEndConfig(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "frontend_ip_configuration.#", "2"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancer_frontEndConfigRemovalWithIP(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "frontend_ip_configuration.#", "1"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancer_frontEndConfigRemoval(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "frontend_ip_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccAzureStackLoadBalancer_tags(t *testing.T) {
	var lb network.LoadBalancer
	resourceName := "azurestack_lb.test"
	ri := acctest.RandInt()
	location := testLocation()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackLoadBalancer_basic(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Purpose", "AcceptanceTests"),
				),
			},
			{
				Config: testAccAzureStackLoadBalancer_updatedTags(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackLoadBalancerExists(resourceName, &lb),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Purpose", "AcceptanceTests"),
				),
			},
		},
	})
}

func testCheckAzureStackLoadBalancerExists(name string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		loadBalancerName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for loadbalancer: %s", loadBalancerName)
		}

		client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, loadBalancerName, "")
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return fmt.Errorf("Bad: LoadBalancer %q (resource group: %q) does not exist", loadBalancerName, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on loadBalancerClient: %+v", err)
		}

		*lb = resp

		return nil
	}
}

func testCheckAzureStackLoadBalancerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_lb" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name, "")

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("LoadBalancer still exists:\n%#v", resp.LoadBalancerPropertiesFormat)
		}
	}

	return nil
}

func testAccAzureStackLoadBalancer_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags {
    Environment = "production"
    Purpose     = "AcceptanceTests"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackLoadBalancer_standard(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_lb" "test" {
  name                = "acctest-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags {
    Environment = "production"
    Purpose     = "AcceptanceTests"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackLoadBalancer_updatedTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags {
    Purpose = "AcceptanceTests"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackLoadBalancer_frontEndConfig(rInt int, location string) string {
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

resource "azurestack_public_ip" "test1" {
  name                         = "another-test-ip-%d"
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

  frontend_ip_configuration {
    name                 = "two-%d"
    public_ip_address_id = "${azurestack_public_ip.test1.id}"
  }
}
`, rInt, location, rInt, rInt, rInt, rInt, rInt)
}

func testAccAzureStackLoadBalancer_frontEndConfigRemovalWithIP(rInt int, location string) string {
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

resource "azurestack_public_ip" "test1" {
  name                         = "another-test-ip-%d"
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
`, rInt, location, rInt, rInt, rInt, rInt)
}

func testAccAzureStackLoadBalancer_frontEndConfigRemoval(rInt int, location string) string {
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
