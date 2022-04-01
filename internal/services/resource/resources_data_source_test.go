package resource_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type ResourcesDataSource struct{}

func TestAccResourcesDataSource_ByName(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_resources", "test")
	r := ResourcesDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.template(data),
		},
		{
			Config: r.ByName(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("resources.#").HasValue("1"),
			),
		},
	})
}

func TestAccResourcesDataSource_ByResourceGroup(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_resources", "test")
	r := ResourcesDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.template(data),
		},
		{
			Config: r.ByResourceGroup(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("resources.#").HasValue("1"),
			),
		},
	})
}

func TestAccResourcesDataSource_ByResourceType(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_resources", "test")
	r := ResourcesDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.template(data),
		},
		{
			Config: r.ByResourceType(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("resources.#").HasValue("1"),
			),
		},
	})
}

func TestAccResourcesDataSource_FilteredByTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_resources", "test")
	r := ResourcesDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.template(data),
		},
		{
			Config: r.FilteredByTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("resources.#").HasValue("1"),
			),
		},
	})
}

func (r ResourcesDataSource) ByName(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_resources" "test" {
  name = azurestack_storage_account.test.name
}
`, r.template(data))
}

func (r ResourcesDataSource) ByResourceGroup(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_resources" "test" {
  resource_group_name = azurestack_storage_account.test.resource_group_name
}
`, r.template(data))
}

func (r ResourcesDataSource) ByResourceType(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_resources" "test" {
  resource_group_name = azurestack_storage_account.test.resource_group_name
  type                = "Microsoft.Storage/storageAccounts"
}
`, r.template(data))
}

func (r ResourcesDataSource) FilteredByTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_resources" "test" {
  name                = azurestack_storage_account.test.name
  resource_group_name = azurestack_storage_account.test.resource_group_name

  required_tags = {
    environment = "production"
  }
}
`, r.template(data))
}

func (ResourcesDataSource) template(data acceptance.TestData) string {
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
