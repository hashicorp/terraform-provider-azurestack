package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAzureStackDnsZone_basic(t *testing.T) {
	resourceName := "azurestack_dns_zone.test"
	ri := acctest.RandInt()
	config := testAccAzureStackDnsZone_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackDnsZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsZoneExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackDnsZone_withTags(t *testing.T) {
	resourceName := "azurestack_dns_zone.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackDnsZone_withTags(ri, location)
	postConfig := testAccAzureStackDnsZone_withTagsUpdate(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackDnsZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsZoneExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsZoneExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testCheckAzureStackDnsZoneExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		zoneName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for DNS zone: %s", zoneName)
		}

		client := testAccProvider.Meta().(*ArmClient).zonesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, resourceGroup, zoneName)
		if err != nil {
			return fmt.Errorf("Bad: Get DNS zone: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: DNS zone %s (resource group: %s) does not exist", zoneName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureStackDnsZoneDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).zonesClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_dns_zone" {
			continue
		}

		zoneName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(ctx, resourceGroup, zoneName)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return nil
			}

			return err
		}

		return fmt.Errorf("DNS Zone still exists:\n%#v", resp)
	}

	return nil
}

func testAccAzureStackDnsZone_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, rInt, location, rInt)
}

func testAccAzureStackDnsZone_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackDnsZone_withTagsUpdate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt)
}
