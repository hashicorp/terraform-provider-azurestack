package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/logic/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
)

func resourceLogicAppTriggerHttpRequest() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceLogicAppTriggerHttpRequestCreateUpdate,
		Read:   resourceLogicAppTriggerHttpRequestRead,
		Update: resourceLogicAppTriggerHttpRequestCreateUpdate,
		Delete: resourceLogicAppTriggerHttpRequestDelete,
		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: pluginsdk.CustomizeDiffShim(func(ctx context.Context, diff *pluginsdk.ResourceDiff, v interface{}) error {
			relativePath := diff.Get("relative_path").(string)
			if relativePath != "" {
				method := diff.Get("method").(string)
				if method == "" {
					return fmt.Errorf("`method` must be specified when `relative_path` is set.")
				}
			}

			return nil
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"logic_app_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"schema": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},

			"method": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					http.MethodDelete,
					http.MethodGet,
					http.MethodPatch,
					http.MethodPost,
					http.MethodPut,
				}, false),
			},

			"relative_path": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.TriggerHttpRequestRelativePath,
			},
		},
	}
}

func resourceLogicAppTriggerHttpRequestCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	schemaRaw := d.Get("schema").(string)
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return fmt.Errorf("Error unmarshalling JSON from Schema: %+v", err)
	}

	inputs := map[string]interface{}{
		"schema": schema,
	}

	if v, ok := d.GetOk("method"); ok {
		inputs["method"] = v.(string)
	}

	if v, ok := d.GetOk("relative_path"); ok {
		inputs["relativePath"] = v.(string)
	}

	trigger := map[string]interface{}{
		"inputs": inputs,
		"kind":   "Http",
		"type":   "Request",
	}

	logicAppId := d.Get("logic_app_id").(string)
	name := d.Get("name").(string)
	if err := resourceLogicAppTriggerUpdate(d, meta, logicAppId, name, trigger, "azurerm_logic_app_trigger_http_request"); err != nil {
		return err
	}

	return resourceLogicAppTriggerHttpRequestRead(d, meta)
}

func resourceLogicAppTriggerHttpRequestRead(d *pluginsdk.ResourceData, meta interface{}) error {
	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	logicAppName := id.Path["workflows"]
	name := id.Path["triggers"]

	t, app, err := retrieveLogicAppTrigger(d, meta, resourceGroup, logicAppName, name)
	if err != nil {
		return err
	}

	if t == nil {
		log.Printf("[DEBUG] Logic App %q (Resource Group %q) does not contain Trigger %q - removing from state", logicAppName, resourceGroup, name)
		d.SetId("")
		return nil
	}

	trigger := *t

	d.Set("name", name)
	d.Set("logic_app_id", app.ID)

	v := trigger["inputs"]
	if v == nil {
		return fmt.Errorf("Error `inputs` was nil for HTTP Trigger %q (Logic App %q / Resource Group %q)", name, logicAppName, resourceGroup)
	}

	inputs, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Error parsing `inputs` for HTTP Trigger %q (Logic App %q / Resource Group %q)", name, logicAppName, resourceGroup)
	}

	if method := inputs["method"]; method != nil {
		d.Set("method", method.(string))
	}

	if relativePath := inputs["relativePath"]; relativePath != nil {
		d.Set("relative_path", relativePath.(string))
	}

	if schemaRaw := inputs["schema"]; schemaRaw != nil {
		schema, err := json.Marshal(schemaRaw)
		if err != nil {
			return fmt.Errorf("Error serializing the Schema to JSON: %+v", err)
		}

		d.Set("schema", string(schema))
	}

	return nil
}

func resourceLogicAppTriggerHttpRequestDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	logicAppName := id.Path["workflows"]
	name := id.Path["triggers"]

	err = resourceLogicAppTriggerRemove(d, meta, resourceGroup, logicAppName, name)
	if err != nil {
		return fmt.Errorf("Error removing Trigger %q from Logic App %q (Resource Group %q): %+v", name, logicAppName, resourceGroup, err)
	}

	return nil
}
