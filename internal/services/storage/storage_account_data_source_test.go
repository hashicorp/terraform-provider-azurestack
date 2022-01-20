package storage_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type StorageAccountDataSource struct{}

func TestAccDataSourceStorageAccount_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_storage_account", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: StorageAccountDataSource{}.basic(data),
		},
		{
			Config: StorageAccountDataSource{}.basicWithDataSource(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("account_tier").HasValue("Standard"),
				check.That(data.ResourceName).Key("account_replication_type").HasValue("LRS"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("production"),
			),
		},
	})
}

func (d StorageAccountDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                = "acctestsads%s"
  resource_group_name = azurestack_resource_group.test.name

  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (d StorageAccountDataSource) basicWithDataSource(data acceptance.TestData) string {
	config := d.basic(data)
	return fmt.Sprintf(`
%s

data "azurestack_storage_account" "test" {
  name                = azurestack_storage_account.test.name
  resource_group_name = azurestack_storage_account.test.resource_group_name
}
`, config)
}
