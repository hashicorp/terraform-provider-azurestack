package netapp_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/netapp/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type NetAppVolumeResource struct {
}

func TestAccNetAppVolume_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}

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

func TestAccNetAppVolume_nfsv41(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.nfsv41(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetAppVolume_crossRegionReplication(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test_secondary")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.crossRegionReplication(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("data_protection_replication.0.endpoint_type").HasValue("dst"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetAppVolume_nfsv3FromSnapshot(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test_snapshot_vol")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.nfsv3FromSnapshot(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep("create_from_snapshot_resource_id"),
	})
}

func TestAccNetAppVolume_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_netapp_volume"),
		},
	})
}

func TestAccNetAppVolume_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("service_level").HasValue("Standard"),
				check.That(data.ResourceName).Key("storage_quota_in_gb").HasValue("101"),
				check.That(data.ResourceName).Key("export_policy_rule.#").HasValue("3"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.FoO").HasValue("BaR"),
				check.That(data.ResourceName).Key("mount_ip_addresses.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetAppVolume_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("storage_quota_in_gb").HasValue("100"),
				check.That(data.ResourceName).Key("export_policy_rule.#").HasValue("0"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("storage_quota_in_gb").HasValue("101"),
				check.That(data.ResourceName).Key("export_policy_rule.#").HasValue("3"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.FoO").HasValue("BaR"),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("storage_quota_in_gb").HasValue("100"),
				check.That(data.ResourceName).Key("export_policy_rule.#").HasValue("0"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetAppVolume_updateSubnet(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}
	resourceGroupName := fmt.Sprintf("acctestRG-netapp-%d", data.RandomInteger)
	oldVNetName := fmt.Sprintf("acctest-VirtualNetwork-%d", data.RandomInteger)
	oldSubnetName := fmt.Sprintf("acctest-Subnet-%d", data.RandomInteger)
	newVNetName := fmt.Sprintf("acctest-updated-VirtualNetwork-%d", data.RandomInteger)
	newSubnetName := fmt.Sprintf("acctest-updated-Subnet-%d", data.RandomInteger)
	uriTemplate := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/subnets/%s"

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	oldSubnetId := fmt.Sprintf(uriTemplate, subscriptionID, resourceGroupName, oldVNetName, oldSubnetName)
	newSubnetId := fmt.Sprintf(uriTemplate, subscriptionID, resourceGroupName, newVNetName, newSubnetName)

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet_id").HasValue(oldSubnetId),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateSubnet(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("subnet_id").HasValue(newSubnetId),
			),
		},
		data.ImportStep(),
	})
}

