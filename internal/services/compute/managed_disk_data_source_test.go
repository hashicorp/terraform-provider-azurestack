package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type ManagedDiskDataSource struct{}

func TestAccManagedDiskDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_managed_disk", "test")
	r := ManagedDiskDataSource{}

	name := fmt.Sprintf("acctestmanageddisk-%d", data.RandomInteger)
	resourceGroupName := fmt.Sprintf("acctestRG-%d", data.RandomInteger)

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data, name, resourceGroupName),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").HasValue(name),
				check.That(data.ResourceName).Key("resource_group_name").HasValue(resourceGroupName),
				check.That(data.ResourceName).Key("storage_account_type").HasValue("Premium_LRS"),
				check.That(data.ResourceName).Key("disk_size_gb").HasValue("10"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("acctest"),
			),
		},
	})
}

func (ManagedDiskDataSource) basic(data acceptance.TestData, name string, resourceGroupName string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "%s"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "%s"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Premium_LRS"
  create_option        = "Empty"
  disk_size_gb         = "10"

  tags = {
    environment = "acctest"
  }
}

data "azurestack_managed_disk" "test" {
  name                = azurestack_managed_disk.test.name
  resource_group_name = azurestack_resource_group.test.name
}
`, resourceGroupName, data.Locations.Primary, name)
}
