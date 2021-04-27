package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2016-04-01/dns"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAzureStackDnsARecord_basic(t *testing.T) {
	resourceName := "azurestack_dns_a_record.test"
	ri := acctest.RandInt()
	config := testAccAzureStackDnsARecord_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackDnsARecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsARecordExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureStackDnsARecord_updateRecords(t *testing.T) {
	resourceName := "azurestack_dns_a_record.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackDnsARecord_basic(ri, location)
	postConfig := testAccAzureStackDnsARecord_updateRecords(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackDnsARecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsARecordExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "records.#", "2"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsARecordExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "records.#", "3"),
				),
			},
		},
	})
}

func TestAccAzureStackDnsARecord_withTags(t *testing.T) {
	resourceName := "azurestack_dns_a_record.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureStackDnsARecord_withTags(ri, location)
	postConfig := testAccAzureStackDnsARecord_withTagsUpdate(ri, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackDnsARecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsARecordExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackDnsARecordExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testCheckAzureStackDnsARecordExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		aName := rs.Primary.Attributes["name"]
		zoneName := rs.Primary.Attributes["zone_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for DNS A record: %s", aName)
		}

		conn := testAccProvider.Meta().(*ArmClient).dnsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := conn.Get(ctx, resourceGroup, zoneName, aName, dns.A)
		if err != nil {
			return fmt.Errorf("Bad: Get A RecordSet: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: DNS A record %s (resource group: %s) does not exist", aName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureStackDnsARecordDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).dnsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_dns_a_record" {
			continue
		}

		aName := rs.Primary.Attributes["name"]
		zoneName := rs.Primary.Attributes["zone_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(ctx, resourceGroup, zoneName, aName, dns.A)

		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return nil
			}

			return err
		}

		return fmt.Errorf("DNS A record still exists:\n%#v", resp.RecordSetProperties)
	}

	return nil
}

func testAccAzureStackDnsARecord_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_dns_a_record" "test" {
  name                = "myarecord%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  zone_name           = "${azurestack_dns_zone.test.name}"
  ttl                 = 300
  records             = ["1.2.3.4", "1.2.4.5"]
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackDnsARecord_updateRecords(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_dns_a_record" "test" {
  name                = "myarecord%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  zone_name           = "${azurestack_dns_zone.test.name}"
  ttl                 = 300
  records             = ["1.2.3.4", "1.2.4.5", "1.2.3.7"]
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackDnsARecord_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_dns_a_record" "test" {
  name                = "myarecord%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  zone_name           = "${azurestack_dns_zone.test.name}"
  ttl                 = 300
  records             = ["1.2.3.4", "1.2.4.5"]

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackDnsARecord_withTagsUpdate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG_%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_dns_a_record" "test" {
  name                = "myarecord%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  zone_name           = "${azurestack_dns_zone.test.name}"
  ttl                 = 300
  records             = ["1.2.3.4", "1.2.4.5"]

  tags = {
    environment = "staging"
  }
}
`, rInt, location, rInt, rInt)
}
