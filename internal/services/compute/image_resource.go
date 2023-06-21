// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceid"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func image() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: imageCreateUpdate,
		Read:   imageRead,
		Update: imageCreateUpdate,
		Delete: imageDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ImageID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(90 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(90 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"source_virtual_machine_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourceid.ValidateResourceID,
			},

			"os_disk": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"os_type": {
							Type:             pluginsdk.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(compute.Linux),
								string(compute.Windows),
							}, true),
						},

						"os_state": {
							Type:             pluginsdk.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(compute.Generalized),
								string(compute.Specialized),
							}, true),
						},

						"managed_disk_id": {
							Type:             pluginsdk.TypeString,
							Computed:         true,
							Optional:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     resourceid.ValidateResourceID,
						},

						"blob_uri": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsURLWithHTTPorHTTPS,
						},

						"caching": {
							Type:             pluginsdk.TypeString,
							Optional:         true,
							Default:          string(compute.CachingTypesNone),
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(compute.CachingTypesNone),
								string(compute.CachingTypesReadOnly),
								string(compute.CachingTypesReadWrite),
							}, true),
						},

						"size_gb": {
							Type:         pluginsdk.TypeInt,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"data_disk": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"lun": {
							Type:     pluginsdk.TypeInt,
							Optional: true,
						},

						"managed_disk_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: resourceid.ValidateResourceID,
						},

						"blob_uri": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsURLWithHTTPorHTTPS,
						},

						"caching": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Default:  string(compute.CachingTypesNone),
							ValidateFunc: validation.StringInSlice([]string{
								string(compute.CachingTypesNone),
								string(compute.CachingTypesReadOnly),
								string(compute.CachingTypesReadWrite),
							}, true),
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"size_gb": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func imageCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.ImageClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureStack Image creation.")

	id := parse.NewImageID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	zoneResilient := false

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_image", id.ID())
		}
	}

	location := location.Normalize(d.Get("location").(string))
	expandedTags := tags.Expand(d.Get("tags").(map[string]interface{}))

	properties := compute.ImageProperties{}

	storageProfile := compute.ImageStorageProfile{
		OsDisk:        expandAzureStackImageOsDisk(d),
		DataDisks:     expandAzureStackImageDataDisks(d),
		ZoneResilient: &zoneResilient,
	}

	sourceVM := compute.SubResource{}
	if v, ok := d.GetOk("source_virtual_machine_id"); ok {
		vmID := v.(string)
		sourceVM = compute.SubResource{
			ID: &vmID,
		}
	}

	// either source VM or storage profile can be specified, but not both
	if sourceVM.ID == nil {
		// if both sourceVM and storageProfile are empty, return an error
		if storageProfile.OsDisk == nil && len(*storageProfile.DataDisks) == 0 {
			return fmt.Errorf("[ERROR] Cannot create image when both source VM and storage profile are empty")
		}

		properties.StorageProfile = &storageProfile
	} else {
		// creating an image from source VM
		properties.SourceVirtualMachine = &sourceVM
	}

	createImage := compute.Image{
		Name:            &id.Name,
		Location:        &location,
		Tags:            expandedTags,
		ImageProperties: &properties,
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, createImage)
	if err != nil {
		return err
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return err
	}

	d.SetId(id.ID())

	return imageRead(d, meta)
}

func imageRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.ImageClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ImageID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[ERROR] Error making Read request on AzureStack Image %q : %+v", id.String(), err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	// either source VM or storage profile can be specified, but not both
	if resp.SourceVirtualMachine != nil {
		d.Set("source_virtual_machine_id", resp.SourceVirtualMachine.ID)
	} else if resp.StorageProfile != nil {
		if disk := resp.StorageProfile.OsDisk; disk != nil {
			if err := d.Set("os_disk", flattenAzureStackImageOSDisk(disk)); err != nil {
				return fmt.Errorf("[DEBUG] Error setting AzureStack Image OS Disk error: %+v", err)
			}
		}

		if disks := resp.StorageProfile.DataDisks; disks != nil {
			if err := d.Set("data_disk", flattenAzureStackImageDataDisks(disks)); err != nil {
				return fmt.Errorf("[DEBUG] Error setting AzureStack Image Data Disks error: %+v", err)
			}
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func imageDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.ImageClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ImageID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return err
	}

	return future.WaitForCompletionRef(ctx, client.Client)
}

