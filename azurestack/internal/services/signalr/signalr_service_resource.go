package signalr

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/signalr/mgmt/2020-05-01/signalr"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/signalr/parse"
	signalrValidate "github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/signalr/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceArmSignalRService() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmSignalRServiceCreate,
		Read:   resourceArmSignalRServiceRead,
		Update: resourceArmSignalRServiceUpdate,
		Delete: resourceArmSignalRServiceDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ServiceID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Free_F1",
								"Standard_S1",
							}, false),
						},

						"capacity": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntInSlice([]int{1, 2, 5, 10, 20, 50, 100}),
						},
					},
				},
			},

			"features": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"flag": {
							Type:     pluginsdk.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								// Looks like the default has changed, ours will need to be updated in AzureRM 3.0.
								// issue has been created https://github.com/Azure/azure-sdk-for-go/issues/9619
								"EnableMessagingLogs",
								string(signalr.EnableConnectivityLogs),
								string(signalr.ServiceMode),
							}, false),
						},

						"value": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
					},
				},
			},

			"upstream_endpoint": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"category_pattern": {
							Type:     pluginsdk.TypeList,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},

						"event_pattern": {
							Type:     pluginsdk.TypeList,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},

						"hub_pattern": {
							Type:     pluginsdk.TypeList,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},

						"url_template": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: signalrValidate.UrlTemplate,
						},
					},
				},
			},

			"cors": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"allowed_origins": {
							Type:     pluginsdk.TypeSet,
							Required: true,
							Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
						},
					},
				},
			},

			"hostname": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"ip_address": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"public_port": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"server_port": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"primary_access_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_access_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmSignalRServiceCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SignalR.Client
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	resourceGroup := d.Get("resource_group_name").(string)

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("Error checking for presence of existing SignalR %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_signalr_service", *existing.ID)
	}

	sku := d.Get("sku").([]interface{})
	t := d.Get("tags").(map[string]interface{})
	featureFlags := d.Get("features").(*pluginsdk.Set).List()
	cors := d.Get("cors").([]interface{})
	expandedTags := tags.Expand(t)
	upstreamSettings := d.Get("upstream_endpoint").(*pluginsdk.Set).List()

	expandedFeatures := expandSignalRFeatures(featureFlags)

	// Upstream configurations are only allowed when the SignalR service is in `Serverless` mode
	if len(upstreamSettings) > 0 && !signalRIsInServerlessMode(expandedFeatures) {
		return fmt.Errorf("Upstream configurations are only allowed when the SignalR Service is in `Serverless` mode")
	}

	properties := &signalr.Properties{
		Cors:     expandSignalRCors(cors),
		Features: expandedFeatures,
		Upstream: expandUpstreamSettings(upstreamSettings),
	}

	resourceType := &signalr.ResourceType{
		Location:   utils.String(location),
		Sku:        expandSignalRServiceSku(sku),
		Tags:       expandedTags,
		Properties: properties,
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, name, resourceType)
	if err != nil {
		return fmt.Errorf("Error creating or updating SignalR %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for the result of creating or updating SignalR %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("SignalR Service %q (Resource Group %q) ID is empty", name, resourceGroup)
	}
	d.SetId(*read.ID)

	return resourceArmSignalRServiceUpdate(d, meta)
}

func resourceArmSignalRServiceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SignalR.Client
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.SignalRName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] SignalR %q was not found in Resource Group %q - removing from state!", id.SignalRName, id.ResourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error getting SignalR %q (Resource Group %q): %+v", id.SignalRName, id.ResourceGroup, err)
	}

	keys, err := client.ListKeys(ctx, id.ResourceGroup, id.SignalRName)
	if err != nil {
		return fmt.Errorf("Error getting keys of SignalR %q (Resource Group %q): %+v", id.SignalRName, id.ResourceGroup, err)
	}

	d.Set("name", id.SignalRName)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if err = d.Set("sku", flattenSignalRServiceSku(resp.Sku)); err != nil {
		return fmt.Errorf("Error setting `sku`: %+v", err)
	}

	if properties := resp.Properties; properties != nil {
		d.Set("hostname", properties.HostName)
		d.Set("ip_address", properties.ExternalIP)
		d.Set("public_port", properties.PublicPort)
		d.Set("server_port", properties.ServerPort)

		if err := d.Set("features", flattenSignalRFeatures(properties.Features)); err != nil {
			return fmt.Errorf("Error setting `features`: %+v", err)
		}

		if err := d.Set("cors", flattenSignalRCors(properties.Cors)); err != nil {
			return fmt.Errorf("Error setting `cors`: %+v", err)
		}

		if err := d.Set("upstream_endpoint", flattenUpstreamSettings(properties.Upstream)); err != nil {
			return fmt.Errorf("Error setting `upstream_endpoint`: %+v", err)
		}
	}

	d.Set("primary_access_key", keys.PrimaryKey)
	d.Set("primary_connection_string", keys.PrimaryConnectionString)
	d.Set("secondary_access_key", keys.SecondaryKey)
	d.Set("secondary_connection_string", keys.SecondaryConnectionString)

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmSignalRServiceUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SignalR.Client
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceID(d.Id())
	if err != nil {
		return err
	}

	resourceType := &signalr.ResourceType{}

	if d.HasChanges("cors", "features", "upstream_endpoint") {
		resourceType.Properties = &signalr.Properties{}

		if d.HasChange("cors") {
			corsRaw := d.Get("cors").([]interface{})
			resourceType.Properties.Cors = expandSignalRCors(corsRaw)
		}

		if d.HasChange("features") {
			featuresRaw := d.Get("features").(*pluginsdk.Set).List()
			resourceType.Properties.Features = expandSignalRFeatures(featuresRaw)
		}

		if d.HasChange("upstream_endpoint") {
			featuresRaw := d.Get("upstream_endpoint").(*pluginsdk.Set).List()
			resourceType.Properties.Upstream = expandUpstreamSettings(featuresRaw)
		}
	}

	if d.HasChange("sku") {
		sku := d.Get("sku").([]interface{})
		resourceType.Sku = expandSignalRServiceSku(sku)
	}

	if d.HasChange("tags") {
		tagsRaw := d.Get("tags").(map[string]interface{})
		resourceType.Tags = tags.Expand(tagsRaw)
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.SignalRName, resourceType)
	if err != nil {
		return fmt.Errorf("updating SignalR Service %q (Resource Group %q): %+v", id.SignalRName, id.ResourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the update of SignalR Service %q (Resource Group %q): %+v", id.SignalRName, id.ResourceGroup, err)
	}

	return resourceArmSignalRServiceRead(d, meta)
}

func resourceArmSignalRServiceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).SignalR.Client
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.SignalRName)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("deleting SignalR Service %q (Resource Group %q): %+v", id.SignalRName, id.ResourceGroup, err)
		}
		return nil
	}
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for the deletion of SignalR Service %q (Resource Group %q): %+v", id.SignalRName, id.ResourceGroup, err)
		}
	}

	return nil
}

func signalRIsInServerlessMode(features *[]signalr.Feature) bool {
	if features == nil {
		return false
	}

	for _, feature := range *features {
		if feature.Flag == signalr.ServiceMode && feature.Value != nil {
			return *feature.Value == "Serverless"
		}
	}

	return false
}

func expandSignalRFeatures(input []interface{}) *[]signalr.Feature {
	features := make([]signalr.Feature, 0)
	for _, featureValue := range input {
		value := featureValue.(map[string]interface{})

		feature := signalr.Feature{
			Flag:  signalr.FeatureFlags(value["flag"].(string)),
			Value: utils.String(value["value"].(string)),
		}

		features = append(features, feature)
	}

	return &features
}

func flattenSignalRFeatures(features *[]signalr.Feature) []interface{} {
	if features == nil {
		return []interface{}{}
	}

	result := make([]interface{}, 0)
	for _, feature := range *features {
		value := ""
		if feature.Value != nil {
			value = *feature.Value
		}

		result = append(result, map[string]interface{}{
			"flag":  string(feature.Flag),
			"value": value,
		})
	}
	return result
}

