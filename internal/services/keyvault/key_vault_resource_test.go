package keyvault_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault/parse"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type KeyVaultResource struct{}

func TestAccKeyVault_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("sku_name").HasValue("standard"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKeyVault_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_key_vault"),
		},
	})
}

// NOTE: argument named "service_endpoints" is not expected in resource "azurestack_subnet".
// func TestAccKeyVault_networkAcls(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.networkAcls(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			Config: r.networkAclsUpdated(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// NOTE: argument named "service_endpoints" is not expected in resource "azurestack_subnet".
// func TestAccKeyVault_networkAclsAllowed(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.networkAclsAllowed(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// NOTE:  "azurestack_storage_account" has no argument, nested block, or exported attribute named "identity"
// func TestAccKeyVault_accessPolicyUpperLimit(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.accessPolicyUpperLimit(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 	})
// }

func TestAccKeyVault_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccKeyVault_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("access_policy.0.application_id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKeyVault_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Create"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Set"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		{
			Config: r.update(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("enabled_for_deployment").HasValue("true"),
				check.That(data.ResourceName).Key("enabled_for_disk_encryption").HasValue("true"),
				check.That(data.ResourceName).Key("enabled_for_template_deployment").HasValue("true"),
				check.That(data.ResourceName).Key("enable_rbac_authorization").HasValue("false"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("Staging"),
			),
		},
		{
			Config: r.noAccessPolicyBlocks(data),
			Check: resource.ComposeTestCheckFunc(
				// There are no access_policy blocks in this configuration
				// at all, which means to ignore any existing policies and
				// so the one created in previous steps is still present.
				check.That(data.ResourceName).Key("access_policy.#").HasValue("1"),
			),
		},
		{
			Config: r.accessPolicyExplicitZero(data),
			Check: resource.ComposeTestCheckFunc(
				// This config explicitly sets access_policy = [], which
				// means to delete any existing policies.
				check.That(data.ResourceName).Key("access_policy.#").HasValue("0"),
			),
		},
	})
}

// NOTE: At this moment, ASH only supports standard keyvault.
// func TestAccKeyVault_upgradeSKU(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.basic(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("sku_name").HasValue("standard"),
// 			),
// 		},
// 		data.ImportStep(), {
// 			Config: r.basicPremiumSKU(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("sku_name").HasValue("premium"),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// NOTE: contact not yet supported
// func TestAccKeyVault_updateContacts(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.basic(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			Config: r.updateContacts(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			Config: r.basic(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

func TestAccKeyVault_justCert(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.justCert(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("access_policy.0.certificate_permissions.0").HasValue("Get"),
			),
		},
		data.ImportStep(),
	})
}

// func TestAccKeyVault_softDelete(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.softDelete(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("false"),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// func TestAccKeyVault_softDeleteRecovery(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			// create it regularly
// 			Config: r.softDelete(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("false"),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			// delete the key vault
// 			Config: r.softDeleteAbsent(data),
// 		},
// 		{
// 			// attempting to re-create it requires recovery, which is enabled by default
// 			Config: r.softDelete(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("false"),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// func TestAccKeyVault_softDeleteRecoveryDisabled(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			// create it regularly
// 			Config: r.softDeleteRecoveryDisabled(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("false"),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			// delete the key vault
// 			Config: r.softDeleteAbsent(data),
// 		},
// 		{
// 			// attempting to re-create it requires recovery, which is enabled by default
// 			Config:      r.softDeleteRecoveryDisabled(data),
// 			ExpectError: regexp.MustCompile("An existing soft-deleted Key Vault exists with the Name"),
// 		},
// 	})
// }

// func TestAccKeyVault_purgeProtectionEnabled(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.purgeProtection(data, true),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("true"),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// func TestAccKeyVault_purgeProtectionAndSoftDeleteEnabled(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.purgeProtectionAndSoftDelete(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("true"),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// func TestAccKeyVault_purgeProtectionViaUpdate(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.purgeProtection(data, false),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("false"),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			Config: r.purgeProtection(data, true),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("true"),
// 			),
// 		},
// 		data.ImportStep(),
// 	})
// }

