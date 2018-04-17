package azurestack

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/authentication"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"subscription_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZURE_SUBSCRIPTION_ID", ""),
			},
			"tenant_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZURE_TENANT_ID", ""),
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZURE_CLIENT_ID", ""),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZURE_CLIENT_SECRET", ""),
			},
			"arm_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZURE_ARM_ENDPOINT", ""),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},
		ResourcesMap: map[string]*schema.Resource{
			"azurestack_resource_group": resourceAzureStackResourceGroup(),
		},
		ConfigureFunc: providerConfigureClient,
	}
}

func providerConfigureClient(d *schema.ResourceData) (interface{}, error) {
	config := &authentication.Config{
		SubscriptionID: d.Get("subscription_id").(string),
		ClientID:       d.Get("client_id").(string),
		ClientSecret:   d.Get("client_secret").(string),
		TenantID:       d.Get("tenant_id").(string),
		ARMEndpoint:    d.Get("arm_endpoint").(string),
	}

	client, err := getArmClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Resource group names can be capitalised, but we store them in lowercase.
// Use a custom diff function to avoid creation of new resources.
func resourceAzurermResourceGroupNameDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return strings.ToLower(old) == strings.ToLower(new)
}

// ignoreCaseDiffSuppressFunc is a DiffSuppressFunc from helper/schema that is
// used to ignore any case-changes in a return value.
func ignoreCaseDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return strings.ToLower(old) == strings.ToLower(new)
}

// ignoreCaseStateFunc is a StateFunc from helper/schema that converts the
// supplied value to lower before saving to state for consistency.
func ignoreCaseStateFunc(val interface{}) string {
	return strings.ToLower(val.(string))
}

func userDataDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	oldValue := userDataStateFunc(old)
	return oldValue == new
}

func userDataStateFunc(v interface{}) string {
	switch s := v.(type) {
	case string:
		s = base64Encode(s)
		hash := sha1.Sum([]byte(s))
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

// base64Encode encodes data if the input isn't already encoded using
// base64.StdEncoding.EncodeToString. If the input is already base64 encoded,
// return the original input unchanged.
func base64Encode(data string) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return data
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func isBase64Encoded(data string) bool {
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}
