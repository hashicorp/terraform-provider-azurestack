package recoveryservices

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2016-06-01/recoveryservices"
	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2019-05-13/backup"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/recoveryservices/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceRecoveryServicesVault() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceRecoveryServicesVaultCreateUpdate,
		Read:   resourceRecoveryServicesVaultRead,
		Update: resourceRecoveryServicesVaultCreateUpdate,
		Delete: resourceRecoveryServicesVaultDelete,

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.RecoveryServicesVaultName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"identity": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"type": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(recoveryservices.SystemAssigned),
							}, false),
						},

						"principal_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},

						"tenant_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.Schema(),

			"sku": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(recoveryservices.RS0),
					string(recoveryservices.Standard),
				}, true),
			},

			"soft_delete_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceRecoveryServicesVaultCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.VaultsClient
	cfgsClient := meta.(*clients.Client).RecoveryServices.VaultsConfigsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	t := d.Get("tags").(map[string]interface{})

	log.Printf("[DEBUG] Creating/updating Recovery Service Vault %q (resource group %q)", name, resourceGroup)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Recovery Service Vault %q (Resource Group %q): %+v", name, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_recovery_services_vault", *existing.ID)
		}
	}

	vault := recoveryservices.Vault{
		Location: utils.String(location),
		Tags:     tags.Expand(t),
		Identity: expandValutIdentity(d.Get("identity").([]interface{})),
		Sku: &recoveryservices.Sku{
			Name: recoveryservices.SkuName(d.Get("sku").(string)),
		},
		Properties: &recoveryservices.VaultProperties{},
	}

	vault, err := client.CreateOrUpdate(ctx, resourceGroup, name, vault)
	if err != nil {
		return fmt.Errorf("Error creating/updating Recovery Service Vault %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	cfg := backup.ResourceVaultConfigResource{
		Properties: &backup.ResourceVaultConfig{
			EnhancedSecurityState: backup.EnhancedSecurityStateEnabled, // always enabled
		},
	}

	if sd := d.Get("soft_delete_enabled").(bool); sd {
		cfg.Properties.SoftDeleteFeatureState = backup.SoftDeleteFeatureStateEnabled
	} else {
		cfg.Properties.SoftDeleteFeatureState = backup.SoftDeleteFeatureStateDisabled
	}

	stateConf := &pluginsdk.StateChangeConf{
		Pending:    []string{"syncing"},
		Target:     []string{"success"},
		MinTimeout: 30 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := cfgsClient.Update(ctx, name, resourceGroup, cfg)
			if err != nil {
				if strings.Contains(err.Error(), "ResourceNotYetSynced") {
					return resp, "syncing", nil
				}
				return resp, "error", fmt.Errorf("Error updating Recovery Service Vault Cfg %q (Resource Group %q): %+v", name, resourceGroup, err)
			}

			return resp, "success", nil
		},
	}

	if d.IsNewResource() {
		stateConf.Timeout = d.Timeout(pluginsdk.TimeoutCreate)
	} else {
		stateConf.Timeout = d.Timeout(pluginsdk.TimeoutUpdate)
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for on update for Recovery Service Vault %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("Error issuing read request for Recovery Service Vault %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if read.ID == nil {
		return fmt.Errorf("Error Recovery Service Vault %q (Resource Group %q): read returned nil", name, resourceGroup)
	}

	d.SetId(*vault.ID)

	return resourceRecoveryServicesVaultRead(d, meta)
}

func resourceRecoveryServicesVaultRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.VaultsClient
	cfgsClient := meta.(*clients.Client).RecoveryServices.VaultsConfigsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["vaults"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Reading Recovery Service Vault %q (resource group %q)", name, resourceGroup)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Recovery Service Vault %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku", string(sku.Name))
	}

	cfg, err := cfgsClient.Get(ctx, name, resourceGroup)
	if err != nil {
		return fmt.Errorf("Error reading Recovery Service Vault Cfg %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if props := cfg.Properties; props != nil {
		d.Set("soft_delete_enabled", props.SoftDeleteFeatureState == backup.SoftDeleteFeatureStateEnabled)
	}

	if err := d.Set("identity", flattenVaultIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("Error setting `identity`: %+v", err)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceRecoveryServicesVaultDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RecoveryServices.VaultsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["vaults"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Deleting Recovery Service Vault %q (resource group %q)", name, resourceGroup)

	resp, err := client.Delete(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("Error issuing delete request for Recovery Service Vault %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	return nil
}

func expandValutIdentity(input []interface{}) *recoveryservices.IdentityData {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})
	return &recoveryservices.IdentityData{
		Type: recoveryservices.ResourceIdentityType(v["type"].(string)),
	}
}

func flattenVaultIdentity(input *recoveryservices.IdentityData) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	principalID := ""
	if input.PrincipalID != nil {
		principalID = *input.PrincipalID
	}

	tenantID := ""
	if input.TenantID != nil {
		tenantID = *input.TenantID
	}

	return []interface{}{
		map[string]interface{}{
			"type":         string(input.Type),
			"principal_id": principalID,
			"tenant_id":    tenantID,
		},
	}
}
