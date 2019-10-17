package tags

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Flatten(tagMap map[string]*string) map[string]interface{} {
	// If tagsMap is nil, len(tagsMap) will be 0.
	output := make(map[string]interface{}, len(tagMap))

	for i, v := range tagMap {
		if v == nil {
			continue
		}

		output[i] = *v
	}

	return output
}

// TODO: change function signature to match azurerm
// func FlattenAndSet(d *schema.ResourceData, tagMap map[string]*string) error {

func FlattenAndSet(d *schema.ResourceData, tagsMap *map[string]*string) {
	// If tagsMap is nil, len(tagsMap) will be 0.
	if tagsMap == nil {
		d.Set("tags", make(map[string]interface{}))
		return
	}

	output := make(map[string]interface{}, len(*tagsMap))

	for i, v := range *tagsMap {
		output[i] = *v
	}

	d.Set("tags", output)
}
