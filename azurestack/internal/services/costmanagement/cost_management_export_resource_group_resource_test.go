package costmanagement_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/costmanagement/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type CostManagementExportResourceGroupResource struct {
}

func TestAccCostManagementExportResourceGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cost_management_export_resource_group", "test")
	r := CostManagementExportResourceGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccCostManagementExportResourceGroup_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cost_management_export_resource_group", "test")
	r := CostManagementExportResourceGroupResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t CostManagementExportResourceGroupResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.CostManagementExportResourceGroupID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.CostManagement.ExportClient.Get(ctx, id.ResourceId, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Cost Management Export ResourceGroup %q (resource group: %q) does not exist", id.Name, id.ResourceId)
	}

	return utils.Bool(resp.ExportProperties != nil), nil
}

func (CostManagementExportResourceGroupResource) basic(data acceptance.TestData) string {
	start := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	end := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-cm-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurerm_resource_group.test.name

  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_cost_management_export_resource_group" "test" {
  name                    = "accrg%d"
  resource_group_id       = azurerm_resource_group.test.id
  recurrence_type         = "Monthly"
  recurrence_period_start = "%sT00:00:00Z"
  recurrence_period_end   = "%sT00:00:00Z"

  delivery_info {
    storage_account_id = azurerm_storage_account.test.id
    container_name     = "acctestcontainer"
    root_folder_path   = "/root"
  }

  query {
    type       = "Usage"
    time_frame = "TheLastMonth"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, start, end)
}

func (CostManagementExportResourceGroupResource) update(data acceptance.TestData) string {
	start := time.Now().AddDate(0, 3, 0).Format("2006-01-02")
	end := time.Now().AddDate(0, 4, 0).Format("2006-01-02")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-cm-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurerm_resource_group.test.name

  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_cost_management_export_resource_group" "test" {
  name                    = "accrg%d"
  resource_group_id       = azurerm_resource_group.test.id
  recurrence_type         = "Monthly"
  recurrence_period_start = "%sT00:00:00Z"
  recurrence_period_end   = "%sT00:00:00Z"

  delivery_info {
    storage_account_id = azurerm_storage_account.test.id
    container_name     = "acctestcontainer"
    root_folder_path   = "/root/updated"
  }

  query {
    type       = "Usage"
    time_frame = "WeekToDate"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomInteger, start, end)
}
