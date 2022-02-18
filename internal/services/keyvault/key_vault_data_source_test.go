package keyvault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type KeyVaultDataSource struct{}

func TestAccKeyVaultDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault", "test")
	r := KeyVaultDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.tenant_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.object_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Create"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Set"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccKeyVaultDataSource_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault", "test")
	r := KeyVaultDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tenant_id").Exists(),
				check.That(data.ResourceName).Key("sku_name").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.tenant_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.object_id").Exists(),
				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Get"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("Production"),
			),
		},
	})
}

// NOTE: argument named "service_endpoints" is not expected in resource "azurestack_subnet".
// func TestAccKeyVaultDataSource_networkAcls(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "data.azurestack_key_vault", "test")
// 	r := KeyVaultDataSource{}

// 	data.DataSourceTest(t, []resource.TestStep{
// 		{
// 			Config: r.networkAcls(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).Key("tenant_id").Exists(),
// 				check.That(data.ResourceName).Key("sku_name").Exists(),
// 				check.That(data.ResourceName).Key("access_policy.0.tenant_id").Exists(),
// 				check.That(data.ResourceName).Key("access_policy.0.object_id").Exists(),
// 				check.That(data.ResourceName).Key("access_policy.0.key_permissions.0").HasValue("Create"),
// 				check.That(data.ResourceName).Key("access_policy.0.secret_permissions.0").HasValue("Set"),
// 				check.That(data.ResourceName).Key("network_acls.#").HasValue("1"),
// 				check.That(data.ResourceName).Key("network_acls.0.default_action").HasValue("Allow"),
// 				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
// 			),
// 		},
// 	})
// }

// func TestAccKeyVaultDataSource_softDelete(t *testing.T) {
// 	data := acceptance.BuildTestData(t, "data.azurestack_key_vault", "test")
// 	r := KeyVaultDataSource{}

// 	data.DataSourceTest(t, []resource.TestStep{
// 		{
// 			Config: r.enableSoftDelete(data),
// 			Check: resource.ComposeTestCheckFunc(
// 				check.That(data.ResourceName).Key("soft_delete_enabled").HasValue("true"),
// 				check.That(data.ResourceName).Key("purge_protection_enabled").HasValue("false"),
// 				check.That(data.ResourceName).Key("sku_name").Exists(),
// 				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
// 			),
// 		},
// 	})
// }

func (KeyVaultDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault" "test" {
  name                = azurestack_key_vault.test.name
  resource_group_name = azurestack_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.basic(data))
}

func (KeyVaultDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault" "test" {
  name                = azurestack_key_vault.test.name
  resource_group_name = azurestack_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.complete(data))
}

// nolint:unused
func (KeyVaultDataSource) networkAcls(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault" "test" {
  name                = azurestack_key_vault.test.name
  resource_group_name = azurestack_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.networkAclsUpdated(data))
}

// nolint:unused
func (KeyVaultDataSource) enableSoftDelete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault" "test" {
  name                = azurestack_key_vault.test.name
  resource_group_name = azurestack_key_vault.test.resource_group_name
}
`, KeyVaultResource{}.softDelete(data))
}
