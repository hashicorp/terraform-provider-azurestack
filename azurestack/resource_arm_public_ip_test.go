package azurestack

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceAzureStackPublicIpDomainNameLabel_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandString(80),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validatePublicIpDomainNameLabel(tc.Value, "azurestack_public_ip")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Public IP Domain Name Label to trigger a validation error")
		}
	}
}

func TestAccAzureStackPublicIpStatic_basic(t *testing.T) {
	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
					resource.TestCheckResourceAttr(resourceName, "public_ip_address_allocation", "static"),
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

// Sku not supported in the profile, skipping
func TestAccAzureStackPublicIpStatic_standard(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_standard(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
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

func TestAccAzureStackPublicIpStatic_disappears(t *testing.T) {
	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
					testCheckAzureStackPublicIpDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackPublicIpStatic_idleTimeout(t *testing.T) {
	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_idleTimeout(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "idle_timeout_in_minutes", "30"),
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

func TestAccAzureStackPublicIpStatic_withTags(t *testing.T) {
	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackPublicIPStatic_withTags(ri, location)
	postConfig := testAccAzureStackPublicIPStatic_withTagsUpdate(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost_center", "MSFT"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
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

func TestAccAzureStackPublicIpStatic_update(t *testing.T) {
	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackPublicIPStatic_basic(ri, location)
	postConfig := testAccAzureStackPublicIPStatic_update(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_label", fmt.Sprintf("acctest-%d", ri)),
				),
			},
		},
	})
}

func TestAccAzureStackPublicIpDynamic_basic(t *testing.T) {
	resourceName := "azurestack_public_ip.test"
	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPDynamic_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackPublicIpExists(resourceName),
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

func TestAccAzureStackPublicIpStatic_importIdError(t *testing.T) {
	resourceName := "azurestack_public_ip.test"

	ri := acctest.RandInt()
	config := testAccAzureStackPublicIPStatic_basic(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/publicIPAdresses/acctestpublicip-%d", os.Getenv("ARM_SUBSCRIPTION_ID"), ri, ri),
				ExpectError:       regexp.MustCompile("Error parsing supplied resource id."),
			},
		},
	})
}

func testCheckAzureStackPublicIpExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		publicIPName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for public ip: %s", publicIPName)
		}

		client := testAccProvider.Meta().(*ArmClient).publicIPClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, publicIPName, "")
		if err != nil {
			return fmt.Errorf("Bad: Get on publicIPClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Public IP %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureStackPublicIpDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		publicIpName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for public ip: %s", publicIpName)
		}

		client := testAccProvider.Meta().(*ArmClient).publicIPClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		future, err := client.Delete(ctx, resourceGroup, publicIpName)
		if err != nil {
			return fmt.Errorf("Error deleting Public IP %q (Resource Group %q): %+v", publicIpName, resourceGroup, err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for deletion of Public IP %q (Resource Group %q): %+v", publicIpName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureStackPublicIpDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).publicIPClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_public_ip" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name, "")

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Public IP still exists:\n%#v", resp.PublicIPAddressPropertiesFormat)
		}
	}

	return nil
}

func testAccAzureStackPublicIPStatic_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}
`, rInt, location, rInt)
}

func testAccAzureStackPublicIPStatic_standard(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
  sku                          = "standard"
}
`, rInt, location, rInt)
}

func testAccAzureStackPublicIPStatic_update(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
  domain_name_label            = "acctest-%d"
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackPublicIPStatic_idleTimeout(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
  idle_timeout_in_minutes      = 30
}
`, rInt, location, rInt)
}

func testAccAzureStackPublicIPDynamic_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "dynamic"
}
`, rInt, location, rInt)
}

func testAccAzureStackPublicIPStatic_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackPublicIPStatic_withTagsUpdate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "acctestpublicip-%d"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt)
}
