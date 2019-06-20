package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureStackAvailabilitySet_basic(t *testing.T) {
	resourceName := "azurestack_availability_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackAvailabilitySetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform_update_domain_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "platform_fault_domain_count", "3"),
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

func TestAccAzureStackAvailabilitySet_disappears(t *testing.T) {
	resourceName := "azurestack_availability_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackAvailabilitySetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform_update_domain_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "platform_fault_domain_count", "3"),
					testCheckAzureStackAvailabilitySetDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackAvailabilitySet_withTags(t *testing.T) {
	resourceName := "azurestack_availability_set.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackAvailabilitySet_withTags(ri, location)
	postConfig := testAccAzureStackAvailabilitySet_withUpdatedTags(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackAvailabilitySetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost_center", "MSFT"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackAvailabilitySetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "staging"),
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

func TestAccAzureStackAvailabilitySet_withDomainCounts(t *testing.T) {
	resourceName := "azurestack_availability_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_withDomainCounts(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackAvailabilitySetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform_update_domain_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "platform_fault_domain_count", "3"),
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

// managed not supported in the profile, skipping
func TestAccAzureStackAvailabilitySet_managed(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_availability_set.test"
	ri := acctest.RandInt()
	config := testAccAzureStackAvailabilitySet_managed(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackAvailabilitySetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackAvailabilitySetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed", "true"),
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

func testCheckAzureStackAvailabilitySetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		availSetName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for availability set: %s", availSetName)
		}

		client := testAccProvider.Meta().(*ArmClient).availSetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, resourceGroup, availSetName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Availability Set %q (resource group: %q) does not exist", name, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on availSetClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackAvailabilitySetDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		availSetName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for availability set: %s", availSetName)
		}

		client := testAccProvider.Meta().(*ArmClient).availSetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Delete(ctx, resourceGroup, availSetName)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Delete on availSetClient: %+v", err)
			}
		}

		return nil
	}
}

func testCheckAzureStackAvailabilitySetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_availability_set" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		client := testAccProvider.Meta().(*ArmClient).availSetClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, resourceGroup, name)

		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Availability Set still exists:\n%#v", resp.AvailabilitySetProperties)
	}

	return nil
}

func testAccAzureStackAvailabilitySet_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                = "acctestavset-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location, rInt)
}

func testAccAzureStackAvailabilitySet_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                = "acctestavset-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackAvailabilitySet_withUpdatedTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                = "acctestavset-%d"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackAvailabilitySet_withDomainCounts(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                         = "acctestavset-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  platform_update_domain_count = 3
  platform_fault_domain_count  = 3
}
`, rInt, location, rInt)
}

func testAccAzureStackAvailabilitySet_managed(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_availability_set" "test" {
  name                         = "acctestavset-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  platform_update_domain_count = 3
  platform_fault_domain_count  = 3
  managed                      = true
}
`, rInt, location, rInt)
}
