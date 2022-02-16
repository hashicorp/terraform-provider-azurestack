package storage_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
	"github.com/tombuildsstuff/giovanni/storage/2018-11-09/blob/blobs"
)

type StorageBlobResource struct{}

func TestAccStorageBlob_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.blockEmpty,
			TestResource: r,
		}),
	})
}

func TestAccStorageBlob_appendEmpty(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.appendEmpty(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "type"),
	})
}

func TestAccStorageBlob_blockEmpty(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockEmpty(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "type"),
	})
}

func TestAccStorageBlob_blockFromInlineContent(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockFromInlineContent(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "source_content", "type"),
	})
}

func TestAccStorageBlob_blockFromPublicBlob(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockFromPublicBlob(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "source_uri", "type"),
	})
}

func TestAccStorageBlob_blockFromPublicFile(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockFromPublicFile(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "source_uri", "type"),
	})
}

func TestAccStorageBlob_blockFromExistingBlob(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockFromExistingBlob(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "source_uri", "type"),
	})
}

func TestAccStorageBlob_blockFromLocalFile(t *testing.T) {
	sourceBlob, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Failed to create local source blob file")
	}

	if err := populateTempFile(sourceBlob); err != nil {
		t.Fatalf("Error populating temp file: %s", err)
	}
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockFromLocalBlob(data, sourceBlob.Name()),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				data.CheckWithClient(r.blobMatchesFile(blobs.BlockBlob, sourceBlob.Name())),
			),
		},
		data.ImportStep("parallelism", "size", "source", "type"),
	})
}

func TestAccStorageBlob_pageEmpty(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.pageEmpty(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "type"),
	})
}

func TestAccStorageBlob_pageEmptyPremium(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.pageEmptyPremium(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "type"),
	})
}

func TestAccStorageBlob_pageFromExistingBlob(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.pageFromExistingBlob(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("parallelism", "size", "type", "source_uri"),
	})
}

func TestAccStorageBlob_pageFromLocalFile(t *testing.T) {
	sourceBlob, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Failed to create local source blob file")
	}

	if err := populateTempFile(sourceBlob); err != nil {
		t.Fatalf("Error populating temp file: %s", err)
	}
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.pageFromLocalBlob(data, sourceBlob.Name()),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				data.CheckWithClient(r.blobMatchesFile(blobs.PageBlob, sourceBlob.Name())),
			),
		},
		data.ImportStep("parallelism", "size", "type", "source"),
	})
}

func TestAccStorageBlob_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_storage_blob", "test")
	r := StorageBlobResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.blockFromPublicBlob(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (r StorageBlobResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := blobs.ParseResourceID(state.ID)
	if err != nil {
		return nil, err
	}
	account, err := client.Storage.FindAccount(ctx, id.AccountName)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("unable to locate Account %q for Blob %q (Container %q)", id.AccountName, id.BlobName, id.ContainerName)
	}
	blobsClient, err := client.Storage.BlobsClient(ctx, *account)
	if err != nil {
		return nil, fmt.Errorf("building Blobs Client: %+v", err)
	}
	input := blobs.GetPropertiesInput{}
	resp, err := blobsClient.GetProperties(ctx, id.AccountName, id.ContainerName, id.BlobName, input)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return pointer.FromBool(false), nil
		}
		return nil, fmt.Errorf("retrieving Blob %q (Container %q / Account %q): %+v", id.BlobName, id.ContainerName, id.AccountName, err)
	}
	return pointer.FromBool(true), nil
}

func (r StorageBlobResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := blobs.ParseResourceID(state.ID)
	if err != nil {
		return nil, err
	}
	account, err := client.Storage.FindAccount(ctx, id.AccountName)
	if err != nil {
		return nil, fmt.Errorf("retrievign Account %q for Blob %q (Container %q): %+v", id.AccountName, id.BlobName, id.ContainerName, err)
	}
	blobsClient, err := client.Storage.BlobsClient(ctx, *account)
	if err != nil {
		return nil, fmt.Errorf("building Blobs Client: %+v", err)
	}
	input := blobs.DeleteInput{
		DeleteSnapshots: false,
	}
	if _, err := blobsClient.Delete(ctx, id.AccountName, id.ContainerName, id.BlobName, input); err != nil {
		return nil, fmt.Errorf("deleting Blob %q (Container %q / Account %q): %+v", id.BlobName, id.ContainerName, id.AccountName, err)
	}
	return pointer.FromBool(true), nil
}

func (r StorageBlobResource) blobMatchesFile(kind blobs.BlobType, filePath string) func(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) error {
	return func(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) error {
		name := state.Attributes["name"]
		containerName := state.Attributes["storage_container_name"]
		accountName := state.Attributes["storage_account_name"]

		account, err := clients.Storage.FindAccount(ctx, accountName)
		if err != nil {
			return fmt.Errorf("retrieving Account %q for Blob %q (Container %q): %s", accountName, name, containerName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", accountName)
		}

		client, err := clients.Storage.BlobsClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("building Blobs Client: %s", err)
		}

		// first check the type
		getPropsInput := blobs.GetPropertiesInput{}
		props, err := client.GetProperties(ctx, accountName, containerName, name, getPropsInput)
		if err != nil {
			return fmt.Errorf("retrieving Properties for Blob %q (Container %q): %s", name, containerName, err)
		}

		if props.BlobType != kind {
			return fmt.Errorf("Bad: blob type %q does not match expected type %q", props.BlobType, kind)
		}

		// then compare the content itself
		getInput := blobs.GetInput{}
		actualProps, err := client.Get(ctx, accountName, containerName, name, getInput)
		if err != nil {
			return fmt.Errorf("retrieving Blob %q (Container %q): %s", name, containerName, err)
		}

		actualContents := actualProps.Contents

		// local file for comparison
		expectedContents, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		if string(actualContents) != string(expectedContents) {
			return fmt.Errorf("Bad: Storage Blob %q (storage container: %q) does not match contents", name, containerName)
		}

		return nil
	}
}

