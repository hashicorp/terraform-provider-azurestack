package azurestack

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureStackStorageAccount_importBasic(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	config := testAccAzureStackStorageAccount_basic(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccazureStackStorageAccount_importPremium(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	config := testAccAzureStackStorageAccount_premium(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackStorageAccount_importNonStandardCasing(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	config := testAccAzureStackStorageAccount_nonStandardCasing(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureStackStorageAccount_importBlobEncryption(t *testing.T) {
	resourceName := "azurestack_storage_account.testsa"

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	config := testAccAzureStackStorageAccount_blobEncryption(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// File encryption not supported by the profile 2017-03-09
func TestAccAzureStackStorageAccount_importFileEncryption(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_storage_account.testsa"

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	config := testAccAzureStackStorageAccount_fileEncryption(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// EnableHttpsTraffic is not supported on 2017-03-09
func TestAccAzureStackStorageAccount_importEnableHttpsTrafficOnly(t *testing.T) {

	t.Skip()

	resourceName := "azurestack_storage_account.testsa"

	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	config := testAccAzureStackStorageAccount_enableHttpsTrafficOnly(ri, rs, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackStorageAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
