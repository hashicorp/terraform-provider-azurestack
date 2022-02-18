package keyvault

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "KeyVault"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Key Vault",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"azurestack_key_vault_access_policy": keyVaultAccessPolicyDataSource(),
		"azurestack_key_vault_key":           keyVaultKeyDataSource(),
		"azurestack_key_vault_secret":        keyVaultSecretDataSource(),
		"azurestack_key_vault":               keyVaultDataSource(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"azurestack_key_vault_access_policy": keyVaultAccessPolicy(),
		"azurestack_key_vault_key":           keyVaultKey(),
		"azurestack_key_vault_secret":        keyVaultSecret(),
		"azurestack_key_vault":               keyVault(),
	}
}
