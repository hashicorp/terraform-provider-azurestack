package keyvault

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/keyvault/mgmt/keyvault"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/locks"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/keyvault/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network"
	networkParse "github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/set"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

// As can be seen in the API definition, the Sku Family only supports the value
// `A` and is a required field
// https://github.com/Azure/azure-rest-api-specs/blob/master/arm-keyvault/2015-06-01/swagger/keyvault.json#L239
var armKeyVaultSkuFamily = "A"

var keyVaultResourceName = "azurestack_key_vault"

func keyVault() *schema.Resource {
	return &schema.Resource{
		Create: keyVaultCreate,
		Read:   keyVaultRead,
		Update: keyVaultUpdate,
		Delete: keyVaultDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 2,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: func() map[string]*schema.Schema {
			rSchema := map[string]*schema.Schema{
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validate.VaultName,
				},

				"location": commonschema.Location(),

				"resource_group_name": commonschema.ResourceGroupName(),

				"sku_name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						string(keyvault.Standard),
						// NOTE: At this moment, ASH only supports standard keyvault.
						// string(keyvault.Premium),
					}, false),
				},

				"tenant_id": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.IsUUID,
				},

				"access_policy": {
					Type:       schema.TypeList,
					ConfigMode: schema.SchemaConfigModeAttr,
					Optional:   true,
					Computed:   true,
					MaxItems:   1024,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"tenant_id": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.IsUUID,
							},
							"object_id": {
								Type:     schema.TypeString,
								Required: true,
								// ValidateFunc: validation.IsUUID, NOTE: In some circumstances, object_id is not a valid UUID
							},
							"application_id": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validate.IsUUIDOrEmpty,
							},
							"certificate_permissions": schemaCertificatePermissions(),
							"key_permissions":         schemaKeyPermissions(),
							"secret_permissions":      schemaSecretPermissions(),
							"storage_permissions":     schemaStoragePermissions(),
						},
					},
				},

				"enabled_for_deployment": {
					Type:     schema.TypeBool,
					Optional: true,
				},

				"enabled_for_disk_encryption": {
					Type:     schema.TypeBool,
					Optional: true,
				},

				"enabled_for_template_deployment": {
					Type:     schema.TypeBool,
					Optional: true,
				},

				"enable_rbac_authorization": {
					Type:     schema.TypeBool,
					Optional: true,
				},

				"network_acls": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_action": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(keyvault.Allow),
									string(keyvault.Deny),
								}, false),
							},
							"bypass": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.StringInSlice([]string{
									string(keyvault.None),
									string(keyvault.AzureServices),
								}, false),
							},
							"ip_rules": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									ValidateFunc: validation.Any(
										validation.IsIPv4Address,
										validation.IsCIDR,
									),
								},
								Set: set.HashIPv4AddressOrCIDR,
							},
							"virtual_network_subnet_ids": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								Set:      set.HashStringIgnoreCase,
							},
						},
					},
				},

				"tags": tags.Schema(),

				// Computed
				"vault_uri": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}

			return rSchema
		}(),
	}
}

func keyVaultCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).KeyVault.VaultsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVaultID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	location := location.Normalize(d.Get("location").(string))

	// Locking this resource so we don't make modifications to it at the same time if there is a
	// key vault access policy trying to update it as well
	locks.ByName(id.Name, keyVaultResourceName)
	defer locks.UnlockByName(id.Name, keyVaultResourceName)

	// check for the presence of an existing, live one which should be imported into the state
	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurestack_key_vault", id.ID())
	}

	tenantUUID := uuid.FromStringOrNil(d.Get("tenant_id").(string))
	enabledForDeployment := d.Get("enabled_for_deployment").(bool)
	enabledForDiskEncryption := d.Get("enabled_for_disk_encryption").(bool)
	enabledForTemplateDeployment := d.Get("enabled_for_template_deployment").(bool)
	enableRbacAuthorization := d.Get("enable_rbac_authorization").(bool)

	t := d.Get("tags").(map[string]interface{})

	policies := d.Get("access_policy").([]interface{})
	accessPolicies := expandAccessPolicies(policies)

	networkAclsRaw := d.Get("network_acls").([]interface{})
	networkAcls, subnetIds := expandKeyVaultNetworkAcls(networkAclsRaw)

	sku := keyvault.Sku{
		Family: &armKeyVaultSkuFamily,
		Name:   keyvault.SkuName(d.Get("sku_name").(string)),
	}

	parameters := keyvault.VaultCreateOrUpdateParameters{
		Location: &location,
		Properties: &keyvault.VaultProperties{
			TenantID:                     &tenantUUID,
			Sku:                          &sku,
			AccessPolicies:               accessPolicies,
			EnabledForDeployment:         &enabledForDeployment,
			EnabledForDiskEncryption:     &enabledForDiskEncryption,
			EnabledForTemplateDeployment: &enabledForTemplateDeployment,
			EnableRbacAuthorization:      &enableRbacAuthorization,
			NetworkAcls:                  networkAcls,

			// update: with ASH 2108, soft delete feature causes some major trouble when enabled.
			// no official documentation but this feature seems to be not fully supported
			EnableSoftDelete: utils.Bool(false),
		},
		Tags: tags.Expand(t),
	}

	// also lock on the Virtual Network ID's since modifications in the networking stack are exclusive
	virtualNetworkNames := make([]string, 0)
	for _, v := range subnetIds {
		id, err := networkParse.SubnetIDInsensitively(v)
		if err != nil {
			return err
		}
		if !utils.SliceContainsValue(virtualNetworkNames, id.VirtualNetworkName) {
			virtualNetworkNames = append(virtualNetworkNames, id.VirtualNetworkName)
		}
	}

	locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
	defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, parameters); err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	read, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}
	if read.Properties == nil || read.Properties.VaultURI == nil {
		return fmt.Errorf("retrieving %s: `properties.VaultUri` was nil", id)
	}
	d.SetId(id.ID())
	meta.(*clients.Client).KeyVault.AddToCache(id, *read.Properties.VaultURI)

	if props := read.Properties; props != nil {
		if vault := props.VaultURI; vault != nil {
			log.Printf("[DEBUG] Waiting for %s to become available", id)
			stateConf := &resource.StateChangeConf{
				Pending:                   []string{"pending"},
				Target:                    []string{"available"},
				Refresh:                   keyVaultRefreshFunc(*vault),
				Delay:                     30 * time.Second,
				PollInterval:              10 * time.Second,
				ContinuousTargetOccurence: 10,
				Timeout:                   d.Timeout(schema.TimeoutCreate),
			}

			if _, err := stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf("Error waiting for %s to become available: %s", id, err)
			}
		}
	}

	return keyVaultRead(d, meta)
}

func keyVaultUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).KeyVault.VaultsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VaultID(d.Id())
	if err != nil {
		return err
	}

	// Locking this resource so we don't make modifications to it at the same time if there is a
	// key vault access policy trying to update it as well
	locks.ByName(id.Name, keyVaultResourceName)
	defer locks.UnlockByName(id.Name, keyVaultResourceName)

	d.Partial(true)

	// first pull the existing key vault since we need to lock on several bits of its information
	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	if existing.Properties == nil {
		return fmt.Errorf("retrieving %s: `properties` was nil", *id)
	}

	update := keyvault.VaultPatchParameters{}

	if d.HasChange("access_policy") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		policiesRaw := d.Get("access_policy").([]interface{})
		accessPolicies := expandAccessPolicies(policiesRaw)
		update.Properties.AccessPolicies = accessPolicies
	}

	if d.HasChange("enabled_for_deployment") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnabledForDeployment = utils.Bool(d.Get("enabled_for_deployment").(bool))
	}

	if d.HasChange("enabled_for_disk_encryption") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnabledForDiskEncryption = utils.Bool(d.Get("enabled_for_disk_encryption").(bool))
	}

	if d.HasChange("enabled_for_template_deployment") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnabledForTemplateDeployment = utils.Bool(d.Get("enabled_for_template_deployment").(bool))
	}

	if d.HasChange("enable_rbac_authorization") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.EnableRbacAuthorization = utils.Bool(d.Get("enable_rbac_authorization").(bool))
	}

	if d.HasChange("network_acls") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		networkAclsRaw := d.Get("network_acls").([]interface{})
		networkAcls, subnetIds := expandKeyVaultNetworkAcls(networkAclsRaw)

		// also lock on the Virtual Network ID's since modifications in the networking stack are exclusive
		virtualNetworkNames := make([]string, 0)
		for _, v := range subnetIds {
			id, err := networkParse.SubnetIDInsensitively(v)
			if err != nil {
				return err
			}

			if !utils.SliceContainsValue(virtualNetworkNames, id.VirtualNetworkName) {
				virtualNetworkNames = append(virtualNetworkNames, id.VirtualNetworkName)
			}
		}

		locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
		defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

		update.Properties.NetworkAcls = networkAcls
	}

	if d.HasChange("sku_name") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		update.Properties.Sku = &keyvault.Sku{
			Family: &armKeyVaultSkuFamily,
			Name:   keyvault.SkuName(d.Get("sku_name").(string)),
		}
	}

	if d.HasChange("tenant_id") {
		if update.Properties == nil {
			update.Properties = &keyvault.VaultPatchProperties{}
		}

		tenantUUID := uuid.FromStringOrNil(d.Get("tenant_id").(string))
		update.Properties.TenantID = &tenantUUID
	}

	if d.HasChange("tags") {
		t := d.Get("tags").(map[string]interface{})
		update.Tags = tags.Expand(t)
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.Name, update); err != nil {
		return fmt.Errorf("updating %s: %+v", *id, err)
	}

	d.Partial(false)

	return keyVaultRead(d, meta)
}

func keyVaultRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).KeyVault.VaultsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VaultID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state!", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}
	if resp.Properties == nil {
		return fmt.Errorf("retrieving %s: `properties` was nil", *id)
	}
	if resp.Properties.VaultURI == nil {
		return fmt.Errorf("retrieving %s: `properties.VaultUri` was nil", *id)
	}

	props := *resp.Properties
	meta.(*clients.Client).KeyVault.AddToCache(*id, *resp.Properties.VaultURI)

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	d.Set("tenant_id", props.TenantID.String())
	d.Set("enabled_for_deployment", props.EnabledForDeployment)
	d.Set("enabled_for_disk_encryption", props.EnabledForDiskEncryption)
	d.Set("enabled_for_template_deployment", props.EnabledForTemplateDeployment)
	d.Set("enable_rbac_authorization", props.EnableRbacAuthorization)
	d.Set("vault_uri", props.VaultURI)

	skuName := ""
	if sku := props.Sku; sku != nil {
		// the Azure API is inconsistent here, so rewrite this into the casing we expect
		for _, v := range keyvault.PossibleSkuNameValues() {
			if strings.EqualFold(string(v), string(sku.Name)) {
				skuName = string(v)
			}
		}
	}
	d.Set("sku_name", skuName)

	if err := d.Set("network_acls", flattenKeyVaultNetworkAcls(props.NetworkAcls)); err != nil {
		return fmt.Errorf("setting `network_acls` for KeyVault %q: %+v", *resp.Name, err)
	}

	flattenedPolicies := flattenAccessPolicies(props.AccessPolicies)
	if err := d.Set("access_policy", flattenedPolicies); err != nil {
		return fmt.Errorf("setting `access_policy` for KeyVault %q: %+v", *resp.Name, err)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func keyVaultDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).KeyVault.VaultsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VaultID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.Name, keyVaultResourceName)
	defer locks.UnlockByName(id.Name, keyVaultResourceName)

	read, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	if read.Properties == nil {
		return fmt.Errorf("retrieving %q: `properties` was nil", *id)
	}
	if read.Location == nil {
		return fmt.Errorf("retrieving %q: `location` was nil", *id)
	}

	// ensure we lock on the latest network names, to ensure we handle Azure's networking layer being limited to one change at a time
	virtualNetworkNames := make([]string, 0)
	if props := read.Properties; props != nil {
		if acls := props.NetworkAcls; acls != nil {
			if rules := acls.VirtualNetworkRules; rules != nil {
				for _, v := range *rules {
					if v.ID == nil {
						continue
					}

					subnetId, err := networkParse.SubnetIDInsensitively(*v.ID)
					if err != nil {
						return err
					}

					if !utils.SliceContainsValue(virtualNetworkNames, subnetId.VirtualNetworkName) {
						virtualNetworkNames = append(virtualNetworkNames, subnetId.VirtualNetworkName)
					}
				}
			}
		}
	}

	locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
	defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

	resp, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.WasNotFound(resp.Response) {
			return fmt.Errorf("retrieving %s: %+v", *id, err)
		}
	}

	meta.(*clients.Client).KeyVault.Purge(*id)

	return nil
}

