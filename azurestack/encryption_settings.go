package azurestack

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func encryptionSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeBool,
					Required: true,

					// Azure can change enabled from false to true, but not the other way around, so
					//   to keep idempotency, we'll conservatively set this to ForceNew=true
					ForceNew: true,
				},

				"disk_encryption_key": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"secret_url": {
								Type:     schema.TypeString,
								Required: true,
							},

							"source_vault_id": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"key_encryption_key": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key_url": {
								Type:     schema.TypeString,
								Required: true,
							},

							"source_vault_id": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func expandManagedDiskEncryptionSettings(settings map[string]interface{}) *compute.EncryptionSettings {
	enabled := settings["enabled"].(bool)
	config := &compute.EncryptionSettings{
		Enabled: utils.Bool(enabled),
	}

	if v := settings["disk_encryption_key"].([]interface{}); len(v) > 0 {
		dek := v[0].(map[string]interface{})

		secretURL := dek["secret_url"].(string)
		sourceVaultId := dek["source_vault_id"].(string)
		config.DiskEncryptionKey = &compute.KeyVaultAndSecretReference{
			SecretURL: utils.String(secretURL),
			SourceVault: &compute.SourceVault{
				ID: utils.String(sourceVaultId),
			},
		}
	}

	if v := settings["key_encryption_key"].([]interface{}); len(v) > 0 {
		kek := v[0].(map[string]interface{})

		secretURL := kek["key_url"].(string)
		sourceVaultId := kek["source_vault_id"].(string)
		config.KeyEncryptionKey = &compute.KeyVaultAndKeyReference{
			KeyURL: utils.String(secretURL),
			SourceVault: &compute.SourceVault{
				ID: utils.String(sourceVaultId),
			},
		}
	}

	return config
}

func flattenManagedDiskEncryptionSettings(encryptionSettings *compute.EncryptionSettings) []interface{} {
	value := map[string]interface{}{
		"enabled": *encryptionSettings.Enabled,
	}

	if key := encryptionSettings.DiskEncryptionKey; key != nil {
		keys := make(map[string]interface{})

		keys["secret_url"] = *key.SecretURL
		if vault := key.SourceVault; vault != nil {
			keys["source_vault_id"] = *vault.ID
		}

		value["disk_encryption_key"] = []interface{}{keys}
	}

	if key := encryptionSettings.KeyEncryptionKey; key != nil {
		keys := make(map[string]interface{})

		keys["key_url"] = *key.KeyURL

		if vault := key.SourceVault; vault != nil {
			keys["source_vault_id"] = *vault.ID
		}

		value["key_encryption_key"] = []interface{}{keys}
	}

	output := make([]interface{}, 0)
	output = append(output, value)
	return output
}