func TestAccNetAppVolume_updateExportPolicyRule(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_netapp_volume", "test")
	r := NetAppVolumeResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("export_policy_rule.#").HasValue("3"),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateExportPolicyRule(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("export_policy_rule.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func (t NetAppVolumeResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.VolumeID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.NetApp.VolumeClient.Get(ctx, id.ResourceGroup, id.NetAppAccountName, id.CapacityPoolName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading Netapp Volume (%s): %+v", id.String(), err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (NetAppVolumeResource) basic(data acceptance.TestData) string {
	template := NetAppVolumeResource{}.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_netapp_volume" "test" {
  name                = "acctest-NetAppVolume-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  volume_path         = "my-unique-file-path-%d"
  service_level       = "Standard"
  subnet_id           = azurerm_subnet.test.id
  storage_quota_in_gb = 100
}
`, template, data.RandomInteger, data.RandomInteger)
}

func (NetAppVolumeResource) nfsv41(data acceptance.TestData) string {
	template := NetAppVolumeResource{}.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_netapp_volume" "test" {
  name                = "acctest-NetAppVolume-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  volume_path         = "my-unique-file-path-%d"
  service_level       = "Standard"
  subnet_id           = azurerm_subnet.test.id
  protocols           = ["NFSv4.1"]
  security_style      = "Unix"
  storage_quota_in_gb = 100

  export_policy_rule {
    rule_index        = 1
    allowed_clients   = ["1.2.3.0/24"]
    protocols_enabled = ["NFSv4.1"]
    unix_read_only    = false
    unix_read_write   = true
  }
}
`, template, data.RandomInteger, data.RandomInteger)
}

func (NetAppVolumeResource) crossRegionReplication(data acceptance.TestData) string {
	template := NetAppVolumeResource{}.templateForCrossRegionReplication(data)
	return fmt.Sprintf(`
%[1]s

resource "azurerm_netapp_volume" "test_primary" {
  name                = "acctest-NetAppVolume-primary-%[2]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  volume_path         = "my-unique-file-path-primary-%[2]d"
  service_level       = "Standard"
  subnet_id           = azurerm_subnet.test.id
  protocols           = ["NFSv3"]
  storage_quota_in_gb = 100

  export_policy_rule {
    rule_index        = 1
    allowed_clients   = ["0.0.0.0/0"]
    protocols_enabled = ["NFSv3"]
    unix_read_only    = false
    unix_read_write   = true
  }
}

resource "azurerm_netapp_volume" "test_secondary" {
  name                = "acctest-NetAppVolume-secondary-%[2]d"
  location            = "%[3]s"
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test_secondary.name
  pool_name           = azurerm_netapp_pool.test_secondary.name
  volume_path         = "my-unique-file-path-secondary-%[2]d"
  service_level       = "Standard"
  subnet_id           = azurerm_subnet.test_secondary.id
  protocols           = ["NFSv3"]
  storage_quota_in_gb = 100

  export_policy_rule {
    rule_index        = 1
    allowed_clients   = ["0.0.0.0/0"]
    protocols_enabled = ["NFSv3"]
    unix_read_only    = false
    unix_read_write   = true
  }

  data_protection_replication {
    endpoint_type             = "dst"
    remote_volume_location    = azurerm_resource_group.test.location
    remote_volume_resource_id = azurerm_netapp_volume.test_primary.id
    replication_frequency     = "10minutes"
  }
}
`, template, data.RandomInteger, "northeurope")
}

func (NetAppVolumeResource) nfsv3FromSnapshot(data acceptance.TestData) string {
	template := NetAppVolumeResource{}.template(data)
	return fmt.Sprintf(`
%[1]s

resource "azurerm_netapp_volume" "test" {
  name                = "acctest-NetAppVolume-%[2]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  volume_path         = "my-unique-file-path-%[2]d"
  service_level       = "Standard"
  subnet_id           = azurerm_subnet.test.id
  protocols           = ["NFSv3"]
  storage_quota_in_gb = 100

  export_policy_rule {
    rule_index        = 1
    allowed_clients   = ["1.2.3.0/24"]
    protocols_enabled = ["NFSv3"]
    unix_read_only    = false
    unix_read_write   = true
  }
}

resource "azurerm_netapp_snapshot" "test" {
  name                = "acctest-Snapshot-%[2]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  volume_name         = azurerm_netapp_volume.test.name
}

resource "azurerm_netapp_volume" "test_snapshot_vol" {
  name                             = "acctest-NetAppVolume-NewFromSnapshot-%[2]d"
  location                         = azurerm_resource_group.test.location
  resource_group_name              = azurerm_resource_group.test.name
  account_name                     = azurerm_netapp_account.test.name
  pool_name                        = azurerm_netapp_pool.test.name
  volume_path                      = "my-unique-file-path-snapshot-%[2]d"
  service_level                    = "Standard"
  subnet_id                        = azurerm_subnet.test.id
  protocols                        = ["NFSv3"]
  storage_quota_in_gb              = 200
  create_from_snapshot_resource_id = azurerm_netapp_snapshot.test.id

  export_policy_rule {
    rule_index        = 1
    allowed_clients   = ["0.0.0.0/0"]
    protocols_enabled = ["NFSv3"]
    unix_read_write   = true
  }
}
`, template, data.RandomInteger)
}

func (r NetAppVolumeResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_netapp_volume" "import" {
  name                = azurerm_netapp_volume.test.name
  location            = azurerm_netapp_volume.test.location
  resource_group_name = azurerm_netapp_volume.test.resource_group_name
  account_name        = azurerm_netapp_volume.test.account_name
  pool_name           = azurerm_netapp_volume.test.pool_name
  volume_path         = azurerm_netapp_volume.test.volume_path
  service_level       = azurerm_netapp_volume.test.service_level
  subnet_id           = azurerm_netapp_volume.test.subnet_id
  storage_quota_in_gb = azurerm_netapp_volume.test.storage_quota_in_gb
}
`, r.basic(data))
}

func (r NetAppVolumeResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_netapp_volume" "test" {
  name                = "acctest-NetAppVolume-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  service_level       = "Standard"
  volume_path         = "my-unique-file-path-%d"
  subnet_id           = azurerm_subnet.test.id
  protocols           = ["NFSv3"]
  storage_quota_in_gb = 101

  export_policy_rule {
    rule_index        = 1
    allowed_clients   = ["1.2.3.0/24"]
    protocols_enabled = ["NFSv3"]
    unix_read_only    = false
    unix_read_write   = true
  }

  export_policy_rule {
    rule_index        = 2
    allowed_clients   = ["1.2.5.0"]
    protocols_enabled = ["NFSv3"]
    unix_read_only    = true
    unix_read_write   = false
  }

  export_policy_rule {
    rule_index      = 3
    allowed_clients = ["1.2.6.0/24"]
    cifs_enabled    = false
    nfsv3_enabled   = true
    nfsv4_enabled   = false
    unix_read_only  = true
    unix_read_write = false
  }

  tags = {
    "FoO" = "BaR"
  }
}
`, r.template(data), data.RandomInteger, data.RandomInteger)
}

