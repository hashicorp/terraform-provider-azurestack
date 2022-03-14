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
