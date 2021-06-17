package datalake_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/datalake"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type DataLakeStoreFileResource struct {
}

func TestAccDataLakeStoreFile_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store_file", "test")
	r := DataLakeStoreFileResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("local_file_path"),
	})
}

func TestAccDataLakeStoreFile_largefiles(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store_file", "test")
	r := DataLakeStoreFileResource{}

	// "large" in this context is anything greater than 4 megabytes
	largeSize := 12 * 1024 * 1024 // 12 mb
	bytes := make([]byte, largeSize)
	rand.Read(bytes) // fill with random data

	tmpfile, err := os.CreateTemp("", "azurerm-acc-datalake-file-large")
	if err != nil {
		t.Errorf("Unable to open a temporary file.")
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(bytes); err != nil {
		t.Errorf("Unable to write to temporary file %q: %v", tmpfile.Name(), err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("Unable to close temporary file %q: %v", tmpfile.Name(), err)
	}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.largefiles(data, tmpfile.Name()),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("local_file_path"),
	})
}

func TestAccDataLakeStoreFile_requiresimport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store_file", "test")
	r := DataLakeStoreFileResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_data_lake_store_file"),
		},
	})
}

func (t DataLakeStoreFileResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	client := clients.Datalake.StoreFilesClient
	id, err := datalake.ParseDataLakeStoreFileId(state.ID, client.AdlsFileSystemDNSSuffix)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetFileStatus(ctx, id.StorageAccountName, id.FilePath, utils.Bool(true))
	if err != nil {
		return nil, fmt.Errorf("retrieving Date Lake Store File Rule %q (Account %q): %v", id.FilePath, id.StorageAccountName, err)
	}

	return utils.Bool(resp.FileStatus != nil), nil
}

func (DataLakeStoreFileResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = "%s"
  firewall_state      = "Disabled"
}

resource "azurerm_data_lake_store_file" "test" {
  remote_file_path = "/test/application_gateway_test.cer"
  account_name     = azurerm_data_lake_store.test.name
  local_file_path  = "./testdata/application_gateway_test.cer"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.Locations.Primary)
}

func (DataLakeStoreFileResource) largefiles(data acceptance.TestData, file string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-datalake-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = "%s"
  firewall_state      = "Disabled"
}

resource "azurerm_data_lake_store_file" "test" {
  remote_file_path = "/test/testAccAzureRMDataLakeStoreFile_largefiles.bin"
  account_name     = azurerm_data_lake_store.test.name
  local_file_path  = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.Locations.Primary, file)
}

func (DataLakeStoreFileResource) requiresImport(data acceptance.TestData) string {
	template := DataLakeStoreFileResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_data_lake_store_file" "import" {
  remote_file_path = azurerm_data_lake_store_file.test.remote_file_path
  account_name     = azurerm_data_lake_store_file.test.account_name
  local_file_path  = "./testdata/application_gateway_test.cer"
}
`, template)
}
