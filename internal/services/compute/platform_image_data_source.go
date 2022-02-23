package compute

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func platformImageDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: platformImageDataSourceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"location": commonschema.Location(),

			"publisher": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"offer": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"sku": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"version": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func platformImageDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMImageClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	l := location.Normalize(d.Get("location").(string))
	publisher := d.Get("publisher").(string)
	offer := d.Get("offer").(string)
	sku := d.Get("sku").(string)

	result, err := client.List(ctx, l, publisher, offer, sku, "", utils.Int32(int32(1000)), "name")
	if err != nil {
		return fmt.Errorf("reading Platform Images: %+v", err)
	}

	var image *compute.VirtualMachineImageResource
	if v, ok := d.GetOk("version"); ok {
		version := v.(string)
		items := *result.Value
		for i, item := range items {
			if item.Name != nil && *item.Name == version {
				image = &items[i]
				break
			}
		}
		if image == nil {
			return fmt.Errorf("could not find image (location %q / publisher %q / offer %q / sku %q / version % q): %+v", l, publisher, offer, sku, version, err)
		}
	} else {
		// get the latest image
		// the last value is the latest, apparently.
		image = &(*result.Value)[len(*result.Value)-1]
	}

	d.SetId(*image.ID)
	d.Set("location", location.NormalizeNilable(image.Location))

	d.Set("publisher", publisher)
	d.Set("offer", offer)
	d.Set("sku", sku)
	d.Set("version", image.Name)

	return nil
}
