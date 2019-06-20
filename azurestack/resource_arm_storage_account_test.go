package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestValidateArmStorageAccountType(t *testing.T) {
	testCases := []struct {
		input       string
		shouldError bool
	}{
		{"standard_lrs", false},
		{"invalid", true},
	}

	for _, test := range testCases {
		_, es := validateArmStorageAccountType(test.input, "account_type")

		if test.shouldError && len(es) == 0 {
			t.Fatalf("Expected validating account_type %q to fail", test.input)
		}
	}
}

func TestValidateArmStorageAccountName(t *testing.T) {
	testCases := []struct {
		input       string
		shouldError bool
	}{
		{"ab", true},
		{"ABC", true},
		{"abc", false},
		{"123456789012345678901234", false},
		{"1234567890123456789012345", true},
		{"abc12345", false},
	}

	for _, test := range testCases {
		_, es := validateArmStorageAccountName(test.input, "name")

		if test.shouldError && len(es) == 0 {
			t.Fatalf("Expected validating name %q to fail", test.input)
		}
	}
}

// Update is commented due to:
// Property AccountType that cannot be updated for the
// storage account was specified in the request."

func TestAccAzureStackStorageAccount_basic(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_basic(ri, rs, location)
	// postConfig := testAccAzureStackStorageAccount_update(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_tier", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "account_replication_type", "LRS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "production"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// {
			// 	Config: postConfig,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testCheckAzureStackStorageAccountExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "account_tier", "Standard"),
			// 		resource.TestCheckResourceAttr(resourceName, "account_replication_type", "GRS"),
			// 		resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "tags.environment", "staging"),
			// 	),
			// },
		},
	})
}

func TestAccAzureStackStorageAccount_premium(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_premium(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_tier", "Premium"),
					resource.TestCheckResourceAttr(resourceName, "account_replication_type", "LRS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "production"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackStorageAccount_disappears(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	preConfig := testAccAzureStackStorageAccount_basic(ri, rs, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_tier", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "account_replication_type", "LRS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "production"),
					testCheckAzureStackStorageAccountDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureStackStorageAccount_blobConnectionString(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	preConfig := testAccAzureStackStorageAccount_basic(ri, rs, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "primary_blob_connection_string"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Skipped due to error:
// Data Encryption is not supported to disable in the specified location.

func TestAccAzureStackStorageAccount_blobEncryption(t *testing.T) {
	t.Skip()
	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_blobEncryption(ri, rs, location)
	postConfig := testAccAzureStackStorageAccount_blobEncryptionDisabled(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_blob_encryption", "true"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_blob_encryption", "false"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// File Encryption is not supported in the profile
func TestAccAzureStackStorageAccount_fileEncryption(t *testing.T) {
	t.Skip()

	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_fileEncryption(ri, rs, location)
	postConfig := testAccAzureStackStorageAccount_fileEncryptionDisabled(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_file_encryption", "true"),
				),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_file_encryption", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// HTTPS traffic not supported by the profile
func TestAccAzureStackStorageAccount_enableHttpsTrafficOnly(t *testing.T) {
	t.Skip()

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_enableHttpsTrafficOnly(ri, rs, location)
	postConfig := testAccAzureStackStorageAccount_enableHttpsTrafficOnlyDisabled(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists("azurestack_storage_account.testsa"),
					resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "enable_https_traffic_only", "true"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists("azurestack_storage_account.testsa"),
					resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "enable_https_traffic_only", "false"),
				),
			},
		},
	})
}

// Skip due to limitations of the account, if BlobStorage is specified as an argument
// It returns an error 400, BlobStorage is disabled in my account too when
// I try to create an account
func TestAccAzureStackStorageAccount_blobStorageWithUpdate(t *testing.T) {
	t.Skip()

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_blobStorage(ri, rs, location)
	postConfig := testAccAzureStackStorageAccount_blobStorageUpdate(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists("azurestack_storage_account.testsa"),
					resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "account_kind", "BlobStorage"),
					//  resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "access_tier", "Hot"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists("azurestack_storage_account.testsa"),
					// resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "access_tier", "Cool"),
				),
			},
		},
	})
}

// V2 not supported by the account
func TestAccAzureStackStorageAccount_storageV2WithUpdate(t *testing.T) {

	t.Skip()

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	location := testLocation()
	preConfig := testAccAzureStackStorageAccount_storageV2(ri, rs, location)
	postConfig := testAccAzureStackStorageAccount_storageV2Update(ri, rs, location)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists("azurestack_storage_account.testsa"),
					resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "account_kind", "StorageV2"),
					// resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "access_tier", "Hot"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists("azurestack_storage_account.testsa"),
					// resource.TestCheckResourceAttr("azurestack_storage_account.testsa", "access_tier", "Cool"),
				),
			},
		},
	})
}

func TestAccAzureStackStorageAccount_NonStandardCasing(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	preConfig := testAccAzureStackStorageAccount_nonStandardCasing(ri, rs, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackStorageAccountExists(resourceName),
				),
			},

			{
				Config:             preConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckAzureStackStorageAccountExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		storageAccount := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		// Ensure resource group exists in API
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		conn := testAccProvider.Meta().(*ArmClient).storageServiceClient

		resp, err := conn.GetProperties(ctx, resourceGroup, storageAccount)
		if err != nil {
			return fmt.Errorf("Bad: Get on storageServiceClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: StorageAccount %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureStackStorageAccountDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		storageAccount := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		// Ensure resource group exists in API
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		conn := testAccProvider.Meta().(*ArmClient).storageServiceClient

		_, err := conn.Delete(ctx, resourceGroup, storageAccount)
		if err != nil {
			return fmt.Errorf("Bad: Delete on storageServiceClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureStackStorageAccountDestroy(s *terraform.State) error {
	ctx := testAccProvider.Meta().(*ArmClient).StopContext
	conn := testAccProvider.Meta().(*ArmClient).storageServiceClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_storage_account" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.GetProperties(ctx, resourceGroup, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Storage Account still exists:\n%#v", resp.AccountProperties)
		}
	}

	return nil
}

func testAccAzureStackStorageAccount_basic(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_premium(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "Premium"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_blobEncryption(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  enable_blob_encryption   = true

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_blobEncryptionDisabled(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  enable_blob_encryption   = false

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_fileEncryption(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  enable_file_encryption   = true

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_fileEncryptionDisabled(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  enable_file_encryption   = false

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_enableHttpsTrafficOnly(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                  = "${azurestack_resource_group.testrg.location}"
  account_tier              = "Standard"
  account_replication_type  = "LRS"
  enable_https_traffic_only = true

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_enableHttpsTrafficOnlyDisabled(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                  = "${azurestack_resource_group.testrg.location}"
  account_tier              = "Standard"
  account_replication_type  = "LRS"
  enable_https_traffic_only = false

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_blobStorage(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_kind             = "BlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_blobStorageUpdate(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_kind             = "BlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_storageV2(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_storageV2Update(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = "${azurestack_resource_group.testrg.name}"

  location                 = "${azurestack_resource_group.testrg.location}"
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}

func testAccAzureStackStorageAccount_nonStandardCasing(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "testrg" {
  name     = "testAccAzureStackSA-%d"
  location = "%s"
}

resource "azurestack_storage_account" "testsa" {
  name                     = "unlikely23exst2acct%s"
  resource_group_name      = "${azurestack_resource_group.testrg.name}"
  location                 = "${azurestack_resource_group.testrg.location}"
  account_tier             = "standard"
  account_replication_type = "lrs"

  tags = {
    environment = "production"
  }
}
`, rInt, location, rString)
}
