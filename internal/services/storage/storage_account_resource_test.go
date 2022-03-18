package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type StorageAccountResource struct{}

func TestAccStorageAccount_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("account_tier").HasValue("Standard"),
				check.That(data.ResourceName).Key("account_replication_type").HasValue("LRS"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("production"),
			),
		},
		data.ImportStep(),
		// TODO this needs to be fixed
		/*{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("account_tier").HasValue("Standard"),
				check.That(data.ResourceName).Key("account_replication_type").HasValue("GRS"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("staging"),
			),
		},*/
	})
}

func TestAccStorageAccount_requiresImport(t *testing.T) {
	t.Skip("test skipped, please check comments inside this test")
	/*This test is skipped due to a bug in the Azure autorest client, when the Create request function is used,
	instead of using a POST request use sa PUT which causes the update of the previous resource instead of throwing
	an 'Already exists' error.
	*/
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		// data.RequiresImportErrorStep(r.requiresImport), // TODO: uncomment this until the bug gets resolved
	})
}

func TestAccStorageAccount_premium(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.premium(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("account_tier").HasValue("Premium"),
				check.That(data.ResourceName).Key("account_replication_type").HasValue("LRS"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("production"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStorageAccount_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccStorageAccount_blobConnectionString(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("primary_blob_connection_string").Exists(),
			),
		},
	})
}

func TestAccStorageAccount_NonStandardCasing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.nonStandardCasing(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config:             r.nonStandardCasing(data),
			PlanOnly:           true,
			ExpectNonEmptyPlan: false,
		},
	})
}

func (r StorageAccountResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.StorageAccountID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := client.Storage.AccountsClient.GetProperties(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return pointer.FromBool(false), nil
		}
		return nil, fmt.Errorf("retrieving Storage Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return pointer.FromBool(true), nil
}

func (r StorageAccountResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.StorageAccountID(state.ID)
	if err != nil {
		return nil, err
	}
	if _, err := client.Storage.AccountsClient.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		return nil, fmt.Errorf("deleting Storage Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return pointer.FromBool(true), nil
}

func (r StorageAccountResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                = "unlikely23exst2acct%s"
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

/*
func (r StorageAccountResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_account" "import" {
  name                     = azurestack_storage_account.test.name
  resource_group_name      = azurestack_storage_account.test.resource_group_name
  location                 = azurestack_storage_account.test.location
  account_tier             = azurestack_storage_account.test.account_tier
  account_replication_type = azurestack_storage_account.test.account_replication_type
}
`, template)
}
*/
func (r StorageAccountResource) premium(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurestack_resource_group.test.name

  location                 = azurestack_resource_group.test.location
  account_tier             = "Premium"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (r StorageAccountResource) nonStandardCasing(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                     = "unlikely23exst2acct%s"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "standard"
  account_replication_type = "lrs"

  tags = {
    environment = "production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func TestAccStorageAccount_enableHttpsTrafficOnly(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_account", "test")
	r := StorageAccountResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.enableHttpsTrafficOnly(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("enable_https_traffic_only").HasValue("true"),
			),
		},
		data.ImportStep(),
		{
			Config: r.enableHttpsTrafficOnlyDisabled(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("enable_https_traffic_only").HasValue("false"),
			),
		},
	})
}

func (r StorageAccountResource) enableHttpsTrafficOnly(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurestack_resource_group.test.name

  location                  = azurestack_resource_group.test.location
  account_tier              = "Standard"
  account_replication_type  = "LRS"
  enable_https_traffic_only = true

  tags = {
    environment = "production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (r StorageAccountResource) enableHttpsTrafficOnlyDisabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurestack_resource_group.test.name

  location                  = azurestack_resource_group.test.location
  account_tier              = "Standard"
  account_replication_type  = "LRS"
  enable_https_traffic_only = false

  tags = {
    environment = "production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}
