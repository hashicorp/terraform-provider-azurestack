// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func imageDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: imageDataSourceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name_regex": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				ExactlyOneOf: []string{"name", "name_regex"},
			},

			"sort_descending": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "name_regex"},
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"location": commonschema.LocationComputed(),

			"os_disk": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"blob_uri": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"caching": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"managed_disk_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"os_state": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"os_type": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"size_gb": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"data_disk": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"blob_uri": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"caching": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"lun": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},
						"managed_disk_id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"size_gb": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func imageDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.ImageClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)
	nameRegex, nameRegexOk := d.GetOk("name_regex")

	var img compute.Image

	if !nameRegexOk {
		var err error
		if img, err = client.Get(ctx, resourceGroup, name, ""); err != nil {
			if utils.ResponseWasNotFound(img.Response) {
				return fmt.Errorf("image %q (Resource Group: %s) was not found", name, resourceGroup)
			}
			return fmt.Errorf("making Read request on image %q (Resource Group: %s): %s", name, resourceGroup, err)
		}
	} else {
		r := regexp.MustCompile(nameRegex.(string))
		list := make([]compute.Image, 0)

		resp, err := client.ListByResourceGroupComplete(ctx, resourceGroup)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response().Response) {
				return fmt.Errorf("no Images were found for Resource Group %q", resourceGroup)
			}
			return fmt.Errorf("getting list of images (resource group %q): %+v", resourceGroup, err)
		}

		for resp.NotDone() {
			img = resp.Value()
			if r.Match(([]byte)(*img.Name)) {
				list = append(list, img)
			}
			err = resp.NextWithContext(ctx)

			if err != nil {
				return err
			}
		}

		if 1 > len(list) {
			return fmt.Errorf("no Images were found for Resource Group %q", resourceGroup)
		}

		if len(list) > 1 {
			desc := d.Get("sort_descending").(bool)
			log.Printf("[DEBUG] Image - multiple results found and `sort_descending` is set to: %t", desc)

			sort.Slice(list, func(i, j int) bool {
				return (!desc && *list[i].Name < *list[j].Name) ||
					(desc && *list[i].Name > *list[j].Name)
			})
		}
		img = list[0]
	}

	if img.Name == nil {
		return fmt.Errorf("image name is empty in Resource Group %s", resourceGroup)
	}

	id := parse.NewImageID(subscriptionId, resourceGroup, *img.Name)

	d.SetId(id.ID())
	d.Set("name", img.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("location", location.NormalizeNilable(img.Location))

	if profile := img.StorageProfile; profile != nil {
		if disk := profile.OsDisk; disk != nil {
			if err := d.Set("os_disk", flattenAzureStackImageOSDisk(disk)); err != nil {
				return fmt.Errorf("[DEBUG] Error setting AzureStack Image OS Disk error: %+v", err)
			}
		}

		if disks := profile.DataDisks; disks != nil {
			if err := d.Set("data_disk", flattenAzureStackImageDataDisks(disks)); err != nil {
				return fmt.Errorf("[DEBUG] Error setting AzureStack Image Data Disks error: %+v", err)
			}
		}
	}

	return tags.FlattenAndSet(d, img.Tags)
}
