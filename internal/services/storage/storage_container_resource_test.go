package storage_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type StorageContainerResource struct{}

func TestAccStorageContainer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

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

func TestAccStorageContainer_deleteAndRecreate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.template(data),
		},
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStorageContainer_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccStorageContainer_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.update(data, "private"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("container_access_type").HasValue("private"),
			),
		},
		{
			Config: r.update(data, "container"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("container_access_type").HasValue("container"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStorageContainer_metaData(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.metaData(data, "private"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.metaDataUpdated(data, "private"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.metaDataEmpty(data, "private"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStorageContainer_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccStorageContainer_root(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.root(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").HasValue("$root"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStorageContainer_web(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_container", "test")
	r := StorageContainerResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.web(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").HasValue("$web"),
			),
		},
		data.ImportStep(),
	})
}

func (r StorageContainerResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.StorageContainerDataPlaneID(state.ID)
	if err != nil {
		return nil, err
	}
	account, err := client.Storage.FindAccount(ctx, id.AccountName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Account %q for Container %q: %+v", id.AccountName, id.Name, err)
	}
	if account == nil {
		return nil, fmt.Errorf("unable to locate Storage Account %q", id.AccountName)
	}

	containersClient, err := client.Storage.ContainersClient(ctx, *account)
	if err != nil {
		return nil, fmt.Errorf("building Containers Client: %+v", err)
	}
	prop, err := containersClient.Get(ctx, account.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Container %q (Account %q / Resource Group %q): %+v", id.Name, id.AccountName, account.ResourceGroup, err)
	}
	return pointer.FromBool(prop != nil), nil
}

func (r StorageContainerResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.StorageContainerDataPlaneID(state.ID)
	if err != nil {
		return nil, err
	}
	account, err := client.Storage.FindAccount(ctx, id.AccountName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Account %q for Container %q: %+v", id.AccountName, id.Name, err)
	}
	if account == nil {
		return nil, fmt.Errorf("unable to locate Storage Account %q", id.AccountName)
	}
	containersClient, err := client.Storage.ContainersClient(ctx, *account)
	if err != nil {
		return nil, fmt.Errorf("building Containers Client: %+v", err)
	}
	if err := containersClient.Delete(ctx, account.ResourceGroup, id.AccountName, id.Name); err != nil {
		return nil, fmt.Errorf("deleting Container %q (Account %q / Resource Group %q): %+v", id.Name, id.AccountName, account.ResourceGroup, err)
	}
	return pointer.FromBool(true), nil
}

func (r StorageContainerResource) basic(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}
`, template)
}

func (r StorageContainerResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_container" "import" {
  name                  = azurestack_storage_container.test.name
  storage_account_name  = azurestack_storage_container.test.storage_account_name
  container_access_type = azurestack_storage_container.test.container_access_type
}
`, template)
}

func (r StorageContainerResource) update(data acceptance.TestData, accessType string) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "%s"
}
`, template, accessType)
}

func (r StorageContainerResource) metaData(data acceptance.TestData, accessType string) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "%s"
  metadata = {
    hello = "world"
  }
}
`, template, accessType)
}

func (r StorageContainerResource) metaDataUpdated(data acceptance.TestData, accessType string) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s
resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "%s"
  metadata = {
    hello = "world"
    panda = "pops"
  }
}
`, template, accessType)
}

func (r StorageContainerResource) metaDataEmpty(data acceptance.TestData, accessType string) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s
resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "%s"
  metadata              = {}
}
`, template, accessType)
}

func (r StorageContainerResource) root(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_container" "test" {
  name                  = "$root"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}
`, template)
}

func (r StorageContainerResource) web(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_container" "test" {
  name                  = "$web"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}
`, template)
}

func (r StorageContainerResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func TestValidateStorageContainerName(t *testing.T) {
	validNames := []string{
		"valid-name",
		"valid02-name",
		"$root",
		"$web",
	}
	for _, v := range validNames {
		_, errors := validate.StorageContainerName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Storage Container Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"InvalidName1",
		"-invalidname1",
		"invalid_name",
		"invalid!",
		"ww",
		"$notroot",
		"$notweb",
		strings.Repeat("w", 65),
	}
	for _, v := range invalidNames {
		_, errors := validate.StorageContainerName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Storage Container Name", v)
		}
	}
}
