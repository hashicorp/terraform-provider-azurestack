// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

// NOTE (also in the docs): this is not intended to be used with the `azurestack_virtual_machine_scale_set` resource

func virtualMachineScaleSetExtension() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: virtualMachineScaleSetExtensionCreate,
		Read:   virtualMachineScaleSetExtensionRead,
		Update: virtualMachineScaleSetExtensionUpdate,
		Delete: virtualMachineScaleSetExtensionDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VirtualMachineScaleSetExtensionID(id)
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
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"virtual_machine_scale_set_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.VirtualMachineScaleSetID,
			},

			"publisher": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"type": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"type_handler_version": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"auto_upgrade_minor_version": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"force_update_tag": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"protected_settings": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Sensitive:        true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
			},

			"settings": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
			},
		},
	}
}

func virtualMachineScaleSetExtensionCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetExtensionsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	virtualMachineScaleSetId, err := parse.VirtualMachineScaleSetID(d.Get("virtual_machine_scale_set_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewVirtualMachineScaleSetExtensionID(virtualMachineScaleSetId.SubscriptionId, virtualMachineScaleSetId.ResourceGroup, virtualMachineScaleSetId.Name, d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualMachineScaleSetName, id.ExtensionName, "")
	if err != nil {
		if !utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("checking for existing %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(resp.Response) {
		return tf.ImportAsExistsError("azurestack_virtual_machine_scale_set_extension", *resp.ID)
	}

	settings := map[string]interface{}{}
	if settingsString := d.Get("settings").(string); settingsString != "" {
		s, err := pluginsdk.ExpandJsonFromString(settingsString)
		if err != nil {
			return fmt.Errorf("unable to parse `settings`: %s", err)
		}
		settings = s
	}

	protectedSettings := map[string]interface{}{}
	if protectedSettingsString := d.Get("protected_settings").(string); protectedSettingsString != "" {
		ps, err := pluginsdk.ExpandJsonFromString(protectedSettingsString)
		if err != nil {
			return fmt.Errorf("unable to parse `protected_settings`: %s", err)
		}
		protectedSettings = ps
	}

	props := compute.VirtualMachineScaleSetExtension{
		Name: pointer.FromString(id.ExtensionName),
		VirtualMachineScaleSetExtensionProperties: &compute.VirtualMachineScaleSetExtensionProperties{
			Publisher:               pointer.FromString(d.Get("publisher").(string)),
			Type:                    pointer.FromString(d.Get("type").(string)),
			TypeHandlerVersion:      pointer.FromString(d.Get("type_handler_version").(string)),
			AutoUpgradeMinorVersion: pointer.FromBool(d.Get("auto_upgrade_minor_version").(bool)),
			ProtectedSettings:       protectedSettings,
			Settings:                settings,
		},
	}
	if v, ok := d.GetOk("force_update_tag"); ok {
		props.VirtualMachineScaleSetExtensionProperties.ForceUpdateTag = pointer.FromString(v.(string))
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualMachineScaleSetName, id.ExtensionName, props)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return virtualMachineScaleSetExtensionRead(d, meta)
}

func virtualMachineScaleSetExtensionUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetExtensionsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualMachineScaleSetExtensionID(d.Id())
	if err != nil {
		return err
	}

	props := compute.VirtualMachineScaleSetExtensionProperties{
		// if this isn't specified it defaults to false
		AutoUpgradeMinorVersion: pointer.FromBool(d.Get("auto_upgrade_minor_version").(bool)),
	}

	if d.HasChange("force_update_tag") {
		props.ForceUpdateTag = pointer.FromString(d.Get("force_update_tag").(string))
	}

	if d.HasChange("protected_settings") {
		protectedSettings := map[string]interface{}{}
		if protectedSettingsString := d.Get("protected_settings").(string); protectedSettingsString != "" {
			ps, err := pluginsdk.ExpandJsonFromString(protectedSettingsString)
			if err != nil {
				return fmt.Errorf("unable to parse `protected_settings`: %s", err)
			}
			protectedSettings = ps
		}

		props.ProtectedSettings = protectedSettings
	}

	if d.HasChange("publisher") {
		props.Publisher = pointer.FromString(d.Get("publisher").(string))
	}

	if d.HasChange("settings") {
		settings := map[string]interface{}{}

		if settingsString := d.Get("settings").(string); settingsString != "" {
			s, err := pluginsdk.ExpandJsonFromString(settingsString)
			if err != nil {
				return fmt.Errorf("unable to parse `settings`: %s", err)
			}
			settings = s
		}

		props.Settings = settings
	}

	if d.HasChange("type") {
		props.Type = pointer.FromString(d.Get("type").(string))
	}

	if d.HasChange("type_handler_version") {
		props.TypeHandlerVersion = pointer.FromString(d.Get("type_handler_version").(string))
	}

	extension := compute.VirtualMachineScaleSetExtension{
		Name: pointer.FromString(id.ExtensionName),
		VirtualMachineScaleSetExtensionProperties: &props,
	}
	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualMachineScaleSetName, id.ExtensionName, extension)
	if err != nil {
		return fmt.Errorf("updating Extension %q (Virtual Machine Scale Set %q / Resource Group %q): %+v", id.ExtensionName, id.VirtualMachineScaleSetName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of Extension %q (Virtual Machine Scale Set %q / Resource Group %q): %+v", id.ExtensionName, id.VirtualMachineScaleSetName, id.ResourceGroup, err)
	}

	return virtualMachineScaleSetExtensionRead(d, meta)
}

func virtualMachineScaleSetExtensionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	vmssClient := meta.(*clients.Client).Compute.VMScaleSetClient
	client := meta.(*clients.Client).Compute.VMScaleSetExtensionsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualMachineScaleSetExtensionID(d.Id())
	if err != nil {
		return err
	}

	vmss, err := vmssClient.Get(ctx, id.ResourceGroup, id.VirtualMachineScaleSetName)
	if err != nil {
		if utils.ResponseWasNotFound(vmss.Response) {
			log.Printf("Virtual Machine Scale Set %q was not found in Resource Group %q - removing Extension from state!", id.VirtualMachineScaleSetName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Virtual Machine Scale Set %q (Resource Group %q): %+v", id.VirtualMachineScaleSetName, id.ResourceGroup, err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.VirtualMachineScaleSetName, id.ExtensionName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("Extension %q (Virtual Machine Scale Set %q / Resource Group %q) was not found - removing from state!", id.ExtensionName, id.VirtualMachineScaleSetName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Extension %q (Virtual Machine Scale Set %q / Resource Group %q): %+v", id.ExtensionName, id.VirtualMachineScaleSetName, id.ResourceGroup, err)
	}

	d.Set("name", id.ExtensionName)
	d.Set("virtual_machine_scale_set_id", vmss.ID)

	if props := resp.VirtualMachineScaleSetExtensionProperties; props != nil {
		d.Set("auto_upgrade_minor_version", props.AutoUpgradeMinorVersion)
		d.Set("force_update_tag", props.ForceUpdateTag)
		d.Set("publisher", props.Publisher)
		d.Set("type", props.Type)
		d.Set("type_handler_version", props.TypeHandlerVersion)

		settings := ""
		if props.Settings != nil {
			settingsVal, ok := props.Settings.(map[string]interface{})
			if ok {
				settingsJson, err := pluginsdk.FlattenJsonToString(settingsVal)
				if err != nil {
					return fmt.Errorf("unable to parse settings from response: %s", err)
				}
				settings = settingsJson
			}
		}
		d.Set("settings", settings)
	}

	return nil
}

func virtualMachineScaleSetExtensionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetExtensionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualMachineScaleSetExtensionID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.VirtualMachineScaleSetName, id.ExtensionName)
	if err != nil {
		return fmt.Errorf("deleting Extension %q (Virtual Machine Scale Set %q / Resource Group %q): %+v", id.ExtensionName, id.VirtualMachineScaleSetName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Extension %q (Virtual Machine Scale Set %q / Resource Group %q): %+v", id.ExtensionName, id.VirtualMachineScaleSetName, id.ResourceGroup, err)
	}

	return nil
}
