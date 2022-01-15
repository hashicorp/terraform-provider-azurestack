package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/internal/features"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

func schemaFeatures() *pluginsdk.Schema {
	// NOTE: if there's only one nested field these want to be Required (since there's no point
	//       specifying the block otherwise) - however for 2+ they should be optional
	featuresMap := map[string]*pluginsdk.Schema{
		"resource_group": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*schema.Schema{
					"prevent_deletion_if_contains_resources": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
					},
				},
			},
		},
	}

	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true, // todo change to required in 1.0?
		MaxItems: 1,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: featuresMap,
		},
	}
}

func expandFeatures(input []interface{}) features.UserFeatures {
	// these are the defaults if omitted from the config
	featuresMap := features.Default()

	if len(input) == 0 || input[0] == nil {
		return featuresMap
	}

	val := input[0].(map[string]interface{})

	if raw, ok := val["resource_group"]; ok {
		items := raw.([]interface{})
		if len(items) > 0 {
			resourceGroupRaw := items[0].(map[string]interface{})
			if v, ok := resourceGroupRaw["prevent_deletion_if_contains_resources"]; ok {
				featuresMap.ResourceGroup.PreventDeletionIfContainsResources = v.(bool)
			}
		}
	}

	return featuresMap
}
