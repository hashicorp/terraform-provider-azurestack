package keyvault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type KeyVaultKeyDataSource struct{}

func TestAccKeyVaultKeyDataSource_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_key", "test")
	r := KeyVaultKeyDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_type").HasValue("RSA"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.hello").HasValue("world"),
			),
		},
	})
}

func (KeyVaultKeyDataSource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_key_vault_key" "test" {
  name         = azurestack_key_vault_key.test.name
  key_vault_id = azurestack_key_vault.test.id
}
`, KeyVaultKeyResource{}.complete(data))
}