func keyVaultRefreshFunc(vaultUri string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Checking to see if KeyVault %q is available..", vaultUri)

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}

		conn, err := client.Get(vaultUri)
		if err != nil {
			log.Printf("[DEBUG] Didn't find KeyVault at %q", vaultUri)
			return nil, "pending", fmt.Errorf("Error connecting to %q: %s", vaultUri, err)
		}

		defer conn.Body.Close()

		log.Printf("[DEBUG] Found KeyVault at %q", vaultUri)
		return "available", "available", nil
	}
}

func expandKeyVaultNetworkAcls(input []interface{}) (*keyvault.NetworkRuleSet, []string) {
	subnetIds := make([]string, 0)
	if len(input) == 0 {
		return nil, subnetIds
	}

	v := input[0].(map[string]interface{})

	bypass := v["bypass"].(string)
	defaultAction := v["default_action"].(string)

	ipRulesRaw := v["ip_rules"].(*schema.Set)
	ipRules := make([]keyvault.IPRule, 0)

	for _, v := range ipRulesRaw.List() {
		rule := keyvault.IPRule{
			Value: utils.String(v.(string)),
		}
		ipRules = append(ipRules, rule)
	}

	networkRulesRaw := v["virtual_network_subnet_ids"].(*schema.Set)
	networkRules := make([]keyvault.VirtualNetworkRule, 0)
	for _, v := range networkRulesRaw.List() {
		rawId := v.(string)
		subnetIds = append(subnetIds, rawId)
		rule := keyvault.VirtualNetworkRule{
			ID: utils.String(rawId),
		}
		networkRules = append(networkRules, rule)
	}

	ruleSet := keyvault.NetworkRuleSet{
		Bypass:              keyvault.NetworkRuleBypassOptions(bypass),
		DefaultAction:       keyvault.NetworkRuleAction(defaultAction),
		IPRules:             &ipRules,
		VirtualNetworkRules: &networkRules,
	}
	return &ruleSet, subnetIds
}

func flattenKeyVaultNetworkAcls(input *keyvault.NetworkRuleSet) []interface{} {
	if input == nil {
		return []interface{}{
			map[string]interface{}{
				"bypass":                     string(keyvault.AzureServices),
				"default_action":             string(keyvault.Allow),
				"ip_rules":                   schema.NewSet(schema.HashString, []interface{}{}),
				"virtual_network_subnet_ids": schema.NewSet(schema.HashString, []interface{}{}),
			},
		}
	}

	output := make(map[string]interface{})

	output["bypass"] = string(input.Bypass)
	output["default_action"] = string(input.DefaultAction)

	ipRules := make([]interface{}, 0)
	if input.IPRules != nil {
		for _, v := range *input.IPRules {
			if v.Value == nil {
				continue
			}

			ipRules = append(ipRules, *v.Value)
		}
	}
	output["ip_rules"] = schema.NewSet(schema.HashString, ipRules)

	virtualNetworkRules := make([]interface{}, 0)
	if input.VirtualNetworkRules != nil {
		for _, v := range *input.VirtualNetworkRules {
			if v.ID == nil {
				continue
			}

			id := *v.ID
			subnetId, err := networkParse.SubnetIDInsensitively(*v.ID)
			if err == nil {
				id = subnetId.ID()
			}

			virtualNetworkRules = append(virtualNetworkRules, id)
		}
	}
	output["virtual_network_subnet_ids"] = schema.NewSet(schema.HashString, virtualNetworkRules)

	return []interface{}{output}
}
