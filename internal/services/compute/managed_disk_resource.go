// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceid"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/locks"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func managedDisk() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceManagedDiskCreate,
		Read:   resourceManagedDiskRead,
		Update: resourceManagedDiskUpdate,
		Delete: resourceManagedDiskDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ManagedDiskID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"storage_account_type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.StandardLRS),
					string(compute.PremiumLRS),
				}, false),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"encryption": encryptionSettingsSchema(),

			"disk_size_gb": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.ManagedDiskSizeGB,
			},

			"create_option": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.Copy),
					string(compute.Empty),
					string(compute.FromImage),
					string(compute.Import),
				}, false),
			},

			"hyper_v_generation": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true, // Not supported by disk update
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.HyperVGenerationTypesV1),
					string(compute.HyperVGenerationTypeV2),
				}, false),
			},

			"source_uri": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"source_resource_id": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"storage_account_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true, // Not supported by disk update
				ValidateFunc: resourceid.ValidateResourceID,
			},

			"image_reference_id": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"os_type": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.Windows),
					string(compute.Linux),
				}, true),
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceManagedDiskCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Compute.DisksClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM Managed Disk creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewManagedDiskID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.DiskName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Managed Disk %q (Resource Group %q): %s", name, resourceGroup, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_managed_disk", id.ID())
		}
	}

	location := location.Normalize(d.Get("location").(string))
	createOption := compute.DiskCreateOption(d.Get("create_option").(string))
	storageAccountType := d.Get("storage_account_type").(string)
	osType := d.Get("os_type").(string)

	t := d.Get("tags").(map[string]interface{})
	skuName := compute.DiskStorageAccountTypes(storageAccountType)

	props := &compute.DiskProperties{
		CreationData: &compute.CreationData{
			CreateOption: createOption,
		},
		OsType: compute.OperatingSystemTypes(osType),
	}

	diskSizeGB := d.Get("disk_size_gb").(int)
	if diskSizeGB != 0 {
		props.DiskSizeGB = utils.Int32(int32(diskSizeGB))
	}

	if createOption == compute.Import {
		sourceUri := d.Get("source_uri").(string)
		if sourceUri == "" {
			return fmt.Errorf("`source_uri` must be specified when `create_option` is set to `Import`")
		}

		storageAccountId := d.Get("storage_account_id").(string)
		if storageAccountId == "" {
			return fmt.Errorf("`storage_account_id` must be specified when `create_option` is set to `Import`")
		}

		props.CreationData.StorageAccountID = pointer.FromString(storageAccountId)
		props.CreationData.SourceURI = pointer.FromString(sourceUri)
	}
	if createOption == compute.Copy {
		sourceResourceId := d.Get("source_resource_id").(string)
		if sourceResourceId == "" {
			return fmt.Errorf("`source_resource_id` must be specified when `create_option` is set to `Copy` or `Restore`")
		}

		props.CreationData.SourceResourceID = pointer.FromString(sourceResourceId)
	}
	if createOption == compute.FromImage {
		if imageReferenceId := d.Get("image_reference_id").(string); imageReferenceId != "" {
			props.CreationData.ImageReference = &compute.ImageDiskReference{
				ID: pointer.FromString(imageReferenceId),
			}
		} else {
			return fmt.Errorf("`image_reference_id` must be specified when `create_option` is set to `FromImage`")
		}
	}

	if v, ok := d.GetOk("encryption"); ok {
		encryptionSettings := v.([]interface{})
		settings := encryptionSettings[0].(map[string]interface{})
		props.EncryptionSettingsCollection = expandManagedDiskEncryptionSettings(settings)
	}

	if v, ok := d.GetOk("hyper_v_generation"); ok {
		props.HyperVGeneration = compute.HyperVGeneration(v.(string))
	}

	createDisk := compute.Disk{
		Name:           &name,
		Location:       &location,
		DiskProperties: props,
		Sku: &compute.DiskSku{
			Name: skuName,
		},
		Tags: tags.Expand(t),
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, name, createDisk)
	if err != nil {
		return fmt.Errorf("creating/updating Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for create/update of Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("retrieving Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if read.ID == nil {
		return fmt.Errorf("reading Managed Disk %s (Resource Group %q): ID was nil", name, resourceGroup)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return resourceManagedDiskRead(d, meta)
}

func resourceManagedDiskUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.DisksClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM Managed Disk update.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	storageAccountType := d.Get("storage_account_type").(string)
	shouldShutDown := false

	disk, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(disk.Response) {
			return fmt.Errorf("Managed Disk %q (Resource Group %q) was not found", name, resourceGroup)
		}

		return fmt.Errorf("making Read request on Azure Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	diskUpdate := compute.DiskUpdate{
		DiskUpdateProperties: &compute.DiskUpdateProperties{},
	}

	if d.HasChange("tags") {
		t := d.Get("tags").(map[string]interface{})
		diskUpdate.Tags = tags.Expand(t)
	}

	if d.HasChange("storage_account_type") {
		shouldShutDown = true
		var skuName compute.DiskStorageAccountTypes
		for _, v := range compute.PossibleDiskStorageAccountTypesValues() {
			if strings.EqualFold(storageAccountType, string(v)) {
				skuName = v
			}
		}
		diskUpdate.Sku = &compute.DiskSku{
			Name: skuName,
		}
	}

	if d.HasChange("os_type") {
		diskUpdate.DiskUpdateProperties.OsType = compute.OperatingSystemTypes(d.Get("os_type").(string))
	}

	if d.HasChange("disk_size_gb") {
		if old, new := d.GetChange("disk_size_gb"); new.(int) > old.(int) {
			shouldShutDown = true
			diskUpdate.DiskUpdateProperties.DiskSizeGB = utils.Int32(int32(new.(int)))
		} else {
			return fmt.Errorf("- New size must be greater than original size. Shrinking disks is not supported on Azure")
		}
	}

	// whilst we need to shut this down, if we're not attached to anything there's no point
	if shouldShutDown && disk.ManagedBy == nil {
		shouldShutDown = false
	}

	// if we are attached to a VM we bring down the VM as necessary for the operations which are not allowed while it's online
	if shouldShutDown {
		virtualMachine, err := parse.VirtualMachineID(*disk.ManagedBy)
		if err != nil {
			return fmt.Errorf("parsing VMID %q for disk attachment: %+v", *disk.ManagedBy, err)
		}
		// check instanceView State
		vmClient := meta.(*clients.Client).Compute.VMClient

		locks.ByName(name, virtualMachineResourceName)
		defer locks.UnlockByName(name, virtualMachineResourceName)

		instanceView, err := vmClient.InstanceView(ctx, virtualMachine.ResourceGroup, virtualMachine.Name)
		if err != nil {
			return fmt.Errorf("retrieving InstanceView for Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
		}

		shouldTurnBackOn := true
		shouldDeallocate := true

		if instanceView.Statuses != nil {
			for _, status := range *instanceView.Statuses {
				if status.Code == nil {
					continue
				}

				// could also be the provisioning state which we're not bothered with here
				state := strings.ToLower(*status.Code)
				if !strings.HasPrefix(state, "powerstate/") {
					continue
				}

				state = strings.TrimPrefix(state, "powerstate/")
				switch strings.ToLower(state) {
				case "deallocated":
				case "deallocating":
					shouldTurnBackOn = false
					shouldShutDown = false
					shouldDeallocate = false
				case "stopping":
				case "stopped":
					shouldShutDown = false
					shouldTurnBackOn = false
				}
			}
		}

		// Shutdown
		if shouldShutDown {
			log.Printf("[DEBUG] Shutting Down Virtual Machine %q (Resource Group %q)..", virtualMachine.Name, virtualMachine.ResourceGroup)
			forceShutdown := false
			future, err := vmClient.PowerOff(ctx, virtualMachine.ResourceGroup, virtualMachine.Name, utils.Bool(forceShutdown))
			if err != nil {
				return fmt.Errorf("sending Power Off to Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
			}

			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for Power Off of Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
			}

			log.Printf("[DEBUG] Shut Down Virtual Machine %q (Resource Group %q)..", virtualMachine.Name, virtualMachine.ResourceGroup)
		}

		// De-allocate
		if shouldDeallocate {
			log.Printf("[DEBUG] Deallocating Virtual Machine %q (Resource Group %q)..", virtualMachine.Name, virtualMachine.ResourceGroup)
			// Upgrading to 2021-07-01 exposed a new hibernate paramater to the Deallocate method
			deAllocFuture, err := vmClient.Deallocate(ctx, virtualMachine.ResourceGroup, virtualMachine.Name)
			if err != nil {
				return fmt.Errorf("Deallocating to Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
			}

			if err := deAllocFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for Deallocation of Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
			}

			log.Printf("[DEBUG] Deallocated Virtual Machine %q (Resource Group %q)..", virtualMachine.Name, virtualMachine.ResourceGroup)
		}

		// Update Disk
		updateFuture, err := client.Update(ctx, resourceGroup, name, diskUpdate)
		if err != nil {
			return fmt.Errorf("updating Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
		if err := updateFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for update of Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		if shouldTurnBackOn {
			log.Printf("[DEBUG] Starting Linux Virtual Machine %q (Resource Group %q)..", virtualMachine.Name, virtualMachine.ResourceGroup)
			future, err := vmClient.Start(ctx, virtualMachine.ResourceGroup, virtualMachine.Name)
			if err != nil {
				return fmt.Errorf("starting Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
			}

			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				return fmt.Errorf("waiting for start of Virtual Machine %q (Resource Group %q): %+v", virtualMachine.Name, virtualMachine.ResourceGroup, err)
			}

			log.Printf("[DEBUG] Started Virtual Machine %q (Resource Group %q)..", virtualMachine.Name, virtualMachine.ResourceGroup)
		}
	} else { // otherwise, just update it
		diskFuture, err := client.Update(ctx, resourceGroup, name, diskUpdate)
		if err != nil {
			return fmt.Errorf("expanding managed disk %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		err = diskFuture.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("waiting for expand operation on managed disk %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	return resourceManagedDiskRead(d, meta)
}

func resourceManagedDiskRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.DisksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ManagedDiskID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.DiskName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Disk %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on Azure Managed Disk %s (resource group %s): %s", id.DiskName, id.ResourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if sku := resp.Sku; sku != nil {
		d.Set("storage_account_type", string(sku.Name))
	}

	if props := resp.DiskProperties; props != nil {
		if creationData := props.CreationData; creationData != nil {
			d.Set("create_option", string(creationData.CreateOption))

			// imageReference is returned as well when galleryImageRefernece is used, only check imageReference when galleryImageReference is not returned
			if imageReference := creationData.ImageReference; imageReference != nil && imageReference.ID != nil {
				d.Set("image_reference_id", imageReference.ID)
			}

			d.Set("source_resource_id", creationData.SourceResourceID)
			d.Set("source_uri", creationData.SourceURI)
			d.Set("storage_account_id", creationData.StorageAccountID)
		}

		d.Set("disk_size_gb", props.DiskSizeGB)
		d.Set("os_type", props.OsType)
		d.Set("hyper_v_generation", props.HyperVGeneration)

		if err := d.Set("encryption", flattenManagedDiskEncryptionSettings(props.EncryptionSettingsCollection)); err != nil {
			return fmt.Errorf("setting `encryption`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceManagedDiskDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.DisksClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ManagedDiskID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.DiskName)
	if err != nil {
		return fmt.Errorf("deleting Managed Disk %q (Resource Group %q): %+v", id.DiskName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Managed Disk %q (Resource Group %q): %+v", id.DiskName, id.ResourceGroup, err)
	}

	return nil
}
