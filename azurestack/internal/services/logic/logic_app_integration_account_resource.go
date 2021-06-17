package logic

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/logic/mgmt/2019-05-01/logic"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/location"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/logic/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/logic/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceLogicAppIntegrationAccount() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceLogicAppIntegrationAccountCreateUpdate,
		Read:   resourceLogicAppIntegrationAccountRead,
		Update: resourceLogicAppIntegrationAccountCreateUpdate,
		Delete: resourceLogicAppIntegrationAccountDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.IntegrationAccountID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.IntegrationAccountName(),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"sku_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(logic.IntegrationAccountSkuNameBasic),
					string(logic.IntegrationAccountSkuNameFree),
					string(logic.IntegrationAccountSkuNameStandard),
				}, false),
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceLogicAppIntegrationAccountCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Logic.IntegrationAccountClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing Integration Account %q (Resource Group %q): %+v", name, resourceGroup, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_logic_app_integration_account", *existing.ID)
		}
	}

	account := logic.IntegrationAccount{
		IntegrationAccountProperties: &logic.IntegrationAccountProperties{},
		Location:                     utils.String(location.Normalize(d.Get("location").(string))),
		Sku: &logic.IntegrationAccountSku{
			Name: logic.IntegrationAccountSkuName(d.Get("sku_name").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, name, account); err != nil {
		return fmt.Errorf("creating Integration Account %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("retrieving Integration Account %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("reading Integration Account %q (Resource Group %q): ID is empty or nil", name, resourceGroup)
	}

	d.SetId(*resp.ID)

	return resourceLogicAppIntegrationAccountRead(d, meta)
}

func resourceLogicAppIntegrationAccountRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Logic.IntegrationAccountClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.IntegrationAccountID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Integration Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("sku_name", string(resp.Sku.Name))

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceLogicAppIntegrationAccountDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Logic.IntegrationAccountClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.IntegrationAccountID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		return fmt.Errorf("deleting Integration Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}
