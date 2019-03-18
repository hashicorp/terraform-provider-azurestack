package azurestack

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"strings"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceAzureStackStorageBlobType_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "unknown",
			ErrCount: 1,
		},
		{
			Value:    "page",
			ErrCount: 0,
		},
		{
			Value:    "block",
			ErrCount: 0,
		},
		{
			Value:    "BLOCK",
			ErrCount: 0,
		},
		{
			Value:    "Block",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateArmStorageBlobType(tc.Value, "azurestack_storage_blob")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Storage Blob type to trigger a validation error")
		}
	}
}

func TestResourceAzureStackStorageBlobSize_validation(t *testing.T) {
	cases := []struct {
		Value    int
		ErrCount int
	}{
		{
			Value:    511,
			ErrCount: 1,
		},
		{
			Value:    512,
			ErrCount: 0,
		},
		{
			Value:    1024,
			ErrCount: 0,
		},
		{
			Value:    2048,
			ErrCount: 0,
		},
		{
			Value:    5120,
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateArmStorageBlobSize(tc.Value, "azurestack_storage_blob")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Storage Blob size to trigger a validation error")
		}
	}
}

func TestResourceAzureStackStorageBlobParallelism_validation(t *testing.T) {
	cases := []struct {
		Value    int
		ErrCount int
	}{
		{
			Value:    1,
			ErrCount: 0,
		},
		{
			Value:    0,
			ErrCount: 1,
		},
		{
			Value:    -1,
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateArmStorageBlobParallelism(tc.Value, "azurestack_storage_blob")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Storage Blob parallelism to trigger a validation error")
		}
	}
}

func TestResourceAzureStackStorageBlobAttempts_validation(t *testing.T) {
	cases := []struct {
		Value    int
		ErrCount int
	}{
		{
			Value:    1,
			ErrCount: 0,
		},
		{
			Value:    0,
			ErrCount: 1,
		},
		{
			Value:    -1,
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateArmStorageBlobAttempts(tc.Value, "azurestack_storage_blob")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Storage Blob attempts to trigger a validation error")
		}
	}
}

func TestAccAzureStackStorageBlob_basic(t *testing.T) {
	ri := acctest.RandInt()
	rs := strings.ToLower(acctest.RandString(11))
	config := testAccAzureStackStorageBlob_basic(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageBlobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageBlobExists("azurestack_storage_blob.test"),
				),
			},
		},
	})
}

func TestAccAzureStackStorageBlob_disappears(t *testing.T) {
	ri := acctest.RandInt()
	rs := strings.ToLower(acctest.RandString(11))
	config := testAccAzureStackStorageBlob_basic(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageBlobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageBlobExists("azurestack_storage_blob.test"),
					testCheckAzureStackStorageBlobDisappears("azurestack_storage_blob.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackStorageBlobBlock_source(t *testing.T) {
	ri := acctest.RandInt()
	rs1 := strings.ToLower(acctest.RandString(11))
	sourceBlob, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Failed to create local source blob file")
	}

	_, err = io.CopyN(sourceBlob, rand.Reader, 25*1024*1024)
	if err != nil {
		t.Fatalf("Failed to write random test to source blob")
	}

	err = sourceBlob.Close()
	if err != nil {
		t.Fatalf("Failed to close source blob")
	}

	config := testAccAzureStackStorageBlobBlock_source(ri, rs1, sourceBlob.Name(), testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageBlobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageBlobMatchesFile("azurestack_storage_blob.source", storage.BlobTypeBlock, sourceBlob.Name()),
				),
			},
		},
	})
}