// func TestAccKeyVault_purgeProtectionAttemptToDisable(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
// 	r := KeyVaultResource{}

// 	data.ResourceTest(t, r, []resource.TestStep{
// 		{
// 			Config: r.purgeProtection(data, true),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).ExistsInAzure(r),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("true"),
// 			),
// 		},
// 		data.ImportStep(),
// 		{
// 			Config:      r.purgeProtection(data, false),
// 			ExpectError: regexp.MustCompile("once Purge Protection has been Enabled it's not possible to disable it"),
// 		},
// 	})
// }

func TestAccKeyVault_deletePolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault", "test")
	r := KeyVaultResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.noPolicy(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("access_policy.#").HasValue("0"),
			),
		},
		data.ImportStep(),
	})
}

func (KeyVaultResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.VaultID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.KeyVault.VaultsClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading Key Vault (%s): %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (KeyVaultResource) Destroy(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.VaultID(state.ID)
	if err != nil {
		return nil, err
	}

	if _, err := client.KeyVault.VaultsClient.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		return nil, fmt.Errorf("deleting %s: %+v", id, err)
	}

	return utils.Bool(true), nil
}

func (KeyVaultResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    certificate_permissions = [
      "Managecontacts",
    ]

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r KeyVaultResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault" "import" {
  name                = azurestack_key_vault.test.name
  location            = azurestack_key_vault.test.location
  resource_group_name = azurestack_key_vault.test.resource_group_name
  tenant_id           = azurestack_key_vault.test.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }
}
`, r.basic(data))
}

// nolint:unused
func (KeyVaultResource) networkAclsTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctestvirtnet%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test_a" {
  name                 = "acctestsubneta%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
  service_endpoints    = ["Microsoft.KeyVault"]
}

resource "azurestack_subnet" "test_b" {
  name                 = "acctestsubnetb%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.4.0/24"
  service_endpoints    = ["Microsoft.KeyVault"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

// nolint:unused
func (KeyVaultResource) networkAcls(data acceptance.TestData) string {
	template := KeyVaultResource{}.networkAclsTemplate(data)
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }

  network_acls {
    default_action             = "Deny"
    bypass                     = "None"
    virtual_network_subnet_ids = [azurestack_subnet.test_a.id, azurestack_subnet.test_b.id]
  }
}
`, template, data.RandomInteger)
}

// nolint:unused
func (r KeyVaultResource) networkAclsUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }

  network_acls {
    default_action             = "Allow"
    bypass                     = "AzureServices"
    ip_rules                   = ["123.0.0.102/32", "123.0.0.101"]
    virtual_network_subnet_ids = [azurestack_subnet.test_a.id]
  }
}
`, r.networkAclsTemplate(data), data.RandomInteger)
}

// nolint:unused
func (r KeyVaultResource) networkAclsAllowed(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }

  network_acls {
    default_action = "Allow"
    bypass         = "AzureServices"
  }
}
`, r.networkAclsTemplate(data), data.RandomInteger)
}

func (KeyVaultResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    key_permissions = [
      "Get",
    ]

    secret_permissions = [
      "Get",
    ]
  }

  enabled_for_deployment          = true
  enabled_for_disk_encryption     = true
  enabled_for_template_deployment = true
  enable_rbac_authorization       = false

  tags = {
    environment = "Staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KeyVaultResource) noAccessPolicyBlocks(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  enabled_for_deployment          = true
  enabled_for_disk_encryption     = true
  enabled_for_template_deployment = true
  enable_rbac_authorization       = false

  tags = {
    environment = "Staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KeyVaultResource) accessPolicyExplicitZero(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy = []

  enabled_for_deployment          = true
  enabled_for_disk_encryption     = true
  enabled_for_template_deployment = true
  enable_rbac_authorization       = false

  tags = {
    environment = "Staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KeyVaultResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id      = data.azurestack_client_config.current.tenant_id
    object_id      = data.azurestack_client_config.current.service_principal_object_id
    application_id = data.azurestack_client_config.current.client_id

    certificate_permissions = [
      "Get",
    ]

    key_permissions = [
      "Get",
    ]

    secret_permissions = [
      "Get",
    ]
  }

  tags = {
    environment = "Production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KeyVaultResource) justCert(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    certificate_permissions = [
      "Get",
    ]
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

// nolint:unused
func (KeyVaultResource) accessPolicyUpperLimit(data acceptance.TestData) string {
	var storageAccountConfigs string
	var accessPoliciesConfigs string

	for i := 1; i <= 20; i++ {
		storageAccountConfigs += testAccKeyVault_generateStorageAccountConfigs(i, data.RandomString)
		accessPoliciesConfigs += testAccKeyVault_generateAccessPolicyConfigs(i)
	}

	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  %s
}

%s
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, accessPoliciesConfigs, storageAccountConfigs)
}

