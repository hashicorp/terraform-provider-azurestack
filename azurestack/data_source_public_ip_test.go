package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataAzureStackPublicIP_basic(t *testing.T) {
	dataSourceName := "data.azurestack_public_ip.test"
	ri := acctest.RandInt()

	name := fmt.Sprintf("acctestpublicip-%d", ri)
	resourceGroupName := fmt.Sprintf("acctestRG-%d", ri)

	config := testAccDataAzureStackPublicIPBasic(name, resourceGroupName, ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "resource_group_name", resourceGroupName),
					resource.TestCheckResourceAttr(dataSourceName, "domain_name_label", fmt.Sprintf("acctest-%d", ri)),
					resource.TestCheckResourceAttr(dataSourceName, "idle_timeout_in_minutes", "30"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fqdn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_address"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.environment", "test"),
				),
			},
		},
	})
}

func testAccDataAzureStackPublicIPBasic(name string, resourceGroupName string, rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "%s"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                         = "%s"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
  domain_name_label            = "acctest-%d"
  idle_timeout_in_minutes      = 30

  tags = {
    environment = "test"
  }
}

data "azurestack_public_ip" "test" {
  name                = "${azurestack_public_ip.test.name}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}
`, resourceGroupName, location, name, rInt)
}
