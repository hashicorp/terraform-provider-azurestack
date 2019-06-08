package azurestack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureStackManagedDisk_empty(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	location := testLocation()
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_empty(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
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

func TestAccAzureStackManagedDisk_zeroGbFromPlatformImage(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_zeroGbFromPlatformImage(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
				),
			},
		},
	})
}

func TestAccAzureStackManagedDisk_import(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	location := testLocation()
	ri := acctest.RandInt()
	var vm compute.VirtualMachine
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				//need to create a vm and then delete it so we can use the vhd to test import
				Config:             testAccAzureStackVirtualMachine_basicLinuxMachine(ri, location),
				Destroy:            false,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackVirtualMachineExists("azurestack_virtual_machine.test", &vm),
					testDeleteAzureRMVirtualMachine("azurestack_virtual_machine.test"),
				),
			},
			{
				Config: testAccAzureStackManagedDisk_import(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
				),
			},
		},
	})
}

func TestAccAzureStackManagedDisk_copy(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_copy(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
				),
			},
		},
	})
}

func TestAccAzureStackManagedDisk_fromPlatformImage(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_platformImage(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
				),
			},
		},
	})
}

func TestAccAzureStackManagedDisk_update(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_empty(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "acctest"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost-center", "ops"),
					resource.TestCheckResourceAttr(resourceName, "disk_size_gb", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_account_type", string(compute.StandardLRS)),
				),
			},
			{
				Config: testAccAzureStackManagedDisk_empty_updated(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "acctest"),
					resource.TestCheckResourceAttr(resourceName, "disk_size_gb", "2"),
					resource.TestCheckResourceAttr(resourceName, "storage_account_type", string(compute.PremiumLRS)),
				),
			},
		},
	})
}

func TestAccAzureStackManagedDisk_encryption(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	rs := acctest.RandString(4)
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_encryption(ri, rs, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
					resource.TestCheckResourceAttr(resourceName, "encryption_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "encryption_settings.0.disk_encryption_key.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_settings.0.disk_encryption_key.0.secret_url"),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_settings.0.disk_encryption_key.0.source_vault_id"),
					resource.TestCheckResourceAttr(resourceName, "encryption_settings.0.key_encryption_key.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_settings.0.key_encryption_key.0.key_url"),
					resource.TestCheckResourceAttrSet(resourceName, "encryption_settings.0.key_encryption_key.0.source_vault_id"),
				),
			},
		},
	})
}

func TestAccAzureStackManagedDisk_NonStandardCasing(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	var d compute.Disk

	config := testAccAzureStackManagedDiskNonStandardCasing(ri, testLocation())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
				),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAzureStackManagedDisk_importEmpty_withZone(t *testing.T) {
	resourceName := "azurestack_managed_disk.test"
	ri := acctest.RandInt()
	var d compute.Disk

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureStackManagedDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureStackManagedDisk_empty_withZone(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureStackManagedDiskExists(resourceName, &d, true),
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

func testCheckAzureStackManagedDiskExists(resourceName string, d *compute.Disk, shouldExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		dName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for disk: %s", dName)
		}

		client := testAccProvider.Meta().(*ArmClient).diskClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, dName)
		if err != nil {
			return fmt.Errorf("Bad: Get on diskClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound && shouldExist {
			return fmt.Errorf("Bad: ManagedDisk %q (resource group %q) does not exist", dName, resourceGroup)
		}
		if resp.StatusCode != http.StatusNotFound && !shouldExist {
			return fmt.Errorf("Bad: ManagedDisk %q (resource group %q) still exists", dName, resourceGroup)
		}

		*d = resp

		return nil
	}
}

func testCheckAzureStackManagedDiskDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).diskClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurestack_managed_disk" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Managed Disk still exists: \n%#v", resp.DiskProperties)
		}
	}

	return nil
}

