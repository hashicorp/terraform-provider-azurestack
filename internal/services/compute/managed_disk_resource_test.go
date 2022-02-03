package compute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type ManagedDiskResource struct{}

func TestAccManagedDisk_empty(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.empty(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccManagedDisk_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.empty(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_managed_disk"),
		},
	})
}

func TestAccManagedDisk_zeroGbFromPlatformImage(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.zeroGbFromPlatformImage(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
			ExpectNonEmptyPlan: true, // since the `disk_size_gb` will have changed
		},
	})
}

func TestAccManagedDisk_import(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}
	vm := VirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// need to create a vm and then delete it so we can use the vhd to test import
			Config:             vm.basicLinuxMachine(data),
			Destroy:            false,
			ExpectNonEmptyPlan: true,
			Check: acceptance.ComposeTestCheckFunc(
				// TODO: switch to using `azurestack_linux_virtual_machine` once Binary Testing is enabled
				check.That("azurestack_virtual_machine.test").ExistsInAzure(vm),
				data.CheckWithClientForResource(r.destroyVirtualMachine, "azurestack_virtual_machine.test"),
			),
		},
		{
			Config: r.importConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccManagedDisk_copy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.copy(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccManagedDisk_fromPlatformImage(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.platformImage(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccManagedDisk_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.empty(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("acctest"),
				check.That(data.ResourceName).Key("tags.cost-center").HasValue("ops"),
				check.That(data.ResourceName).Key("disk_size_gb").HasValue("1"),
				check.That(data.ResourceName).Key("storage_account_type").HasValue(string(compute.StandardLRS)),
			),
		},
		{
			Config: r.empty_updated(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("acctest"),
				check.That(data.ResourceName).Key("disk_size_gb").HasValue("2"),
				check.That(data.ResourceName).Key("storage_account_type").HasValue(string(compute.PremiumLRS)),
			),
		},
	})
}

func TestAccManagedDisk_importEmpty_withZone(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.empty_withZone(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccManagedDisk_attachedDiskUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.managedDiskAttached(data, 10),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.managedDiskAttached(data, 20),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("disk_size_gb").HasValue("20"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccManagedDisk_attachedStorageTypeUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_managed_disk", "test")
	r := ManagedDiskResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.storageTypeUpdateWhilstAttached(data, "Standard_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.storageTypeUpdateWhilstAttached(data, "Premium_LRS"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (ManagedDiskResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ManagedDiskID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.DisksClient.Get(ctx, id.ResourceGroup, id.DiskName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Managed Disk %q", id.String())
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (ManagedDiskResource) destroyVirtualMachine(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	vmName := state.Attributes["name"]
	resourceGroup := state.Attributes["resource_group_name"]

	var forceDeletion *bool = nil
	future, err := client.Compute.VMClient.Delete(ctx, resourceGroup, vmName, forceDeletion)
	if err != nil {
		return fmt.Errorf("Bad: Delete on vmClient: %+v", err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Compute.VMClient.Client); err != nil {
		return fmt.Errorf("Bad: Delete on vmClient: %+v", err)
	}

	return nil
}

func (ManagedDiskResource) empty(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (ManagedDiskResource) requiresImport(data acceptance.TestData) string {
	template := ManagedDiskResource{}.empty(data)
	return fmt.Sprintf(`
%s

resource "azurestack_managed_disk" "import" {
  name                 = azurestack_managed_disk.test.name
  location             = azurestack_managed_disk.test.location
  resource_group_name  = azurestack_managed_disk.test.resource_group_name
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

func (ManagedDiskResource) empty_withZone(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
  #zones                = ["1"]

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (ManagedDiskResource) importConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

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
  storage_account_id   = azurestack_storage_account.test.id
  disk_size_gb         = "45"

  tags = {
    environment = "acctest"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (ManagedDiskResource) copy(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "source" {
  name                 = "acctestd1-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
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
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Copy"
  source_resource_id   = azurestack_managed_disk.source.id
  disk_size_gb         = "1"

  tags = {
    environment = "acctest"
    cost-center = "ops"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (ManagedDiskResource) empty_updated(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestd-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Premium_LRS"
  create_option        = "Empty"
  disk_size_gb         = "2"

  tags = {
    environment = "acctest"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (ManagedDiskResource) platformImage(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

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
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  os_type              = "Linux"
  create_option        = "FromImage"
  image_reference_id   = data.azurestack_platform_image.test.id
  storage_account_type = "Standard_LRS"
}
`, data.Locations.Primary, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (ManagedDiskResource) zeroGbFromPlatformImage(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

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
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  os_type              = "Linux"
  create_option        = "FromImage"
  disk_size_gb         = 0
  image_reference_id   = data.azurestack_platform_image.test.id
  storage_account_type = "Standard_LRS"
}
`, data.Locations.Primary, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r ManagedDiskResource) managedDiskAttached(data acceptance.TestData, diskSize int) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_managed_disk" "test" {
  name                 = "%d-disk1"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = %d
}

resource "azurestack_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurestack_managed_disk.test.id
  virtual_machine_id = azurestack_linux_virtual_machine.test.id
  lun                = "0"
  caching            = "None"
}
`, r.templateAttached(data), data.RandomInteger, diskSize)
}

func (r ManagedDiskResource) storageTypeUpdateWhilstAttached(data acceptance.TestData, storageAccountType string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_managed_disk" "test" {
  name                 = "acctestdisk-%d"
  location             = azurestack_resource_group.test.location
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "%s"
  create_option        = "Empty"
  disk_size_gb         = 10
}

resource "azurestack_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurestack_managed_disk.test.id
  virtual_machine_id = azurestack_linux_virtual_machine.test.id
  lun                = "0"
  caching            = "None"
}
`, r.templateAttached(data), data.RandomInteger, storageAccountType)
}

func (ManagedDiskResource) templateAttached(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_linux_virtual_machine" "test" {
  name                            = "acctestvm-%d"
  resource_group_name             = azurestack_resource_group.test.name
  location                        = azurestack_resource_group.test.location
  size                            = "Standard_D2_v3"
  admin_username                  = "adminuser"
  admin_password                  = "Password1234!"
  disable_password_authentication = false

  network_interface_ids = [
    azurestack_network_interface.test.id,
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
