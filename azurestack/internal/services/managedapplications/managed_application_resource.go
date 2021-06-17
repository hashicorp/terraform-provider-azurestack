package managedapplications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/location"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/managedapplications/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/managedapplications/validate"
	resourcesParse "github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/resource/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceManagedApplication() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceManagedApplicationCreateUpdate,
		Read:   resourceManagedApplicationRead,
		Update: resourceManagedApplicationCreateUpdate,
		Delete: resourceManagedApplicationDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ApplicationID(id)
			return err
		}),

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
				ValidateFunc: validate.ApplicationName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"kind": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"MarketPlace",
					"ServiceCatalog",
				}, false),
			},

			"managed_resource_group_name": azure.SchemaResourceGroupName(),

			"application_definition_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.ApplicationDefinitionID,
			},

			"parameters": {
				Type:          pluginsdk.TypeMap,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"parameter_values"},
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"parameter_values": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ConflictsWith:    []string{"parameters"},
			},

			"plan": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"product": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"publisher": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"version": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"promotion_code": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"tags": tags.Schema(),

			"outputs": {
				Type:     pluginsdk.TypeMap,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func resourceManagedApplicationCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagedApplication.ApplicationClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroupName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("failed to check for present of existing Managed Application Name %q (Resource Group %q): %+v", name, resourceGroupName, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_managed_application", *existing.ID)
		}
	}

	parameters := managedapplications.Application{
		Location: utils.String(azure.NormalizeLocation(d.Get("location"))),
		Kind:     utils.String(d.Get("kind").(string)),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("managed_resource_group_name"); ok {
		targetResourceGroupId := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", meta.(*clients.Client).Account.SubscriptionId, v)
		parameters.ApplicationProperties = &managedapplications.ApplicationProperties{
			ManagedResourceGroupID: utils.String(targetResourceGroupId),
		}
	}

	if v, ok := d.GetOk("application_definition_id"); ok {
		parameters.ApplicationDefinitionID = utils.String(v.(string))
	}

	if v, ok := d.GetOk("plan"); ok {
		parameters.Plan = expandManagedApplicationPlan(v.([]interface{}))
	}

	params, err := expandManagedApplicationParameters(d)
	if err != nil {
		return fmt.Errorf("Error expanding `parameters` or `parameter_values`: %+v", err)
	}
	parameters.Parameters = params

	future, err := client.CreateOrUpdate(ctx, resourceGroupName, name, parameters)
	if err != nil {
		return fmt.Errorf("failed to create Managed Application %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("failed to wait for creation of Managed Application %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	resp, err := client.Get(ctx, resourceGroupName, name)
	if err != nil {
		return fmt.Errorf("failed to retrieve Managed Application %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}
	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("cannot read Managed Application %q (Resource Group %q) ID", name, resourceGroupName)
	}
	d.SetId(*resp.ID)

	return resourceManagedApplicationRead(d, meta)
}

func resourceManagedApplicationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagedApplication.ApplicationClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ApplicationID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Managed Application %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to read Managed Application %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("kind", resp.Kind)
	if err := d.Set("plan", flattenManagedApplicationPlan(resp.Plan)); err != nil {
		return fmt.Errorf("setting `plan`: %+v", err)
	}
	if props := resp.ApplicationProperties; props != nil {
		id, err := resourcesParse.ResourceGroupID(*props.ManagedResourceGroupID)
		if err != nil {
			return err
		}

		d.Set("managed_resource_group_name", id.ResourceGroup)
		d.Set("application_definition_id", props.ApplicationDefinitionID)

		parameterValues, err := flattenManagedApplicationParameterValuesValueToString(props.Parameters)
		if err != nil {
			return fmt.Errorf("serializing JSON from `parameter_values`: %+v", err)
		}
		d.Set("parameter_values", parameterValues)

		if err = d.Set("parameters", flattenManagedApplicationParametersOrOutputs(props.Parameters)); err != nil {
			return err
		}

		if err = d.Set("outputs", flattenManagedApplicationParametersOrOutputs(props.Outputs)); err != nil {
			return err
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceManagedApplicationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ManagedApplication.ApplicationClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ApplicationID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("failed to delete Managed Application %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("failed to wait for deleting Managed Application (Managed Application Name %q / Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func expandManagedApplicationPlan(input []interface{}) *managedapplications.Plan {
	if len(input) == 0 {
		return nil
	}
	plan := input[0].(map[string]interface{})

	return &managedapplications.Plan{
		Name:          utils.String(plan["name"].(string)),
		Product:       utils.String(plan["product"].(string)),
		Publisher:     utils.String(plan["publisher"].(string)),
		Version:       utils.String(plan["version"].(string)),
		PromotionCode: utils.String(plan["promotion_code"].(string)),
	}
}

func expandManagedApplicationParameters(d *pluginsdk.ResourceData) (*map[string]interface{}, error) {
	newParams := make(map[string]interface{})

	if v, ok := d.GetOk("parameter_values"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &newParams); err != nil {
			return nil, fmt.Errorf("unmarshalling `parameter_values`: %+v", err)
		}
	}

	if v, ok := d.GetOk("parameters"); ok {
		params := v.(map[string]interface{})

		for key, val := range params {
			newParams[key] = struct {
				Value interface{} `json:"value"`
			}{
				Value: val,
			}
		}
	}

	return &newParams, nil
}

func flattenManagedApplicationPlan(input *managedapplications.Plan) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	name := ""
	if input.Name != nil {
		name = *input.Name
	}
	product := ""
	if input.Product != nil {
		product = *input.Product
	}
	publisher := ""
	if input.Publisher != nil {
		publisher = *input.Publisher
	}
	version := ""
	if input.Version != nil {
		version = *input.Version
	}
	promotionCode := ""
	if input.PromotionCode != nil {
		promotionCode = *input.PromotionCode
	}

	results = append(results, map[string]interface{}{
		"name":           name,
		"product":        product,
		"publisher":      publisher,
		"version":        version,
		"promotion_code": promotionCode,
	})

	return results
}

func flattenManagedApplicationParametersOrOutputs(input interface{}) map[string]interface{} {
	results := make(map[string]interface{})
	if input == nil {
		return results
	}

	for k, v := range input.(map[string]interface{}) {
		if v != nil {
			results[k] = v.(map[string]interface{})["value"].(string)
		}
	}

	return results
}

func flattenManagedApplicationParameterValuesValueToString(input interface{}) (string, error) {
	if input == nil {
		return "", nil
	}

	for k, v := range input.(map[string]interface{}) {
		if v != nil {
			delete(input.(map[string]interface{})[k].(map[string]interface{}), "type")
		}
	}

	result, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	compactJson := bytes.Buffer{}
	if err := json.Compact(&compactJson, result); err != nil {
		return "", err
	}

	return compactJson.String(), nil
}
