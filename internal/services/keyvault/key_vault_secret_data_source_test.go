package keyvault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type KeyVaultSecretDataSource struct{}

func TestAccKeyVaultSecretDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_secret", "test")
	r := KeyVaultSecretDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("value").HasValue("rick-and-morty"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccKeyVaultSecretDataSource_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_secret", "test")
	r := KeyVaultSecretDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("value").HasValue("<rick><morty /></rick>"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.hello").HasValue("world"),
			),
		},
	})
}

func (KeyVaultSecretDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault_secret" "test" {
  name         = azurestack_key_vault_secret.test.name
  key_vault_id = azurestack_key_vault.test.id
}
`, KeyVaultSecretResource{}.basic(data))
}

func (KeyVaultSecretDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault_secret" "test" {
  name         = azurestack_key_vault_secret.test.name
  key_vault_id = azurestack_key_vault.test.id
}
`, KeyVaultSecretResource{}.complete(data))
}
