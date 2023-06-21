// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keyvault

import (
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/keyvault/mgmt/keyvault"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func keyVaultAccessPolicyDataSource() *schema.Resource {
	return &schema.Resource{
		Read: keyVaultAccessPolicyDataSourceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Key Management",
					"Secret Management",
					"Certificate Management",
					"Key & Secret Management",
					"Key & Certificate Management",
					"Secret & Certificate Management",
					"Key, Secret, & Certificate Management",
				}, false),
			},

			// Computed
			"certificate_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"key_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secret_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func keyVaultAccessPolicyDataSourceRead(d *schema.ResourceData, _ interface{}) error {
	name := d.Get("name").(string)
	templateManagementPermissions := map[string][]string{
		"key": {
			string(keyvault.KeyPermissionsGet),
			string(keyvault.KeyPermissionsList),
			string(keyvault.KeyPermissionsUpdate),
			string(keyvault.KeyPermissionsCreate),
			string(keyvault.KeyPermissionsImport),
			string(keyvault.KeyPermissionsDelete),
			string(keyvault.KeyPermissionsRecover),
			string(keyvault.KeyPermissionsBackup),
			string(keyvault.KeyPermissionsRestore),
		},
		"secret": {
			string(keyvault.SecretPermissionsGet),
			string(keyvault.SecretPermissionsList),
			string(keyvault.SecretPermissionsSet),
			string(keyvault.SecretPermissionsDelete),
			string(keyvault.SecretPermissionsRecover),
			string(keyvault.SecretPermissionsBackup),
			string(keyvault.SecretPermissionsRestore),
		},
		"certificate": {
			string(keyvault.Get),
			string(keyvault.List),
			string(keyvault.Update),
			string(keyvault.Create),
			string(keyvault.Import),
			string(keyvault.Delete),
			string(keyvault.Managecontacts),
			string(keyvault.Manageissuers),
			string(keyvault.Getissuers),
			string(keyvault.Listissuers),
			string(keyvault.Setissuers),
			string(keyvault.Deleteissuers),
		},
	}

	d.SetId(name)

	if strings.Contains(name, "Key") {
		d.Set("key_permissions", templateManagementPermissions["key"])
	}
	if strings.Contains(name, "Secret") {
		d.Set("secret_permissions", templateManagementPermissions["secret"])
	}
	if strings.Contains(name, "Certificate") {
		d.Set("certificate_permissions", templateManagementPermissions["certificate"])
	}

	return nil
}
