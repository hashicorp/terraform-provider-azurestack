// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package compute

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/compute/mgmt/compute"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

func importVirtualMachineScaleSet(osType compute.OperatingSystemTypes, resourceType string) pluginsdk.ImporterFunc {
	return func(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) (data []*pluginsdk.ResourceData, err error) {
		id, err := parse.VirtualMachineScaleSetID(d.Id())
		if err != nil {
			return []*pluginsdk.ResourceData{}, err
		}

		client := meta.(*clients.Client).Compute.VMScaleSetClient
		// Upgrading to the 2021-07-01 exposed a new expand parameter in the GET method
		vm, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving Virtual Machine Scale Set %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}

		if vm.VirtualMachineScaleSetProperties == nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving Virtual Machine Scale Set %q (Resource Group %q): `properties` was nil", id.Name, id.ResourceGroup)
		}

		if vm.VirtualMachineScaleSetProperties.VirtualMachineProfile == nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving Virtual Machine Scale Set %q (Resource Group %q): `properties.virtualMachineProfile` was nil", id.Name, id.ResourceGroup)
		}

		if vm.VirtualMachineScaleSetProperties.VirtualMachineProfile.OsProfile == nil {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("retrieving Virtual Machine Scale Set %q (Resource Group %q): `properties.virtualMachineProfile.osProfile` was nil", id.Name, id.ResourceGroup)
		}

		isCorrectOS := false
		hasSshKeys := false
		if profile := vm.VirtualMachineScaleSetProperties.VirtualMachineProfile.OsProfile; profile != nil {
			if profile.LinuxConfiguration != nil && osType == compute.Linux {
				isCorrectOS = true

				if profile.LinuxConfiguration.SSH != nil && profile.LinuxConfiguration.SSH.PublicKeys != nil {
					hasSshKeys = len(*profile.LinuxConfiguration.SSH.PublicKeys) > 0
				}
			}

			if profile.WindowsConfiguration != nil && osType == compute.Windows {
				isCorrectOS = true
			}
		}

		if !isCorrectOS {
			return []*pluginsdk.ResourceData{}, fmt.Errorf("The %q resource only supports %s Virtual Machine Scale Sets", resourceType, string(osType))
		}

		if !hasSshKeys {
			d.Set("admin_password", "ignored-as-imported")
		}

		var updatedExtensions []map[string]interface{}
		if vm.VirtualMachineScaleSetProperties.VirtualMachineProfile.ExtensionProfile != nil {
			if extensionsProfile := vm.VirtualMachineScaleSetProperties.VirtualMachineProfile.ExtensionProfile; extensionsProfile != nil {
				for _, v := range *extensionsProfile.Extensions {
					v.ProtectedSettings = ""
				}
				updatedExtensions, err = flattenVirtualMachineScaleSetExtensions(extensionsProfile, d)
				if err != nil {
					return []*pluginsdk.ResourceData{}, fmt.Errorf("could not read VMSS extensions data for %q (resource group %q)", id.Name, id.ResourceGroup)
				}
			}
		}
		d.Set("extension", updatedExtensions)

		return []*pluginsdk.ResourceData{d}, nil
	}
}