// nolint:unused
func testAccKeyVault_generateStorageAccountConfigs(accountNum int, rs string) string {
	return fmt.Sprintf(`
resource "azurestack_storage_account" "test%d" {
  name                     = "testsa%s%d"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = azurestack_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "GRS"

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "testing"
  }
}
`, accountNum, rs, accountNum)
}

// nolint:unused
func testAccKeyVault_generateAccessPolicyConfigs(accountNum int) string {
	// due to a weird terraform fmt issue where:
	//   "${azurestack_storage_account.test%d.identity.0.principal_id}"
	// becomes
	//   "${azurestack_storage_account.test % d.identity.0.principal_id}"
	//
	// lets inject this separately so we can run terrafmt on this file

	oid := fmt.Sprintf("${azurestack_storage_account.test%d.identity.0.principal_id}", accountNum)

	return fmt.Sprintf(`
access_policy {
  tenant_id = data.azurestack_client_config.current.tenant_id
  object_id = "%s"

  key_permissions    = ["get", "create", "delete", "list", "restore", "recover", "unwrapkey", "wrapkey", "purge", "encrypt", "decrypt", "sign", "verify"]
  secret_permissions = ["get"]
}
`, oid)
}

// nolint:unused
func (KeyVaultResource) purgeProtection(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                     = "vault%d"
  location                 = azurestack_resource_group.test.location
  resource_group_name      = azurestack_resource_group.test.name
  tenant_id                = data.azurestack_client_config.current.tenant_id
  sku_name                 = "standard"
  purge_protection_enabled = "%t"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, enabled)
}

// nolint:unused
func (KeyVaultResource) softDelete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

// nolint:unused
func (KeyVaultResource) softDeleteAbsent(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {
    key_vault {
      recover_soft_deleted_key_vaults = false
    }
  }
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

// nolint:unused
func (KeyVaultResource) softDeleteRecoveryDisabled(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {
    key_vault {
      recover_soft_deleted_key_vaults = false
    }
  }
}

data "azurestack_client_config" "current" {}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                       = "vault%d"
  location                   = azurestack_resource_group.test.location
  resource_group_name        = azurestack_resource_group.test.name
  tenant_id                  = data.azurestack_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

// nolint:unused
func (KeyVaultResource) purgeProtectionAndSoftDelete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                       = "vault%d"
  location                   = azurestack_resource_group.test.location
  resource_group_name        = azurestack_resource_group.test.name
  tenant_id                  = data.azurestack_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
  purge_protection_enabled   = true
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KeyVaultResource) noPolicy(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "standard"

  access_policy = []
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

// nolint:unused
func (KeyVaultResource) updateContacts(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-kv-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                       = "vault%d"
  location                   = azurestack_resource_group.test.location
  resource_group_name        = azurestack_resource_group.test.name
  tenant_id                  = data.azurestack_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    certificate_permissions = [
      "Managecontacts",
    ]

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }

  contact {
    email = "example@example.com"
    name  = "example"
    phone = "01234567890"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

// NOTE: at this moment, only Standard SKU is supported
// nolint:unused
func (KeyVaultResource) basicPremiumSKU(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_client_config" "current" {
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "vault%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id
  sku_name            = "premium"

  access_policy {
    tenant_id = data.azurestack_client_config.current.tenant_id
    object_id = data.azurestack_client_config.current.service_principal_object_id

    certificate_permissions = [
      "Managecontacts",
    ]

    key_permissions = [
      "Create",
    ]

    secret_permissions = [
      "Set",
    ]
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