func testDeleteAzureRMVirtualMachine(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		vmName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for virtual machine: %s", vmName)
		}

		client := testAccProvider.Meta().(*ArmClient).vmClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, vmName)
		if err != nil {
			return fmt.Errorf("Bad: Delete on vmClient: %+v", err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("Bad: Delete on vmClient: %+v", err)
		}

		return nil
	}
}

func testAccAzureStackManagedDisk_empty(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackManagedDisk_requiresImport(rInt int, location string) string {
	template := testAccAzureStackManagedDisk_empty(rInt, location)
	return fmt.Sprintf(`
%s

resource "azurestack_managed_disk" "import" {
  name                 = "${azurestack_managed_disk.test.name}"
  location             = "${azurestack_managed_disk.test.location}"
  resource_group_name  = "${azurestack_managed_disk.test.resource_group_name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, template)
}

func testAccAzureStackManagedDisk_empty_withZone(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
  zones                = ["1"]

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackManagedDisk_import(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa%d"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Import"
  source_uri           = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
  disk_size_gb         = "45"

  tags = {
    environment = "acctest"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackManagedDisk_copy(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "source" {
  name                 = "acctestd1-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd2-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Copy"
  source_resource_id   = "${azurestack_managed_disk.source.id}"
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureStackManagedDisk_empty_updated(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Premium_LRS"
  create_option        = "Empty"
  disk_size_gb         = "2"

  tags = {
    environment = "acctest"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackManagedDiskNonStandardCasing(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "standard_lrs"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, rInt, location, rInt)
}

func testAccAzureStackManagedDisk_platformImage(rInt int, location string) string {
	return fmt.Sprintf(`
data "azurestack_platform_image" "test" {
  location  = "%s"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  os_type              = "Linux"
  create_option        = "FromImage"
  image_reference_id   = "${data.azurestack_platform_image.test.id}"
  storage_account_type = "Standard_LRS"
}
`, location, rInt, location, rInt)
}

func testAccAzureStackManagedDisk_zeroGbFromPlatformImage(rInt int, location string) string {
	return fmt.Sprintf(`
data "azurestack_platform_image" "test" {
  location  = "%s"
  publisher = "Canonical"
  offer     = "UbuntuServer"
  sku       = "16.04-LTS"
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  os_type              = "Linux"
  create_option        = "FromImage"
  disk_size_gb         = "0"
  image_reference_id   = "${data.azurestack_platform_image.test.id}"
  storage_account_type = "Standard_LRS"
}
`, location, rInt, location, rInt)
}

func testAccAzureStackManagedDisk_encryption(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
data "azurestack_client_config" "current" {}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "acctestkv-%s"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  tenant_id           = "${data.azurestack_client_config.current.tenant_id}"

  sku {
    name = "premium"
  }

  access_policy {
    tenant_id = "${data.azurestack_client_config.current.tenant_id}"
    object_id = "${data.azurestack_client_config.current.service_principal_object_id}"

    key_permissions = [
      "create",
      "delete",
      "get",
    ]

    secret_permissions = [
      "delete",
      "get",
      "set",
    ]
  }

  enabled_for_disk_encryption = true

  tags = {
    environment = "Production"
  }
}

resource "azurestack_key_vault_secret" "test" {
  name      = "secret-%s"
  value     = "szechuan"
  vault_uri = "${azurestack_key_vault.test.vault_uri}"
}

resource "azurestack_key_vault_key" "test" {
  name      = "key-%s"
  vault_uri = "${azurestack_key_vault.test.vault_uri}"
  key_type  = "EC"
  key_size  = 2048

  key_opts = [
    "sign",
    "verify",
  ]
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = "${azurestack_resource_group.test.location}"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  encryption_settings {
    enabled = true

    disk_encryption_key {
      secret_url      = "${azurestack_key_vault_secret.test.id}"
      source_vault_id = "${azurestack_key_vault.test.id}"
    }

    key_encryption_key {
      key_url         = "${azurestack_key_vault_key.test.id}"
      source_vault_id = "${azurestack_key_vault.test.id}"
    }
  }

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, rInt, location, rString, rString, rString, rInt)
}