func TestAccAzureStackStorageBlobPage_source(t *testing.T) {
	ri := acctest.RandInt()
	rs := strings.ToLower(acctest.RandString(11))
	sourceBlob, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Failed to create local source blob file")
	}

	err = sourceBlob.Truncate(25*1024*1024 + 512)
	if err != nil {
		t.Fatalf("Failed to truncate file to 25M")
	}

	for i := int64(0); i < 20; i = i + 2 {
		randomBytes := make([]byte, 1*1024*1024)
		_, err = rand.Read(randomBytes)
		if err != nil {
			t.Fatalf("Failed to read random bytes")
		}

		_, err = sourceBlob.WriteAt(randomBytes, i*1024*1024)
		if err != nil {
			t.Fatalf("Failed to write random bytes to file")
		}
	}

	randomBytes := make([]byte, 5*1024*1024)
	_, err = rand.Read(randomBytes)
	if err != nil {
		t.Fatalf("Failed to read random bytes")
	}

	_, err = sourceBlob.WriteAt(randomBytes, 20*1024*1024)
	if err != nil {
		t.Fatalf("Failed to write random bytes to file")
	}

	err = sourceBlob.Close()
	if err != nil {
		t.Fatalf("Failed to close source blob")
	}

	config := testAccAzureStackStorageBlobPage_source(ri, rs, sourceBlob.Name(), testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageBlobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageBlobMatchesFile("azurestack_storage_blob.source", storage.BlobTypePage, sourceBlob.Name()),
				),
			},
		},
	})
}

func TestAccAzureStackStorageBlob_source_uri(t *testing.T) {
	ri := acctest.RandInt()
	rs := strings.ToLower(acctest.RandString(11))
	sourceBlob, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Failed to create local source blob file")
	}

	_, err = io.CopyN(sourceBlob, rand.Reader, 25*1024*1024)
	if err != nil {
		t.Fatalf("Failed to write random test to source blob")
	}

	err = sourceBlob.Close()
	if err != nil {
		t.Fatalf("Failed to close source blob")
	}

	config := testAccAzureStackStorageBlob_source_uri(ri, rs, sourceBlob.Name(), testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageBlobDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageBlobMatchesFile("azurestack_storage_blob.destination", storage.BlobTypeBlock, sourceBlob.Name()),
				),
			},
		},
	})
}

func testCheckAzureStackStorageBlobExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		storageAccountName := rs.Primary.Attributes["storage_account_name"]
		storageContainerName := rs.Primary.Attributes["storage_container_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for storage blob: %s", name)
		}

		armClient := testAccProvider.Meta().(*ArmClient)
		ctx := armClient.StopContext
		blobClient, accountExists, err := armClient.getBlobStorageClientForStorageAccount(ctx, resourceGroup, storageAccountName)
		if err != nil {
			return err
		}
		if !accountExists {
			return fmt.Errorf("Bad: Storage Account %q does not exist", storageAccountName)
		}

		container := blobClient.GetContainerReference(storageContainerName)
		blob := container.GetBlobReference(name)
		exists, err := blob.Exists()
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Bad: Storage Blob %q (storage container: %q) does not exist", name, storageContainerName)
		}

		return nil
	}
}

func testCheckAzureStackStorageBlobDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		storageAccountName := rs.Primary.Attributes["storage_account_name"]
		storageContainerName := rs.Primary.Attributes["storage_container_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for storage blob: %s", name)
		}

		armClient := testAccProvider.Meta().(*ArmClient)
		ctx := armClient.StopContext
		blobClient, accountExists, err := armClient.getBlobStorageClientForStorageAccount(ctx, resourceGroup, storageAccountName)
		if err != nil {
			return err
		}
		if !accountExists {
			return fmt.Errorf("Bad: Storage Account %q does not exist", storageAccountName)
		}

		container := blobClient.GetContainerReference(storageContainerName)
		blob := container.GetBlobReference(name)
		options := &storage.DeleteBlobOptions{}
		_, err = blob.DeleteIfExists(options)
		return err
	}
}

func testCheckAzureStackStorageBlobMatchesFile(name string, kind storage.BlobType, filePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		storageAccountName := rs.Primary.Attributes["storage_account_name"]
		storageContainerName := rs.Primary.Attributes["storage_container_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for storage blob: %s", name)
		}

		armClient := testAccProvider.Meta().(*ArmClient)
		ctx := armClient.StopContext
		blobClient, accountExists, err := armClient.getBlobStorageClientForStorageAccount(ctx, resourceGroup, storageAccountName)
		if err != nil {
			return err
		}
		if !accountExists {
			return fmt.Errorf("Bad: Storage Account %q does not exist", storageAccountName)
		}

		containerReference := blobClient.GetContainerReference(storageContainerName)
		blobReference := containerReference.GetBlobReference(name)
		propertyOptions := &storage.GetBlobPropertiesOptions{}
		err = blobReference.GetProperties(propertyOptions)
		if err != nil {
			return err
		}

		properties := blobReference.Properties

		if properties.BlobType != kind {
			return fmt.Errorf("Bad: blob type %q does not match expected type %q", properties.BlobType, kind)
		}

		getOptions := &storage.GetBlobOptions{}
		blob, err := blobReference.Get(getOptions)
		if err != nil {
			return err
		}

		contents, err := ioutil.ReadAll(blob)
		if err != nil {
			return err
		}
		defer blob.Close()

		expectedContents, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		if string(contents) != string(expectedContents) {
			return fmt.Errorf("Bad: Storage Blob %q (storage container: %q) does not match contents", name, storageContainerName)
		}

		return nil
	}
}

func testCheckAzureStackStorageBlobDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_storage_blob" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		storageAccountName := rs.Primary.Attributes["storage_account_name"]
		storageContainerName := rs.Primary.Attributes["storage_container_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for storage blob: %s", name)
		}

		armClient := testAccProvider.Meta().(*ArmClient)
		ctx := armClient.StopContext
		blobClient, accountExists, err := armClient.getBlobStorageClientForStorageAccount(ctx, resourceGroup, storageAccountName)
		if err != nil {
			return nil
		}
		if !accountExists {
			return nil
		}

		container := blobClient.GetContainerReference(storageContainerName)
		blob := container.GetBlobReference(name)
		exists, err := blob.Exists()
		if err != nil {
			return nil
		}

		if exists {
			return fmt.Errorf("Bad: Storage Blob %q (storage container: %q) still exists", name, storageContainerName)
		}
	}

	return nil
}

func testAccAzureStackStorageBlob_basic(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_storage_blob" "test" {
  name = "herpderp1.vhd"

  resource_group_name    = "${azurestack_resource_group.test.name}"
  storage_account_name   = "${azurestack_storage_account.test.name}"
  storage_container_name = "${azurestack_storage_container.test.name}"

  type = "page"
  size = 5120
}
`, rInt, location, rString)
}

func testAccAzureStackStorageBlobBlock_source(rInt int, rString string, sourceBlobName string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "source" {
  name                     = "acctestacc%s"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "source" {
  name                  = "source"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.source.name}"
  container_access_type = "blob"
}

resource "azurestack_storage_blob" "source" {
  name = "source.vhd"

  resource_group_name    = "${azurestack_resource_group.test.name}"
  storage_account_name   = "${azurestack_storage_account.source.name}"
  storage_container_name = "${azurestack_storage_container.source.name}"

  type        = "block"
  source      = "%s"
  parallelism = 4
  attempts    = 2
}
`, rInt, location, rString, sourceBlobName)
}

func testAccAzureStackStorageBlobPage_source(rInt int, rString string, sourceBlobName string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "source" {
  name                     = "acctestacc%s"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "source" {
  name                  = "source"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.source.name}"
  container_access_type = "blob"
}

resource "azurestack_storage_blob" "source" {
  name = "source.vhd"

  resource_group_name    = "${azurestack_resource_group.test.name}"
  storage_account_name   = "${azurestack_storage_account.source.name}"
  storage_container_name = "${azurestack_storage_container.source.name}"

  type        = "page"
  source      = "%s"
  parallelism = 3
  attempts    = 3
}
`, rInt, location, rString, sourceBlobName)
}

func testAccAzureStackStorageBlob_source_uri(rInt int, rString string, sourceBlobName string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "source" {
  name                     = "acctestacc%s"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "source" {
  name                  = "source"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.source.name}"
  container_access_type = "blob"
}

resource "azurestack_storage_blob" "source" {
  name = "source.vhd"

  resource_group_name    = "${azurestack_resource_group.test.name}"
  storage_account_name   = "${azurestack_storage_account.source.name}"
  storage_container_name = "${azurestack_storage_container.source.name}"

  type        = "block"
  source      = "%s"
  parallelism = 4
  attempts    = 2
}

resource "azurestack_storage_blob" "destination" {
  name                   = "destination.vhd"
  resource_group_name    = "${azurestack_resource_group.test.name}"
  storage_account_name   = "${azurestack_storage_account.source.name}"
  storage_container_name = "${azurestack_storage_container.source.name}"
  source_uri             = "${azurestack_storage_blob.source.url}"
}
`, rInt, location, rString, sourceBlobName)
}
