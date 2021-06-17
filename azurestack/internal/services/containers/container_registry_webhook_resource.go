package containers

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/mgmt/2020-11-01-preview/containerregistry"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/containers/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceContainerRegistryWebhook() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceContainerRegistryWebhookCreate,
		Read:   resourceContainerRegistryWebhookRead,
		Update: resourceContainerRegistryWebhookUpdate,
		Delete: resourceContainerRegistryWebhookDelete,

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
				ValidateFunc: validate.ContainerRegistryWebhookName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"registry_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ContainerRegistryName,
			},

			"service_uri": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ContainerRegistryWebhookServiceUri,
			},

			"custom_headers": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"status": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  containerregistry.WebhookStatusEnabled,
				ValidateFunc: validation.StringInSlice([]string{
					string(containerregistry.WebhookStatusDisabled),
					string(containerregistry.WebhookStatusEnabled),
				}, false),
			},

			"scope": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  "",
			},

			"actions": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(containerregistry.ChartDelete),
						string(containerregistry.ChartPush),
						string(containerregistry.Delete),
						string(containerregistry.Push),
						string(containerregistry.Quarantine),
					}, false),
				},
			},

			"location": azure.SchemaLocation(),

			"tags": tags.Schema(),
		},
	}
}

func resourceContainerRegistryWebhookCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.WebhooksClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for  Container Registry Webhook creation.")

	resourceGroup := d.Get("resource_group_name").(string)
	registryName := d.Get("registry_name").(string)
	name := d.Get("name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, registryName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Container Registry Webhook %q (Resource Group %q, Registry %q): %s", name, resourceGroup, registryName, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_container_registry_webhook", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	webhook := containerregistry.WebhookCreateParameters{
		Location:                          &location,
		WebhookPropertiesCreateParameters: expandWebhookPropertiesCreateParameters(d),
		Tags:                              tags.Expand(t),
	}

	future, err := client.Create(ctx, resourceGroup, registryName, name, webhook)
	if err != nil {
		return fmt.Errorf("Error creating Container Registry Webhook %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation of Container Registry %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	read, err := client.Get(ctx, resourceGroup, registryName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Container Registry %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Container Registry %q (resource group %q, Registry %q) ID", name, resourceGroup, registryName)
	}

	d.SetId(*read.ID)

	return resourceContainerRegistryWebhookRead(d, meta)
}

func resourceContainerRegistryWebhookUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.WebhooksClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for  Container Registry Webhook update.")

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	registryName := id.Path["registries"]
	name := id.Path["webhooks"]

	t := d.Get("tags").(map[string]interface{})

	webhook := containerregistry.WebhookUpdateParameters{
		WebhookPropertiesUpdateParameters: expandWebhookPropertiesUpdateParameters(d),
		Tags:                              tags.Expand(t),
	}

	future, err := client.Update(ctx, resourceGroup, registryName, name, webhook)
	if err != nil {
		return fmt.Errorf("Error updating Container Registry Webhook %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion of Container Registry Webhook %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	return resourceContainerRegistryWebhookRead(d, meta)
}

func resourceContainerRegistryWebhookRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.WebhooksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	registryName := id.Path["registries"]
	name := id.Path["webhooks"]

	resp, err := client.Get(ctx, resourceGroup, registryName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Container Registry Webhook %q was not found in Resource Group %q for Registry %q", name, resourceGroup, registryName)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Azure Container Registry Webhook %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	callbackConfig, err := client.GetCallbackConfig(ctx, resourceGroup, registryName, name)
	if err != nil {
		return fmt.Errorf("Error making Read request on Azure Container Registry Webhook Callback Config %q (Resource Group %q, Registry %q): %+v", name, resourceGroup, registryName, err)
	}

	d.Set("resource_group_name", resourceGroup)
	d.Set("registry_name", registryName)
	d.Set("name", name)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	d.Set("service_uri", callbackConfig.ServiceURI)

	if callbackConfig.CustomHeaders != nil {
		customHeaders := make(map[string]string)
		for k, v := range callbackConfig.CustomHeaders {
			customHeaders[k] = *v
		}
		d.Set("custom_headers", customHeaders)
	}

	if webhookProps := resp.WebhookProperties; webhookProps != nil {
		if webhookProps.Status != "" {
			d.Set("status", string(webhookProps.Status))
		}

		if webhookProps.Scope != nil {
			d.Set("scope", webhookProps.Scope)
		}

		webhookActions := make([]string, len(*webhookProps.Actions))
		for i, action := range *webhookProps.Actions {
			webhookActions[i] = string(action)
		}
		d.Set("actions", webhookActions)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceContainerRegistryWebhookDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Containers.WebhooksClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	registryName := id.Path["registries"]
	name := id.Path["webhooks"]

	future, err := client.Delete(ctx, resourceGroup, registryName, name)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("Error issuing Azure ARM delete request of Container Registry Webhook '%s': %+v", name, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("Error issuing Azure ARM delete request of Container Registry Webhook '%s': %+v", name, err)
	}

	return nil
}

func expandWebhookPropertiesCreateParameters(d *pluginsdk.ResourceData) *containerregistry.WebhookPropertiesCreateParameters {
	serviceUri := d.Get("service_uri").(string)
	scope := d.Get("scope").(string)

	customHeaders := make(map[string]*string)
	for k, v := range d.Get("custom_headers").(map[string]interface{}) {
		customHeaders[k] = utils.String(v.(string))
	}

	actions := expandWebhookActions(d)

	webhookProperties := containerregistry.WebhookPropertiesCreateParameters{
		ServiceURI:    &serviceUri,
		CustomHeaders: customHeaders,
		Actions:       actions,
		Scope:         &scope,
	}

	webhookProperties.Status = containerregistry.WebhookStatus(d.Get("status").(string))

	return &webhookProperties
}

func expandWebhookPropertiesUpdateParameters(d *pluginsdk.ResourceData) *containerregistry.WebhookPropertiesUpdateParameters {
	serviceUri := d.Get("service_uri").(string)
	scope := d.Get("scope").(string)

	customHeaders := make(map[string]*string)
	for k, v := range d.Get("custom_headers").(map[string]interface{}) {
		customHeaders[k] = utils.String(v.(string))
	}

	webhookProperties := containerregistry.WebhookPropertiesUpdateParameters{
		ServiceURI:    &serviceUri,
		CustomHeaders: customHeaders,
		Actions:       expandWebhookActions(d),
		Scope:         &scope,
		Status:        containerregistry.WebhookStatus(d.Get("status").(string)),
	}

	return &webhookProperties
}

func expandWebhookActions(d *pluginsdk.ResourceData) *[]containerregistry.WebhookAction {
	actions := make([]containerregistry.WebhookAction, 0)
	for _, action := range d.Get("actions").(*pluginsdk.Set).List() {
		actions = append(actions, containerregistry.WebhookAction(action.(string)))
	}

	return &actions
}
