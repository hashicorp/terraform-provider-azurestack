package keyvault_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/resourceids"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type KeyVaultAccessPolicyResource struct{}

func TestAccKeyVaultAccessPolicy_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("key_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.1").HasValue("Set"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKeyVaultAccessPolicy_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("key_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.1").HasValue("Set"),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_key_vault_access_policy"),
		},
	})
}

func TestAccKeyVaultAccessPolicy_multiple(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_access_policy", "test_with_application_id")
	r := KeyVaultAccessPolicyResource{}
	resourceName2 := "azurestack_key_vault_access_policy.test_no_application_id"

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.multiple(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("key_permissions.0").HasValue("Create"),
				check.That(data.ResourceName).Key("key_permissions.1").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.1").HasValue("Delete"),
				check.That(data.ResourceName).Key("certificate_permissions.0").HasValue("Create"),
				check.That(data.ResourceName).Key("certificate_permissions.1").HasValue("Delete"),
				resource.TestCheckResourceAttr(resourceName2, "key_permissions.0", "List"),
				resource.TestCheckResourceAttr(resourceName2, "key_permissions.1", "Encrypt"),
				resource.TestCheckResourceAttr(resourceName2, "secret_permissions.0", "List"),
				resource.TestCheckResourceAttr(resourceName2, "secret_permissions.1", "Delete"),
				resource.TestCheckResourceAttr(resourceName2, "certificate_permissions.0", "List"),
				resource.TestCheckResourceAttr(resourceName2, "certificate_permissions.1", "Delete"),
			),
		},
		data.ImportStep(),
		{
			ResourceName:      resourceName2,
			ImportState:       true,
			ImportStateVerify: true,
		},
	})
}

func TestAccKeyVaultAccessPolicy_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("key_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("secret_permissions.1").HasValue("Set"),
			),
		},
		{
			Config: r.update(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("key_permissions.0").HasValue("List"),
				check.That(data.ResourceName).Key("key_permissions.1").HasValue("Encrypt"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicy_nonExistentVault(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config:             r.nonExistentVault(data),
			ExpectNonEmptyPlan: true,
			ExpectError:        regexp.MustCompile(`Error retrieving Key Vault`),
		},
	})
}

func (t KeyVaultAccessPolicyResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := resourceids.ParseAzureResourceID(state.ID)
	if err != nil {
		return nil, err
	}
	resGroup := id.ResourceGroup
	vaultName := id.Path["vaults"]
	objectId := id.Path["objectId"]
	applicationId := id.Path["applicationId"]

	resp, err := clients.KeyVault.VaultsClient.Get(ctx, resGroup, vaultName)
	if err != nil {
		return nil, fmt.Errorf("reading Key Vault (%s): %+v", id, err)
	}

	return utils.Bool(keyvault.FindKeyVaultAccessPolicy(resp.Properties.AccessPolicies, objectId, applicationId) != nil), nil
}

func (r KeyVaultAccessPolicyResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault_access_policy" "test" {
  key_vault_id = azurestack_key_vault.test.id

  key_permissions = [
    "Get",
  ]

  secret_permissions = [
    "Get",
    "Set",
  ]

  tenant_id = data.azurestack_client_config.current.tenant_id
  object_id = data.azurestack_client_config.current.object_id
}
`, r.template(data))
}

func (r KeyVaultAccessPolicyResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault_access_policy" "import" {
  key_vault_id = azurestack_key_vault.test.id
  tenant_id    = azurestack_key_vault_access_policy.test.tenant_id
  object_id    = azurestack_key_vault_access_policy.test.object_id

  key_permissions = [
    "Get",
  ]

  secret_permissions = [
    "Get",
    "Set",
  ]
}
`, r.basic(data))
}

func (r KeyVaultAccessPolicyResource) multiple(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault_access_policy" "test_with_application_id" {
  key_vault_id = azurestack_key_vault.test.id

  key_permissions = [
    "Create",
    "Get",
  ]

  secret_permissions = [
    "Get",
    "Delete",
  ]

  certificate_permissions = [
    "Create",
    "Delete",
  ]

  application_id = data.azurestack_client_config.current.client_id
  tenant_id      = data.azurestack_client_config.current.tenant_id
  object_id      = data.azurestack_client_config.current.object_id
}

resource "azurestack_key_vault_access_policy" "test_no_application_id" {
  key_vault_id = azurestack_key_vault.test.id

  key_permissions = [
    "List",
    "Encrypt",
  ]

  secret_permissions = [
    "List",
    "Delete",
  ]

  certificate_permissions = [
    "List",
    "Delete",
  ]

  storage_permissions = [
    "Backup",
    "Delete",
    "Deletesas",
    "Get",
    "Getsas",
    "List",
    "Listsas",
    "Purge",
    "Recover",
    "Regeneratekey",
    "Restore",
    "Set",
    "Setsas",
    "Update",
  ]

  tenant_id = data.azurestack_client_config.current.tenant_id
  object_id = data.azurestack_client_config.current.object_id
}
`, r.template(data))
}

func (r KeyVaultAccessPolicyResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault_access_policy" "test" {
  key_vault_id = azurestack_key_vault.test.id

  key_permissions = [
    "List",
    "Encrypt",
  ]

  secret_permissions = []

  tenant_id = data.azurestack_client_config.current.tenant_id
  object_id = data.azurestack_client_config.current.object_id
}
`, r.template(data))
}

func (KeyVaultAccessPolicyResource) template(data acceptance.TestData) string {
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
  name                = "acctestkv-%s"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id

  sku_name = "standard"

  tags = {
    environment = "Production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (KeyVaultAccessPolicyResource) nonExistentVault(data acceptance.TestData) string {
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
  name                = "acctestkv-%s"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  tenant_id           = data.azurestack_client_config.current.tenant_id

  sku_name = "standard"

  tags = {
    environment = "Production"
  }
}

resource "azurestack_key_vault_access_policy" "test" {
  # Must appear to be URL, but not actually exist - appending a string works
  key_vault_id = "${azurestack_key_vault.test.id}NOPE"

  tenant_id = data.azurestack_client_config.current.tenant_id
  object_id = data.azurestack_client_config.current.object_id

  key_permissions = [
    "Get",
  ]

  secret_permissions = [
    "Get",
  ]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}
