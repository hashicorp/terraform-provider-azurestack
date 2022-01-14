package azurestack

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/azurestack/helpers/pointer"
)

func dataSourceArmPlatformImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmPlatformImageRead,
		Schema: map[string]*schema.Schema{
			"location": locationSchema(),

			"publisher": {
				Type:     schema.TypeString,
				Required: true,
			},

			"offer": {
				Type:     schema.TypeString,
				Required: true,
			},

			"sku": {
				Type:     schema.TypeString,
				Required: true,
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceArmPlatformImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).vmImageClient
	ctx := meta.(*ArmClient).StopContext

	location := azureStackNormalizeLocation(d.Get("location").(string))
	publisher := d.Get("publisher").(string)
	offer := d.Get("offer").(string)
	sku := d.Get("sku").(string)

	result, err := client.List(ctx, location, publisher, offer, sku, "", pointer.FromInt32(1000), "name")
	if err != nil {
		return fmt.Errorf("reading Platform Images: %+v", err)
	}

	// the last value is the latest, apparently.
	latestVersion := (*result.Value)[len(*result.Value)-1]

	d.SetId(*latestVersion.ID)
	if location := latestVersion.Location; location != nil {
		d.Set("location", azureStackNormalizeLocation(*location))
	}

	d.Set("publisher", publisher)
	d.Set("offer", offer)
	d.Set("sku", sku)
	d.Set("version", latestVersion.Name)

	return nil
}
