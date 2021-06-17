package applicationinsights

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/appinsights/mgmt/2015-05-01/insights"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceApplicationInsightsAPIKey() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApplicationInsightsAPIKeyCreate,
		Read:   resourceApplicationInsightsAPIKeyRead,
		Delete: resourceApplicationInsightsAPIKeyDelete,
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
				ValidateFunc: validation.NoZeroValues,
			},

			"application_insights_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"read_permissions": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				ForceNew: true,
				Set:      pluginsdk.HashString,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"agentconfig", "aggregate", "api", "draft", "extendqueries", "search"}, false),
				},
			},

			"write_permissions": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				ForceNew: true,
				Set:      pluginsdk.HashString,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"annotations"}, false),
				},
			},

			"api_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceApplicationInsightsAPIKeyCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppInsights.APIKeysClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Application Insights API key creation.")

	name := d.Get("name").(string)
	appInsightsID := d.Get("application_insights_id").(string)

	id, err := azure.ParseAzureResourceID(appInsightsID)
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	appInsightsName := id.Path["components"]

	var existingAPIKeyList insights.ApplicationInsightsComponentAPIKeyListResult
	var keyId string
	existingAPIKeyList, err = client.List(ctx, resGroup, appInsightsName)
	if err != nil {
		if !utils.ResponseWasNotFound(existingAPIKeyList.Response) {
			return fmt.Errorf("checking for presence of existing Application Insights API key list (Application Insights %q / Resource Group %q): %s", appInsightsName, resGroup, err)
		}
	}

	for _, existingAPIKey := range *existingAPIKeyList.Value {
		existingAPIKeyId, err := azure.ParseAzureResourceID(*existingAPIKey.ID)
		if err != nil {
			return err
		}

		existingAppInsightsName := existingAPIKeyId.Path["components"]
		if appInsightsName == existingAppInsightsName {
			keyId = existingAPIKeyId.Path["apikeys"]
			break
		}
	}

	var existing insights.ApplicationInsightsComponentAPIKey
	existing, err = client.Get(ctx, resGroup, appInsightsName, keyId)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing Application Insights API key %q (Resource Group %q): %s", name, resGroup, err)
		}
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_application_insights_api_key", *existing.ID)
	}

	apiKeyProperties := insights.APIKeyRequest{
		Name:                  &name,
		LinkedReadProperties:  expandApplicationInsightsAPIKeyLinkedProperties(d.Get("read_permissions").(*pluginsdk.Set), appInsightsID),
		LinkedWriteProperties: expandApplicationInsightsAPIKeyLinkedProperties(d.Get("write_permissions").(*pluginsdk.Set), appInsightsID),
	}

	result, err := client.Create(ctx, resGroup, appInsightsName, apiKeyProperties)
	if err != nil {
		return fmt.Errorf("Error creating Application Insights API key %q (Resource Group %q): %+v", name, resGroup, err)
	}

	if result.APIKey == nil {
		return fmt.Errorf("Error creating Application Insights API key %q (Resource Group %q): got empty API key", name, resGroup)
	}

	d.SetId(*result.ID)

	// API key can only retrieved at key creation
	d.Set("api_key", result.APIKey)

	return resourceApplicationInsightsAPIKeyRead(d, meta)
}

func resourceApplicationInsightsAPIKeyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppInsights.APIKeysClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading AzureRM Application Insights API key '%s'", id)

	resGroup := id.ResourceGroup
	appInsightsName := id.Path["components"]
	keyID := id.Path["apikeys"]

	result, err := client.Get(ctx, resGroup, appInsightsName, keyID)
	if err != nil {
		if utils.ResponseWasNotFound(result.Response) {
			log.Printf("[WARN] AzureRM Application Insights API key '%s' not found, removing from state", id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on AzureRM Application Insights API key '%s': %+v", keyID, err)
	}

	d.Set("application_insights_id", fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/components/%s", client.SubscriptionID, resGroup, appInsightsName))

	d.Set("name", result.Name)
	readProps := flattenApplicationInsightsAPIKeyLinkedProperties(result.LinkedReadProperties)
	if err := d.Set("read_permissions", readProps); err != nil {
		return fmt.Errorf("Error flattening `read_permissions `: %s", err)
	}
	writeProps := flattenApplicationInsightsAPIKeyLinkedProperties(result.LinkedWriteProperties)
	if err := d.Set("write_permissions", writeProps); err != nil {
		return fmt.Errorf("Error flattening `write_permissions `: %s", err)
	}

	return nil
}

func resourceApplicationInsightsAPIKeyDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppInsights.APIKeysClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	appInsightsName := id.Path["components"]
	keyID := id.Path["apikeys"]

	log.Printf("[DEBUG] Deleting AzureRM Application Insights API key '%s' (resource group '%s')", keyID, resGroup)

	result, err := client.Delete(ctx, resGroup, appInsightsName, keyID)
	if err != nil {
		if utils.ResponseWasNotFound(result.Response) {
			return nil
		}
		return fmt.Errorf("Error issuing AzureRM delete request for Application Insights API key '%s': %+v", keyID, err)
	}

	return nil
}
