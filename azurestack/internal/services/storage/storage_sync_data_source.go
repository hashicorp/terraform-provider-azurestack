package storage

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/location"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/storage/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourceStorageSync() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceStorageSyncRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.StorageSyncName,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"location": azure.SchemaLocationForDataSource(),

			"incoming_traffic_policy": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceStorageSyncRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.SyncServiceClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Storage Sync %q was not found in Resource Group %q", name, resourceGroup)
		}
		return fmt.Errorf("retrieving Storage Sync %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if id := resp.ID; id != nil {
		d.SetId(*resp.ID)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if props := resp.ServiceProperties; props != nil {
		d.Set("incoming_traffic_policy", props.IncomingTrafficPolicy)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}
