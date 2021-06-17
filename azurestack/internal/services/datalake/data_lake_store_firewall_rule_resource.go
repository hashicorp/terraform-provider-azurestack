package datalake

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datalake/store/mgmt/2016-11-01/account"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	commonValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/datalake/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceDataLakeStoreFirewallRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmDateLakeStoreAccountFirewallRuleCreateUpdate,
		Read:   resourceArmDateLakeStoreAccountFirewallRuleRead,
		Update: resourceArmDateLakeStoreAccountFirewallRuleCreateUpdate,
		Delete: resourceArmDateLakeStoreAccountFirewallRuleDelete,
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
				ValidateFunc: validate.FirewallRuleName(),
			},

			"account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AccountName(),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"start_ip_address": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: commonValidate.IPv4Address,
			},

			"end_ip_address": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: commonValidate.IPv4Address,
			},
		},
	}
}

func resourceArmDateLakeStoreAccountFirewallRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Datalake.StoreFirewallRulesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	accountName := d.Get("account_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, accountName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Date Lake Store Firewall Rule %q (Account %q / Resource Group %q): %s", name, accountName, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_data_lake_store_firewall_rule", *existing.ID)
		}
	}

	startIPAddress := d.Get("start_ip_address").(string)
	endIPAddress := d.Get("end_ip_address").(string)

	log.Printf("[INFO] preparing arguments for Date Lake Store Firewall Rule creation  %q (Resource Group %q)", name, resourceGroup)

	dateLakeStore := account.CreateOrUpdateFirewallRuleParameters{
		CreateOrUpdateFirewallRuleProperties: &account.CreateOrUpdateFirewallRuleProperties{
			StartIPAddress: utils.String(startIPAddress),
			EndIPAddress:   utils.String(endIPAddress),
		},
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, accountName, name, dateLakeStore); err != nil {
		return fmt.Errorf("Error issuing create request for Data Lake Store %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, accountName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Data Lake Store Firewall Rule %q (Account %q / Resource Group %q): %+v", name, accountName, resourceGroup, err)
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Data Lake Store %q (Account %q / Resource Group %q) ID", name, accountName, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmDateLakeStoreAccountFirewallRuleRead(d, meta)
}

func resourceArmDateLakeStoreAccountFirewallRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Datalake.StoreFirewallRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	accountName := id.Path["accounts"]
	name := id.Path["firewallRules"]

	resp, err := client.Get(ctx, resourceGroup, accountName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[WARN] Data Lake Store Firewall Rule %q was not found (Account %q / Resource Group %q)", name, accountName, resourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure Data Lake Store Firewall Rule %q (Account %q / Resource Group %q): %+v", name, accountName, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("account_name", accountName)
	d.Set("resource_group_name", resourceGroup)

	if props := resp.FirewallRuleProperties; props != nil {
		d.Set("start_ip_address", props.StartIPAddress)
		d.Set("end_ip_address", props.EndIPAddress)
	}

	return nil
}

func resourceArmDateLakeStoreAccountFirewallRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Datalake.StoreFirewallRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	accountName := id.Path["accounts"]
	name := id.Path["firewallRules"]

	resp, err := client.Delete(ctx, resourceGroup, accountName, name)
	if err != nil {
		if response.WasNotFound(resp.Response) {
			return nil
		}
		return fmt.Errorf("Error issuing delete request for Data Lake Store Firewall Rule %q (Account %q / Resource Group %q): %+v", name, accountName, resourceGroup, err)
	}

	return nil
}
