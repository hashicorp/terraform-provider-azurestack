package keyvault_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type KeyVaultSecretResource struct{}

func TestAccKeyVaultSecret_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("value").HasValue("rick-and-morty"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKeyVaultSecret_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("value").HasValue("rick-and-morty"),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_key_vault_secret"),
		},
	})
}

func TestAccKeyVaultSecret_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccKeyVaultSecret_disappearsWhenParentKeyVaultDeleted(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				data.CheckWithClientForResource(r.destroyParentKeyVault, "azurestack_key_vault.test"),
			),
			ExpectNonEmptyPlan: true,
		},
	})
}

func TestAccKeyVaultSecret_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("not_before_date").HasValue("2019-01-01T01:02:03Z"),
				check.That(data.ResourceName).Key("expiration_date").HasValue("2020-01-01T01:02:03Z"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.hello").HasValue("world"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKeyVaultSecret_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("value").HasValue("rick-and-morty"),
			),
		},
		{
			Config: r.basicUpdated(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("value").HasValue("szechuan"),
			),
		},
	})
}

func TestAccKeyVaultSecret_updatingValueChangedExternally(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("value").HasValue("rick-and-morty"),
				data.CheckWithClient(r.updateSecretValue("mad-scientist")),
			),
			ExpectNonEmptyPlan: true,
		},
		{
			Config: r.updateTags(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:   r.updateTags(data),
			PlanOnly: true,
		},
		data.ImportStep(),
	})
}