func (r StorageBlobResource) appendEmpty(data acceptance.TestData) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
provider "azurestack" {}

%s

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Append"
}
`, template)
}

func (r StorageBlobResource) blockEmpty(data acceptance.TestData) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
provider "azurestack" {}

%s

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
}
`, template)
}

func (r StorageBlobResource) blockFromInlineContent(data acceptance.TestData) string {
	template := r.template(data, "blob")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "test" {
  name                   = "rick.morty"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
  source_content         = "Wubba Lubba Dub Dub"
}
`, template)
}

func (r StorageBlobResource) blockFromPublicBlob(data acceptance.TestData) string {
	template := r.template(data, "blob")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "source" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
  source_uri             = "http://old-releases.ubuntu.com/releases/bionic/ubuntu-18.04-desktop-amd64.iso"
}

resource "azurestack_storage_container" "second" {
  name                  = "second"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}

resource "azurestack_storage_blob" "test" {
  name                   = "copied.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.second.name
  type                   = "Block"
  source_uri             = azurestack_storage_blob.source.id
}
`, template)
}

func (r StorageBlobResource) blockFromPublicFile(data acceptance.TestData) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
  source_uri             = "http://old-releases.ubuntu.com/releases/bionic/ubuntu-18.04-desktop-amd64.iso"
}
`, template)
}

func (r StorageBlobResource) blockFromExistingBlob(data acceptance.TestData) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "source" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
  source_uri             = "http://old-releases.ubuntu.com/releases/bionic/ubuntu-18.04-desktop-amd64.iso"
}

resource "azurestack_storage_blob" "test" {
  name                   = "copied.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
  source_uri             = azurestack_storage_blob.source.id
}
`, template)
}

func (r StorageBlobResource) blockFromLocalBlob(data acceptance.TestData, fileName string) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Block"
  source                 = "%s"
}
`, template, fileName)
}

func (r StorageBlobResource) pageEmpty(data acceptance.TestData) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Page"
  size                   = 5120
}
`, template)
}

func (r StorageBlobResource) pageEmptyPremium(data acceptance.TestData) string {
	template := r.templatePremium(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Page"
  size                   = 5120
}
`, template)
}

func (r StorageBlobResource) pageFromExistingBlob(data acceptance.TestData) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "source" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Page"
  size                   = 5120
}

resource "azurestack_storage_blob" "test" {
  name                   = "copied.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Page"
  source_uri             = azurestack_storage_blob.source.id
}
`, template)
}

func (r StorageBlobResource) pageFromLocalBlob(data acceptance.TestData, fileName string) string {
	template := r.template(data, "private")
	return fmt.Sprintf(`
%s

provider "azurestack" {
  features {}
}

resource "azurestack_storage_blob" "test" {
  name                   = "example.vhd"
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name
  type                   = "Page"
  source                 = "%s"
}
`, template, fileName)
}

func (r StorageBlobResource) requiresImport(data acceptance.TestData) string {
	template := r.blockFromPublicBlob(data)
	return fmt.Sprintf(`
%s

resource "azurestack_storage_blob" "import" {
  name                   = azurestack_storage_blob.test.name
  storage_account_name   = azurestack_storage_blob.test.storage_account_name
  storage_container_name = azurestack_storage_blob.test.storage_container_name
  type                   = azurestack_storage_blob.test.type
  size                   = azurestack_storage_blob.test.size
}
`, template)
}

func (r StorageBlobResource) template(data acceptance.TestData, accessLevel string) string {
	return fmt.Sprintf(`
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
}

resource "azurestack_storage_container" "test" {
  name                  = "test"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, accessLevel)
}

func (r StorageBlobResource) templatePremium(data acceptance.TestData, accessLevel string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Premium"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "test"
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, accessLevel)
}

func populateTempFile(input *os.File) error {
	if err := input.Truncate(25*1024*1024 + 512); err != nil {
		return fmt.Errorf("Failed to truncate file to 25M")
	}

	for i := int64(0); i < 20; i += 2 {
		randomBytes := make([]byte, 1*1024*1024)
		if _, err := rand.Read(randomBytes); err != nil {
			return fmt.Errorf("Failed to read random bytes")
		}

		if _, err := input.WriteAt(randomBytes, i*1024*1024); err != nil {
			return fmt.Errorf("Failed to write random bytes to file")
		}
	}

	randomBytes := make([]byte, 5*1024*1024)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("Failed to read random bytes")
	}

	if _, err := input.WriteAt(randomBytes, 20*1024*1024); err != nil {
		return fmt.Errorf("Failed to write random bytes to file")
	}

	if err := input.Close(); err != nil {
		return fmt.Errorf("Failed to close source blob")
	}

	return nil
}