func (r NetAppVolumeResource) updateSubnet(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_network" "updated" {
  name                = "acctest-updated-VirtualNetwork-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["10.1.0.0/16"]
}

resource "azurerm_subnet" "updated" {
  name                 = "acctest-updated-Subnet-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.updated.name
  address_prefix       = "10.1.3.0/24"

  delegation {
    name = "testdelegation2"

    service_delegation {
      name    = "Microsoft.Netapp/volumes"
      actions = ["Microsoft.Network/networkinterfaces/*", "Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_netapp_volume" "test" {
  name                = "acctest-updated-NetAppVolume-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  volume_path         = "my-updated-unique-file-path-%d"
  service_level       = "Standard"
  subnet_id           = azurerm_subnet.updated.id
  protocols           = ["NFSv3"]
  storage_quota_in_gb = 100
}
`, r.template(data), data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r NetAppVolumeResource) updateExportPolicyRule(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_netapp_volume" "test" {
  name                = "acctest-NetAppVolume-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  pool_name           = azurerm_netapp_pool.test.name
  service_level       = "Standard"
  volume_path         = "my-unique-file-path-%d"
  subnet_id           = azurerm_subnet.test.id
  protocols           = ["NFSv3"]
  storage_quota_in_gb = 101

  export_policy_rule {
    rule_index          = 1
    allowed_clients     = ["1.2.4.0/24", "1.3.4.0"]
    protocols_enabled   = ["NFSv3"]
    unix_read_only      = false
    unix_read_write     = true
    root_access_enabled = true
  }

  tags = {
    "FoO" = "BaR"
  }
}
`, r.template(data), data.RandomInteger, data.RandomInteger)
}

func (r NetAppVolumeResource) templateForCrossRegionReplication(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_virtual_network" "test_secondary" {
  name                = "acctest-VirtualNetwork-secondary-%[2]d"
  location            = "%[3]s"
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["10.6.0.0/16"]
}

resource "azurerm_subnet" "test_secondary" {
  name                 = "acctest-Subnet-secondary-%[2]d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test_secondary.name
  address_prefix       = "10.6.2.0/24"

  delegation {
    name = "testdelegation"

    service_delegation {
      name    = "Microsoft.Netapp/volumes"
      actions = ["Microsoft.Network/networkinterfaces/*", "Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_netapp_account" "test_secondary" {
  name                = "acctest-NetAppAccount-secondary-%[2]d"
  location            = "%[3]s"
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_netapp_pool" "test_secondary" {
  name                = "acctest-NetAppPool-secondary-%[2]d"
  location            = "%[3]s"
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test_secondary.name
  service_level       = "Standard"
  size_in_tb          = 4
}
`, r.template(data), data.RandomInteger, "northeurope")
}

func (NetAppVolumeResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-netapp-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctest-VirtualNetwork-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["10.6.0.0/16"]
}

resource "azurerm_subnet" "test" {
  name                 = "acctest-Subnet-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.6.2.0/24"

  delegation {
    name = "testdelegation"

    service_delegation {
      name    = "Microsoft.Netapp/volumes"
      actions = ["Microsoft.Network/networkinterfaces/*", "Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_netapp_account" "test" {
  name                = "acctest-NetAppAccount-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_netapp_pool" "test" {
  name                = "acctest-NetAppPool-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  account_name        = azurerm_netapp_account.test.name
  service_level       = "Standard"
  size_in_tb          = 4
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