func TestAccKeyVaultSecret_withExternalAccessPolicy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_key_vault_secret", "test")
	r := KeyVaultSecretResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.withExternalAccessPolicy(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.withExternalAccessPolicyUpdate(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (KeyVaultSecretResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	client := clients.KeyVault.ManagementClient
	keyVaultsClient := clients.KeyVault

	id, err := parse.ParseNestedItemID(state.ID)
	if err != nil {
		return nil, err
	}

	keyVaultIdRaw, err := keyVaultsClient.KeyVaultIDFromBaseUrl(ctx, clients.Resource, id.KeyVaultBaseUrl)
	if err != nil || keyVaultIdRaw == nil {
		return nil, fmt.Errorf("retrieving the Resource ID the Key Vault at URL %q: %s", id.KeyVaultBaseUrl, err)
	}
	keyVaultId, err := parse.VaultID(*keyVaultIdRaw)
	if err != nil {
		return nil, err
	}

	ok, err := keyVaultsClient.Exists(ctx, *keyVaultId)
	if err != nil || !ok {
		return nil, fmt.Errorf("checking if key vault %q for Certificate %q in Vault at url %q exists: %v", *keyVaultId, id.Name, id.KeyVaultBaseUrl, err)
	}

	// we always want to get the latest version
	resp, err := client.GetSecret(ctx, id.KeyVaultBaseUrl, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("making Read request on Azure KeyVault Secret %s: %+v", id.Name, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (KeyVaultSecretResource) Destroy(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	dataPlaneClient := client.KeyVault.ManagementClient

	name := state.Attributes["name"]
	keyVaultId, err := parse.VaultID(state.Attributes["key_vault_id"])
	if err != nil {
		return nil, err
	}
	vaultBaseUrl, err := client.KeyVault.BaseUriForKeyVault(ctx, *keyVaultId)
	if err != nil {
		return nil, fmt.Errorf("looking up Secret %q vault url from id %q: %+v", name, keyVaultId, err)
	}

	if _, err := dataPlaneClient.DeleteSecret(ctx, *vaultBaseUrl, name); err != nil {
		return nil, fmt.Errorf("Bad: Delete on keyVaultManagementClient: %+v", err)
	}

	return utils.Bool(true), nil
}

func (KeyVaultSecretResource) destroyParentKeyVault(ctx context.Context, client *clients.Client, state *terraform.InstanceState) error {
	ok, err := KeyVaultResource{}.Destroy(ctx, client, state)
	if err != nil {
		return err
	}

	if ok == nil || !*ok {
		return fmt.Errorf("deleting parent key vault failed")
	}

	return nil
}

func (r KeyVaultSecretResource) updateSecretValue(value string) acceptance.ClientCheckFunc {
	return func(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) error {
		dataPlaneClient := clients.KeyVault.ManagementClient

		name := state.Attributes["name"]
		keyVaultId, err := parse.VaultID(state.Attributes["key_vault_id"])
		if err != nil {
			return err
		}

		vaultBaseUrl, err := clients.KeyVault.BaseUriForKeyVault(ctx, *keyVaultId)
		if err != nil {
			return fmt.Errorf("looking up Secret %q vault url from id %q: %+v", name, keyVaultId, err)
		}

		updated := keyvault.SecretSetParameters{
			Value: utils.String(value),
		}
		if _, err = dataPlaneClient.SetSecret(ctx, *vaultBaseUrl, name, updated); err != nil {
			return fmt.Errorf("updating secret: %+v", err)
		}
		return nil
	}
}

func (r KeyVaultSecretResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_key_vault_secret" "test" {
  name         = "secret-%s"
  value        = "rick-and-morty"
  key_vault_id = azurestack_key_vault.test.id
}
`, r.template(data), data.RandomString)
}

func (r KeyVaultSecretResource) updateTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_key_vault_secret" "test" {
  name         = "secret-%s"
  value        = "mad-scientist"
  key_vault_id = azurestack_key_vault.test.id

  tags = {
    Rick = "Morty"
  }
}
`, r.template(data), data.RandomString)
}

func (r KeyVaultSecretResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_key_vault_secret" "import" {
  name         = azurestack_key_vault_secret.test.name
  value        = azurestack_key_vault_secret.test.value
  key_vault_id = azurestack_key_vault_secret.test.key_vault_id
}
`, r.basic(data))
}

func (r KeyVaultSecretResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_key_vault_secret" "test" {
  name            = "secret-%s"
  value           = "<rick><morty /></rick>"
  key_vault_id    = azurestack_key_vault.test.id
  content_type    = "application/xml"
  not_before_date = "2019-01-01T01:02:03Z"
  expiration_date = "2020-01-01T01:02:03Z"

  tags = {
    "hello" = "world"
  }
}
`, r.template(data), data.RandomString)
}

func (r KeyVaultSecretResource) basicUpdated(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

%s

resource "azurestack_key_vault_secret" "test" {
  name         = "secret-%s"
  value        = "szechuan"
  key_vault_id = azurestack_key_vault.test.id
}
`, r.template(data), data.RandomString)
}

func (KeyVaultSecretResource) withExternalAccessPolicy(data acceptance.TestData) string {
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
  sku_name            = "standard"

  tags = {
    environment = "Production"
  }
}

resource "azurestack_key_vault_access_policy" "test" {
  key_vault_id = azurestack_key_vault.test.id
  tenant_id    = data.azurestack_client_config.current.tenant_id
  object_id    = data.azurestack_client_config.current.service_principal_object_id
  key_permissions = [
    "Create",
    "Get",
  ]
  secret_permissions = [
    "Set",
    "Get",
    "Delete",
    "Purge",
    "Recover"
  ]
}

resource "azurestack_key_vault_secret" "test" {
  name         = "secret-%s"
  value        = "rick-and-morty"
  key_vault_id = azurestack_key_vault.test.id
  depends_on   = [azurestack_key_vault_access_policy.test]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}

func (KeyVaultSecretResource) withExternalAccessPolicyUpdate(data acceptance.TestData) string {
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
  sku_name            = "standard"

  tags = {
    environment = "Production"
  }
}

resource "azurestack_key_vault_access_policy" "test" {
  key_vault_id = azurestack_key_vault.test.id
  tenant_id    = data.azurestack_client_config.current.tenant_id
  object_id    = data.azurestack_client_config.current.service_principal_object_id
  key_permissions = [
    "Create",
    "Get",
  ]
  secret_permissions = [
    "Set",
    "Get",
    "Delete",
    "Purge",
    "Recover"
  ]
}

resource "azurestack_key_vault_secret" "test" {
  name         = "secret-%s"
  value        = "rick-and-morty"
  key_vault_id = azurestack_key_vault.test.id
  depends_on   = [azurestack_key_vault_access_policy.test]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}

func (KeyVaultSecretResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
data "azurestack_client_config" "current" {}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_key_vault" "test" {
  name                = "acctestkv-%s"
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
      "Delete",
      "Purge",
      "Recover",
      "Set",
    ]
  }

  tags = {
    environment = "Production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}
