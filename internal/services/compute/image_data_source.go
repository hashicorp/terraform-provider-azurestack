package compute

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
)

func ImageDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: imageDataSourceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{

			"resource_group_name": commonschema.ResourceGroupName(),
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func imageDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.ImageClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	var img compute.Image

	img, err := client.Get(ctx, resourceGroup, name, "")
	if err != nil {
		return fmt.Errorf("image %q (Resource Group: %s) was not found", name, resourceGroup)
	}

	if img.Name == nil {
		return fmt.Errorf("image name is empty in Resource Group %s", resourceGroup)
	}

	d.SetId(*img.ID)
	d.Set("name", img.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("location", location.NormalizeNilable(img.Location))

	return nil
}
