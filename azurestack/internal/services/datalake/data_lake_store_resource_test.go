package datalake_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type DataLakeStoreResource struct {
}

func TestAccDataLakeStore_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store", "test")
	r := DataLakeStoreResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tier").HasValue("Consumption"),
				check.That(data.ResourceName).Key("encryption_state").HasValue("Enabled"),
				check.That(data.ResourceName).Key("encryption_type").HasValue("ServiceManaged"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDataLakeStore_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store", "test")
	r := DataLakeStoreResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_data_lake_store"),
		},
	})
}

func TestAccDataLakeStore_tier(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store", "test")
	r := DataLakeStoreResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.tier(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tier").HasValue("Commitment_1TB"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDataLakeStore_encryptionDisabled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store", "test")
	r := DataLakeStoreResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.encryptionDisabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("encryption_state").HasValue("Disabled"),
				check.That(data.ResourceName).Key("encryption_type").HasValue(""),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDataLakeStore_firewallUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store", "test")
	r := DataLakeStoreResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.firewall(data, "Enabled", "Enabled"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("firewall_state").HasValue("Enabled"),
				check.That(data.ResourceName).Key("firewall_allow_azure_ips").HasValue("Enabled"),
			),
		},
		{
			Config: r.firewall(data, "Enabled", "Disabled"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("firewall_state").HasValue("Enabled"),
				check.That(data.ResourceName).Key("firewall_allow_azure_ips").HasValue("Disabled"),
			),
		},
		{
			Config: r.firewall(data, "Disabled", "Enabled"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("firewall_state").HasValue("Disabled"),
				check.That(data.ResourceName).Key("firewall_allow_azure_ips").HasValue("Enabled"),
			),
		},
		{
			Config: r.firewall(data, "Disabled", "Disabled"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("firewall_state").HasValue("Disabled"),
				check.That(data.ResourceName).Key("firewall_allow_azure_ips").HasValue("Disabled"),
			),
		},
	})
}

func TestAccDataLakeStore_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store", "test")
	r := DataLakeStoreResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
			),
		},
		{
			Config: r.withTagsUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func (t DataLakeStoreResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := azure.ParseAzureResourceID(state.ID)
	if err != nil {
		return nil, err
	}

	name := id.Path["accounts"]

	resp, err := clients.Datalake.StoreAccountsClient.Get(ctx, id.ResourceGroup, name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Date Lake Store %q (resource group: %q): %+v", name, id.ResourceGroup, err)
	}

	return utils.Bool(resp.DataLakeStoreAccountProperties != nil), nil
}

func (DataLakeStoreResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "acctest%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}
`, data.RandomInteger, data.Locations.Primary, strconv.Itoa(data.RandomInteger)[2:17])
}

func (DataLakeStoreResource) requiresImport(data acceptance.TestData) string {
	template := DataLakeStoreResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_data_lake_store" "import" {
  name                = azurerm_data_lake_store.test.name
  resource_group_name = azurerm_data_lake_store.test.resource_group_name
  location            = azurerm_data_lake_store.test.location
}
`, template)
}

func (DataLakeStoreResource) tier(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "acctest%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  tier                = "Commitment_1TB"
}
`, data.RandomInteger, data.Locations.Primary, strconv.Itoa(data.RandomInteger)[2:17])
}

func (DataLakeStoreResource) encryptionDisabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "acctest%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  encryption_state    = "Disabled"
}
`, data.RandomInteger, data.Locations.Primary, strconv.Itoa(data.RandomInteger)[2:17])
}

func (DataLakeStoreResource) firewall(data acceptance.TestData, firewallState string, firewallAllowAzureIPs string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                     = "acctest%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  firewall_state           = "%s"
  firewall_allow_azure_ips = "%s"
}
`, data.RandomInteger, data.Locations.Primary, strconv.Itoa(data.RandomInteger)[2:17], firewallState, firewallAllowAzureIPs)
}

func (DataLakeStoreResource) withTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "acctest%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, strconv.Itoa(data.RandomInteger)[2:17])
}

func (DataLakeStoreResource) withTagsUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "acctest%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, strconv.Itoa(data.RandomInteger)[2:17])
}
