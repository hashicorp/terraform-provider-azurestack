package keyvault_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type KeyVaultAccessPolicyDataSource struct{}

func TestAccKeyVaultAccessPolicyDataSource_key(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key Management"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				resource.TestCheckNoResourceAttr(data.ResourceName, "secret_permissions"),
				resource.TestCheckNoResourceAttr(data.ResourceName, "certificate_permissions"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicyDataSource_secret(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Secret Management"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr(data.ResourceName, "key_permissions"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				resource.TestCheckNoResourceAttr(data.ResourceName, "certificate_permissions"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicyDataSource_certificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Certificate Management"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr(data.ResourceName, "key_permissions"),
				resource.TestCheckNoResourceAttr(data.ResourceName, "secret_permissions"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicyDataSource_keySecret(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key & Secret Management"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				resource.TestCheckNoResourceAttr(data.ResourceName, "certificate_permissions"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicyDataSource_keyCertificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key & Certificate Management"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				resource.TestCheckNoResourceAttr(data.ResourceName, "secret_permissions"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicyDataSource_secretCertificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Secret & Certificate Management"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr(data.ResourceName, "key_permissions"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func TestAccKeyVaultAccessPolicyDataSource_keySecretCertificate(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_key_vault_access_policy", "test")
	r := KeyVaultAccessPolicyDataSource{}
	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.testAccDataSourceKeyVaultAccessPolicy("Key, Secret, & Certificate Management"),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("key_permissions.#").HasValue("9"),
				check.That(data.ResourceName).Key("secret_permissions.#").HasValue("7"),
				check.That(data.ResourceName).Key("certificate_permissions.#").HasValue("12"),
			),
		},
	})
}

func (r KeyVaultAccessPolicyDataSource) testAccDataSourceKeyVaultAccessPolicy(name string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

data "azurestack_key_vault_access_policy" "test" {
  name = "%s"
}
`, name)
}
