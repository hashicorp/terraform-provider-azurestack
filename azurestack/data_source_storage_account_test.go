package azurestack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAzureStackStorageAccount_basic(t *testing.T) {
	dataSourceName := "data.azurestack_storage_account.test"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccDataSourceAzureStackStorageAccount_basic(ri, rs, location)
	config := testAccDataSourceAzureStackStorageAccount_basicWithDataSource(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "account_tier", "Standard"),
					resource.TestCheckResourceAttr(dataSourceName, "account_replication_type", "LRS"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.environment", "production"),
				),
			},
		},
	})
}

func testAccDataSourceAzureStackStorageAccount_basic(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestsa-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                = "acctestsads%s"
  resource_group_name = "${azurestack_resource_group.test.name}"

  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccDataSourceAzureStackStorageAccount_basicWithDataSource(rInt int, rString string, location string) string {
	config := testAccDataSourceAzureStackStorageAccount_basic(rInt, rString, location)
	return fmt.Sprintf(`
%s

data "azurestack_storage_account" "test" {
  name                = "${azurestack_storage_account.test.name}"
  resource_group_name = "${azurestack_storage_account.test.resource_group_name}"
}
`, config)
}