func flattenAzureStackImageOSDisk(osDisk *compute.ImageOSDisk) []interface{} {
	result := make(map[string]interface{})

	if disk := osDisk; disk != nil {
		if uri := osDisk.BlobURI; uri != nil {
			result["blob_uri"] = *uri
		}
		if diskSizeDB := osDisk.DiskSizeGB; diskSizeDB != nil {
			result["size_gb"] = *diskSizeDB
		}
		if disk := osDisk.ManagedDisk; disk != nil {
			result["managed_disk_id"] = *disk.ID
		}
		result["caching"] = string(osDisk.Caching)
		result["os_type"] = osDisk.OsType
		result["os_state"] = osDisk.OsState
	}

	return []interface{}{result}
}

func flattenAzureStackImageDataDisks(diskImages *[]compute.ImageDataDisk) []interface{} {
	result := make([]interface{}, 0)

	if images := diskImages; images != nil {
		for _, disk := range *images {
			l := make(map[string]interface{})
			if disk.BlobURI != nil {
				l["blob_uri"] = *disk.BlobURI
			}
			l["caching"] = string(disk.Caching)
			if disk.DiskSizeGB != nil {
				l["size_gb"] = *disk.DiskSizeGB
			}
			if v := disk.Lun; v != nil {
				l["lun"] = *v
			}
			if disk.ManagedDisk != nil && disk.ManagedDisk.ID != nil {
				l["managed_disk_id"] = *disk.ManagedDisk.ID
			}

			result = append(result, l)
		}
	}

	return result
}

func expandAzureStackImageOsDisk(d *pluginsdk.ResourceData) *compute.ImageOSDisk {
	osDisk := &compute.ImageOSDisk{}
	disks := d.Get("os_disk").([]interface{})

	if len(disks) > 0 {
		config := disks[0].(map[string]interface{})

		if v := config["os_type"].(string); v != "" {
			osType := compute.OperatingSystemTypes(v)
			osDisk.OsType = osType
		}

		if v := config["os_state"].(string); v != "" {
			osState := compute.OperatingSystemStateTypes(v)
			osDisk.OsState = osState
		}
		managedDiskID := config["managed_disk_id"].(string)
		if managedDiskID != "" {
			managedDisk := &compute.SubResource{
				ID: &managedDiskID,
			}
			osDisk.ManagedDisk = managedDisk
		}

		blobURI := config["blob_uri"].(string)
		osDisk.BlobURI = &blobURI

		if v := config["caching"].(string); v != "" {
			caching := compute.CachingTypes(v)
			osDisk.Caching = caching
		}

		if size := config["size_gb"]; size != 0 {
			diskSize := int32(size.(int))
			osDisk.DiskSizeGB = &diskSize
		}
	}

	return osDisk
}

func expandAzureStackImageDataDisks(d *pluginsdk.ResourceData) *[]compute.ImageDataDisk {
	disks := d.Get("data_disk").([]interface{})

	dataDisks := make([]compute.ImageDataDisk, 0, len(disks))
	for _, diskConfig := range disks {
		config := diskConfig.(map[string]interface{})

		managedDiskID := config["managed_disk_id"].(string)

		blobURI := config["blob_uri"].(string)
		lun := int32(config["lun"].(int))

		dataDisk := compute.ImageDataDisk{
			Lun:     &lun,
			BlobURI: &blobURI,
		}

		if size := config["size_gb"]; size != 0 {
			diskSize := int32(size.(int))
			dataDisk.DiskSizeGB = &diskSize
		}

		if v := config["caching"].(string); v != "" {
			caching := compute.CachingTypes(v)
			dataDisk.Caching = caching
		}

		if managedDiskID != "" {
			managedDisk := &compute.SubResource{
				ID: &managedDiskID,
			}
			dataDisk.ManagedDisk = managedDisk
		}

		dataDisks = append(dataDisks, dataDisk)
	}

	return &dataDisks
}
