package tags

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// SchemaDataSource returns the Schema which should be used for Tags on a Data Source
func SchemaDataSource() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
	}
}

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeMap,
		Optional:     true,
		Computed:     true,
		ValidateFunc: Validate,
	}
}
