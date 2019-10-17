package azurestack

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
)

func dataSourceArmResourceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmResourceGroupRead,

		Schema: map[string]*schema.Schema{
			"name":     resourceGroupNameForDataSourceSchema(),
			"location": locationForDataSourceSchema(),
			"tags":     tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmResourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).resourceGroupsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resp, err := client.Get(ctx, name)
	if err != nil {
		return err
	}

	d.SetId(*resp.ID)

	return resourceArmResourceGroupRead(d, meta)
}