func expandUpstreamSettings(input []interface{}) *signalr.ServerlessUpstreamSettings {
	upstreamTemplates := make([]signalr.UpstreamTemplate, 0)

	for _, upstreamSetting := range input {
		setting := upstreamSetting.(map[string]interface{})

		upstreamTemplate := signalr.UpstreamTemplate{
			HubPattern:      utils.String(strings.Join(*utils.ExpandStringSlice(setting["hub_pattern"].([]interface{})), ",")),
			EventPattern:    utils.String(strings.Join(*utils.ExpandStringSlice(setting["event_pattern"].([]interface{})), ",")),
			CategoryPattern: utils.String(strings.Join(*utils.ExpandStringSlice(setting["category_pattern"].([]interface{})), ",")),
			URLTemplate:     utils.String(setting["url_template"].(string)),
		}

		upstreamTemplates = append(upstreamTemplates, upstreamTemplate)
	}

	return &signalr.ServerlessUpstreamSettings{
		Templates: &upstreamTemplates,
	}
}

func flattenUpstreamSettings(upstreamSettings *signalr.ServerlessUpstreamSettings) []interface{} {
	result := make([]interface{}, 0)
	if upstreamSettings == nil || upstreamSettings.Templates == nil {
		return result
	}

	for _, settings := range *upstreamSettings.Templates {
		categoryPattern := make([]interface{}, 0)
		if settings.CategoryPattern != nil {
			categoryPatterns := strings.Split(*settings.CategoryPattern, ",")
			categoryPattern = utils.FlattenStringSlice(&categoryPatterns)
		}

		eventPattern := make([]interface{}, 0)
		if settings.EventPattern != nil {
			eventPatterns := strings.Split(*settings.EventPattern, ",")
			eventPattern = utils.FlattenStringSlice(&eventPatterns)
		}

		hubPattern := make([]interface{}, 0)
		if settings.HubPattern != nil {
			hubPatterns := strings.Split(*settings.HubPattern, ",")
			hubPattern = utils.FlattenStringSlice(&hubPatterns)
		}

		urlTemplate := ""
		if settings.URLTemplate != nil {
			urlTemplate = *settings.URLTemplate
		}

		result = append(result, map[string]interface{}{
			"url_template":     urlTemplate,
			"hub_pattern":      hubPattern,
			"event_pattern":    eventPattern,
			"category_pattern": categoryPattern,
		})
	}
	return result
}

func expandSignalRCors(input []interface{}) *signalr.CorsSettings {
	corsSettings := signalr.CorsSettings{}

	if len(input) == 0 || input[0] == nil {
		return &corsSettings
	}

	setting := input[0].(map[string]interface{})
	origins := setting["allowed_origins"].(*pluginsdk.Set).List()

	allowedOrigins := make([]string, 0)
	for _, param := range origins {
		allowedOrigins = append(allowedOrigins, param.(string))
	}

	corsSettings.AllowedOrigins = &allowedOrigins

	return &corsSettings
}

func flattenSignalRCors(input *signalr.CorsSettings) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	result := make(map[string]interface{})

	allowedOrigins := make([]interface{}, 0)
	if s := input.AllowedOrigins; s != nil {
		for _, v := range *s {
			allowedOrigins = append(allowedOrigins, v)
		}
	}
	result["allowed_origins"] = pluginsdk.NewSet(pluginsdk.HashString, allowedOrigins)

	return append(results, result)
}

func expandSignalRServiceSku(input []interface{}) *signalr.ResourceSku {
	v := input[0].(map[string]interface{})
	return &signalr.ResourceSku{
		Name:     utils.String(v["name"].(string)),
		Capacity: utils.Int32(int32(v["capacity"].(int))),
	}
}

func flattenSignalRServiceSku(input *signalr.ResourceSku) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	capacity := 0
	if input.Capacity != nil {
		capacity = int(*input.Capacity)
	}

	name := ""
	if input.Name != nil {
		name = *input.Name
	}

	return []interface{}{
		map[string]interface{}{
			"capacity": capacity,
			"name":     name,
		},
	}
}
