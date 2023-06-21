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
	computeValidate "github.com/hashicorp/terraform-provider-azurestack/internal/services/compute/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/base64"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func resourceWindowsVirtualMachineScaleSet() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceWindowsVirtualMachineScaleSetCreate,
		Read:   resourceWindowsVirtualMachineScaleSetRead,
		Update: resourceWindowsVirtualMachineScaleSetUpdate,
		Delete: resourceWindowsVirtualMachineScaleSetDelete,

		Importer: pluginsdk.ImporterValidatingResourceIdThen(func(id string) error {
			_, err := parse.VirtualMachineScaleSetID(id)
			return err
		}, importVirtualMachineScaleSet(compute.Windows, "azurestack_windows_virtual_machine_scale_set")),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		// TODO: exposing requireGuestProvisionSignal once it's available
		// https://github.com/Azure/azure-rest-api-specs/pull/7246

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: computeValidate.VirtualMachineName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"location": commonschema.Location(),

			// Required
			"admin_username": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"admin_password": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				Sensitive:        true,
				DiffSuppressFunc: adminPasswordDiffSuppressFunc,
				ValidateFunc:     validation.StringIsNotEmpty,
			},

			"network_interface": VirtualMachineScaleSetNetworkInterfaceSchema(),

			"os_disk": VirtualMachineScaleSetOSDiskSchema(),

			"instances": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"sku": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			// Optional
			"additional_capabilities": VirtualMachineScaleSetAdditionalCapabilitiesSchema(),

			"additional_unattend_content": additionalUnattendContentSchema(),

			"automatic_os_upgrade_policy": VirtualMachineScaleSetAutomatedOSUpgradePolicySchema(),

			"automatic_instance_repair": VirtualMachineScaleSetAutomaticRepairsPolicySchema(),

			"boot_diagnostics": bootDiagnosticsSchema(),

			"computer_name_prefix": {
				Type:     pluginsdk.TypeString,
				Optional: true,

				// Computed since we reuse the VM name if one's not specified
				Computed: true,
				ForceNew: true,

				ValidateFunc: computeValidate.WindowsComputerNamePrefix,
			},

			"custom_data": base64.OptionalSchema(false),

			"data_disk": VirtualMachineScaleSetDataDiskSchema(),

			"do_not_run_extensions_on_overprovisioned_machines": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			// TODO 4.0: change this from enable_* to *_enabled
			"enable_automatic_updates": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"encryption_at_host_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
			},

			"extension": VirtualMachineScaleSetExtensionsSchema(),

			// TODO: Uncomment extensions_time_budget when it's fixed
			/*"extensions_time_budget": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Default:      "PT1H30M",
				ValidateFunc: utils.ISO8601DurationBetween("PT15M", "PT2H"),
			},*/

			"health_probe_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourceid.ValidateResourceID,
			},

			"license_type": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"None",
					"Windows_Client",
					"Windows_Server",
				}, false),
				DiffSuppressFunc: func(_, old, new string, _ *pluginsdk.ResourceData) bool {
					if old == "None" && new == "" || old == "" && new == "None" {
						return true
					}

					return false
				},
			},

			"overprovision": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"plan": planSchema(),

			"platform_fault_domain_count": {
				Type:     pluginsdk.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"provision_vm_agent": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},

			"secret": windowsSecretSchema(),

			"single_placement_group": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"source_image_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourceid.ValidateResourceID,
			},

			"source_image_reference": sourceImageReferenceSchema(false),

			"tags": tags.Schema(),

			"timezone": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: computeValidate.VirtualMachineTimeZone(),
			},

			"upgrade_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  string(compute.UpgradeModeManual),
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.UpgradeModeAutomatic),
					string(compute.UpgradeModeManual),
					string(compute.UpgradeModeRolling),
				}, false),
			},

			"winrm_listener": winRmListenerSchema(),

			"scale_in_policy": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(compute.Default),
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.Default),
					string(compute.NewestVM),
					string(compute.OldestVM),
				}, false),
			},

			"terminate_notification": VirtualMachineScaleSetTerminateNotificationSchema(),

			// Computed
			"unique_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWindowsVirtualMachineScaleSetCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVirtualMachineScaleSetID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	exists, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(exists.Response) {
			return fmt.Errorf("checking for existing Windows %s: %+v", id, err)
		}
	}

	if !utils.ResponseWasNotFound(exists.Response) {
		return tf.ImportAsExistsError("azurestack_windows_virtual_machine_scale_set", *exists.ID)
	}

	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	additionalCapabilitiesRaw := d.Get("additional_capabilities").([]interface{})
	additionalCapabilities := ExpandVirtualMachineScaleSetAdditionalCapabilities(additionalCapabilitiesRaw)

	additionalUnattendContentRaw := d.Get("additional_unattend_content").([]interface{})
	additionalUnattendContent := expandAdditionalUnattendContent(additionalUnattendContentRaw)

	bootDiagnosticsRaw := d.Get("boot_diagnostics").([]interface{})
	bootDiagnostics := expandBootDiagnostics(bootDiagnosticsRaw)

	dataDisksRaw := d.Get("data_disk").([]interface{})
	ultraSSDEnabled := d.Get("additional_capabilities.0.ultra_ssd_enabled").(bool)
	dataDisks, err := ExpandVirtualMachineScaleSetDataDisk(dataDisksRaw, ultraSSDEnabled)
	if err != nil {
		return fmt.Errorf("expanding `data_disk`: %+v", err)
	}

	networkInterfacesRaw := d.Get("network_interface").([]interface{})
	networkInterfaces, err := ExpandVirtualMachineScaleSetNetworkInterface(networkInterfacesRaw)
	if err != nil {
		return fmt.Errorf("expanding `network_interface`: %+v", err)
	}

	osDiskRaw := d.Get("os_disk").([]interface{})
	osDisk := ExpandVirtualMachineScaleSetOSDisk(osDiskRaw, compute.Windows)

	planRaw := d.Get("plan").([]interface{})
	plan := expandPlan(planRaw)

	sourceImageReferenceRaw := d.Get("source_image_reference").([]interface{})
	sourceImageId := d.Get("source_image_id").(string)
	sourceImageReference, err := expandSourceImageReference(sourceImageReferenceRaw, sourceImageId)
	if err != nil {
		return err
	}

	healthProbeId := d.Get("health_probe_id").(string)
	upgradeMode := compute.UpgradeMode(d.Get("upgrade_mode").(string))
	automaticOSUpgradePolicyRaw := d.Get("automatic_os_upgrade_policy").([]interface{})
	automaticOSUpgradePolicy := ExpandVirtualMachineScaleSetAutomaticUpgradePolicy(automaticOSUpgradePolicyRaw)

	if upgradeMode != compute.UpgradeModeAutomatic && len(automaticOSUpgradePolicyRaw) > 0 {
		return fmt.Errorf("an `automatic_os_upgrade_policy` block cannot be specified when `upgrade_mode` is not set to `Automatic`")
	}

	winRmListenersRaw := d.Get("winrm_listener").(*pluginsdk.Set).List()
	winRmListeners := expandWinRMListener(winRmListenersRaw)

	secretsRaw := d.Get("secret").([]interface{})
	secrets := expandWindowsSecrets(secretsRaw)

	var computerNamePrefix string
	if v, ok := d.GetOk("computer_name_prefix"); ok && len(v.(string)) > 0 {
		computerNamePrefix = v.(string)
	} else {
		_, errs := computeValidate.WindowsComputerNamePrefix(d.Get("name"), "computer_name_prefix")
		if len(errs) > 0 {
			return fmt.Errorf("unable to assume default computer name prefix %s. Please adjust the %q, or specify an explicit %q", errs[0], "name", "computer_name_prefix")
		}
		computerNamePrefix = id.Name
	}

	networkProfile := &compute.VirtualMachineScaleSetNetworkProfile{
		NetworkInterfaceConfigurations: networkInterfaces,
	}
	if healthProbeId != "" {
		networkProfile.HealthProbe = &compute.APIEntityReference{
			ID: utils.String(healthProbeId),
		}
	}

	upgradePolicy := compute.UpgradePolicy{
		Mode:                     upgradeMode,
		AutomaticOSUpgradePolicy: automaticOSUpgradePolicy,
	}

	virtualMachineProfile := compute.VirtualMachineScaleSetVMProfile{
		OsProfile: &compute.VirtualMachineScaleSetOSProfile{
			AdminPassword:      utils.String(d.Get("admin_password").(string)),
			AdminUsername:      utils.String(d.Get("admin_username").(string)),
			ComputerNamePrefix: utils.String(computerNamePrefix),
			WindowsConfiguration: &compute.WindowsConfiguration{
				ProvisionVMAgent: utils.Bool(d.Get("provision_vm_agent").(bool)),
				WinRM:            winRmListeners,
			},
			Secrets: secrets,
		},
		DiagnosticsProfile: bootDiagnostics,
		NetworkProfile:     networkProfile,
		StorageProfile: &compute.VirtualMachineScaleSetStorageProfile{
			ImageReference: sourceImageReference,
			OsDisk:         osDisk,
			DataDisks:      dataDisks,
		},
	}

	hasHealthExtension := false
	if vmExtensionsRaw, ok := d.GetOk("extension"); ok {
		virtualMachineProfile.ExtensionProfile, hasHealthExtension, err = expandVirtualMachineScaleSetExtensions(vmExtensionsRaw.(*pluginsdk.Set).List())
		if err != nil {
			return err
		}
	}

	// otherwise the service return the error:
	// Rolling Upgrade mode is not supported for this Virtual Machine Scale Set because a health probe or health extension was not provided.
	if upgradeMode == compute.UpgradeModeRolling && (healthProbeId == "" && !hasHealthExtension) {
		return fmt.Errorf("`health_probe_id` must be set or a health extension must be specified when `upgrade_mode` is set to %q", string(upgradeMode))
	}

	enableAutomaticUpdates := d.Get("enable_automatic_updates").(bool)
	virtualMachineProfile.OsProfile.WindowsConfiguration.EnableAutomaticUpdates = utils.Bool(enableAutomaticUpdates)

	if v, ok := d.GetOk("custom_data"); ok {
		virtualMachineProfile.OsProfile.CustomData = utils.String(v.(string))
	}

	if encryptionAtHostEnabled, ok := d.GetOk("encryption_at_host_enabled"); ok {
		if virtualMachineProfile.SecurityProfile == nil {
			virtualMachineProfile.SecurityProfile = &compute.SecurityProfile{}
		}
		virtualMachineProfile.SecurityProfile.EncryptionAtHost = utils.Bool(encryptionAtHostEnabled.(bool))
	}

	if len(additionalUnattendContentRaw) > 0 {
		virtualMachineProfile.OsProfile.WindowsConfiguration.AdditionalUnattendContent = additionalUnattendContent
	}

	if v, ok := d.GetOk("license_type"); ok {
		virtualMachineProfile.LicenseType = utils.String(v.(string))
	}

	if v, ok := d.GetOk("timezone"); ok {
		virtualMachineProfile.OsProfile.WindowsConfiguration.TimeZone = utils.String(v.(string))
	}

	if v, ok := d.GetOk("terminate_notification"); ok {
		virtualMachineProfile.ScheduledEventsProfile = ExpandVirtualMachineScaleSetScheduledEventsProfile(v.([]interface{}))
	}

	scaleInPolicy := d.Get("scale_in_policy").(string)
	automaticRepairsPolicyRaw := d.Get("automatic_instance_repair").([]interface{})
	automaticRepairsPolicy := ExpandVirtualMachineScaleSetAutomaticRepairsPolicy(automaticRepairsPolicyRaw)

	props := compute.VirtualMachineScaleSet{
		Location: utils.String(location),
		Sku: &compute.Sku{
			Name:     utils.String(d.Get("sku").(string)),
			Capacity: utils.Int64(int64(d.Get("instances").(int))),

			// doesn't appear this can be set to anything else, even Promo machines are Standard
			Tier: utils.String("Standard"),
		},
		Plan: plan,
		Tags: tags.Expand(t),
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			AdditionalCapabilities:                 additionalCapabilities,
			AutomaticRepairsPolicy:                 automaticRepairsPolicy,
			DoNotRunExtensionsOnOverprovisionedVMs: utils.Bool(d.Get("do_not_run_extensions_on_overprovisioned_machines").(bool)),
			Overprovision:                          utils.Bool(d.Get("overprovision").(bool)),
			SinglePlacementGroup:                   utils.Bool(d.Get("single_placement_group").(bool)),
			VirtualMachineProfile:                  &virtualMachineProfile,
			UpgradePolicy:                          &upgradePolicy,
			ScaleInPolicy: &compute.ScaleInPolicy{
				Rules: &[]compute.VirtualMachineScaleSetScaleInRules{compute.VirtualMachineScaleSetScaleInRules(scaleInPolicy)},
			},
		},
	}

	if v, ok := d.GetOk("platform_fault_domain_count"); ok {
		props.VirtualMachineScaleSetProperties.PlatformFaultDomainCount = utils.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Creating Windows %s.", id)
	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, props)
	if err != nil {
		return fmt.Errorf("creating Windows %s: %+v", id, err)
	}

	log.Printf("[DEBUG] Waiting for Windows %s to be created.", id)
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Windows %s: %+v", id, err)
	}
	log.Printf("[DEBUG] Windows %s was created", id)

	d.SetId(id.ID())

	return resourceWindowsVirtualMachineScaleSetRead(d, meta)
}

func resourceWindowsVirtualMachineScaleSetUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualMachineScaleSetID(d.Id())
	if err != nil {
		return err
	}

	updateInstances := false

	// retrieve
	// Upgrading to the 2021-07-01 exposed a new expand parameter in the GET method
	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	if existing.VirtualMachineScaleSetProperties == nil {
		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): `properties` was nil", id.Name, id.ResourceGroup)
	}
	if existing.VirtualMachineScaleSetProperties.VirtualMachineProfile == nil {
		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): `properties.virtualMachineProfile` was nil", id.Name, id.ResourceGroup)
	}
	if existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile == nil {
		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): `properties.virtualMachineProfile,storageProfile` was nil", id.Name, id.ResourceGroup)
	}

	updateProps := compute.VirtualMachineScaleSetUpdateProperties{
		VirtualMachineProfile: &compute.VirtualMachineScaleSetUpdateVMProfile{
			// if an image reference has been configured previously (it has to be), we would better to include that in this
			// update request to avoid some circumstances that the API will complain ImageReference is null
			// issue tracking: https://github.com/Azure/azure-rest-api-specs/issues/10322
			StorageProfile: &compute.VirtualMachineScaleSetUpdateStorageProfile{
				ImageReference: existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.ImageReference,
			},
		},
		// if an upgrade policy's been configured previously (which it will have) it must be threaded through
		// this doesn't matter for Manual - but breaks when updating anything on a Automatic and Rolling Mode Scale Set
		UpgradePolicy: existing.VirtualMachineScaleSetProperties.UpgradePolicy,
	}
	update := compute.VirtualMachineScaleSetUpdate{}

	upgradeMode := compute.UpgradeMode(d.Get("upgrade_mode").(string))
	// first try and pull this from existing vm, which covers no changes being made to this block
	automaticOSUpgradeIsEnabled := false
	if policy := existing.VirtualMachineScaleSetProperties.UpgradePolicy; policy != nil {
		if policy.AutomaticOSUpgradePolicy != nil && policy.AutomaticOSUpgradePolicy.EnableAutomaticOSUpgrade != nil {
			automaticOSUpgradeIsEnabled = *policy.AutomaticOSUpgradePolicy.EnableAutomaticOSUpgrade
		}
	}
	if d.HasChange("automatic_os_upgrade_policy") {
		upgradePolicy := compute.UpgradePolicy{}
		if existing.VirtualMachineScaleSetProperties.UpgradePolicy == nil {
			upgradePolicy = compute.UpgradePolicy{
				Mode: compute.UpgradeMode(d.Get("upgrade_mode").(string)),
			}
		} else {
			upgradePolicy = *existing.VirtualMachineScaleSetProperties.UpgradePolicy
			upgradePolicy.Mode = compute.UpgradeMode(d.Get("upgrade_mode").(string))
		}

		if d.HasChange("automatic_os_upgrade_policy") {
			automaticRaw := d.Get("automatic_os_upgrade_policy").([]interface{})
			upgradePolicy.AutomaticOSUpgradePolicy = ExpandVirtualMachineScaleSetAutomaticUpgradePolicy(automaticRaw)

			// however if this block has been changed then we need to pull it
			// we can guarantee this always has a value since it'll have been expanded and thus is safe to de-ref
			automaticOSUpgradeIsEnabled = *upgradePolicy.AutomaticOSUpgradePolicy.EnableAutomaticOSUpgrade
		}
		updateProps.UpgradePolicy = &upgradePolicy
	}

	if d.HasChange("single_placement_group") {
		updateProps.SinglePlacementGroup = utils.Bool(d.Get("single_placement_group").(bool))
	}

	if d.HasChange("enable_automatic_updates") ||
		d.HasChange("custom_data") ||
		d.HasChange("provision_vm_agent") ||
		d.HasChange("secret") ||
		d.HasChange("timezone") {
		osProfile := compute.VirtualMachineScaleSetUpdateOSProfile{}

		if d.HasChange("enable_automatic_updates") || d.HasChange("provision_vm_agent") || d.HasChange("timezone") {
			windowsConfig := compute.WindowsConfiguration{}

			if d.HasChange("enable_automatic_updates") {
				if upgradeMode == compute.UpgradeModeAutomatic {
					return fmt.Errorf("`enable_automatic_updates` cannot be changed for when `upgrade_mode` is `Automatic`")
				}

				windowsConfig.EnableAutomaticUpdates = utils.Bool(d.Get("enable_automatic_updates").(bool))
			}

			if d.HasChange("provision_vm_agent") {
				windowsConfig.ProvisionVMAgent = utils.Bool(d.Get("provision_vm_agent").(bool))
			}

			if d.HasChange("timezone") {
				windowsConfig.TimeZone = utils.String(d.Get("timezone").(string))
			}

			osProfile.WindowsConfiguration = &windowsConfig
		}

		if d.HasChange("custom_data") {
			updateInstances = true

			// customData can only be sent if it's a base64 encoded string,
			// so it's not possible to remove this without tainting the resource
			if v, ok := d.GetOk("custom_data"); ok {
				osProfile.CustomData = utils.String(v.(string))
			}
		}

		if d.HasChange("secret") {
			secretsRaw := d.Get("secret").([]interface{})
			osProfile.Secrets = expandWindowsSecrets(secretsRaw)
		}

		updateProps.VirtualMachineProfile.OsProfile = &osProfile
	}

	if d.HasChange("data_disk") || d.HasChange("os_disk") || d.HasChange("source_image_id") || d.HasChange("source_image_reference") {
		updateInstances = true

		if updateProps.VirtualMachineProfile.StorageProfile == nil {
			updateProps.VirtualMachineProfile.StorageProfile = &compute.VirtualMachineScaleSetUpdateStorageProfile{}
		}

		if d.HasChange("data_disk") {
			ultraSSDEnabled := d.Get("additional_capabilities.0.ultra_ssd_enabled").(bool)
			dataDisks, err := ExpandVirtualMachineScaleSetDataDisk(d.Get("data_disk").([]interface{}), ultraSSDEnabled)
			if err != nil {
				return fmt.Errorf("expanding `data_disk`: %+v", err)
			}
			updateProps.VirtualMachineProfile.StorageProfile.DataDisks = dataDisks
		}

		if d.HasChange("os_disk") {
			osDiskRaw := d.Get("os_disk").([]interface{})
			updateProps.VirtualMachineProfile.StorageProfile.OsDisk = ExpandVirtualMachineScaleSetOSDiskUpdate(osDiskRaw)
		}

		if d.HasChange("source_image_id") || d.HasChange("source_image_reference") {
			sourceImageReferenceRaw := d.Get("source_image_reference").([]interface{})
			sourceImageId := d.Get("source_image_id").(string)
			sourceImageReference, err := expandSourceImageReference(sourceImageReferenceRaw, sourceImageId)
			if err != nil {
				return err
			}

			// Must include all storage profile properties when updating disk image.  See: https://github.com/hashicorp/terraform-provider-azurestack/issues/8273
			updateProps.VirtualMachineProfile.StorageProfile.DataDisks = existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.DataDisks
			updateProps.VirtualMachineProfile.StorageProfile.ImageReference = sourceImageReference
			updateProps.VirtualMachineProfile.StorageProfile.OsDisk = &compute.VirtualMachineScaleSetUpdateOSDisk{
				Caching:                 existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.OsDisk.Caching,
				WriteAcceleratorEnabled: existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.OsDisk.WriteAcceleratorEnabled,
				DiskSizeGB:              existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.OsDisk.DiskSizeGB,
				Image:                   existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.OsDisk.Image,
				VhdContainers:           existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.OsDisk.VhdContainers,
				ManagedDisk:             existing.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.OsDisk.ManagedDisk,
			}
		}
	}

	if d.HasChange("network_interface") {
		networkInterfacesRaw := d.Get("network_interface").([]interface{})
		networkInterfaces, err := ExpandVirtualMachineScaleSetNetworkInterfaceUpdate(networkInterfacesRaw)
		if err != nil {
			return fmt.Errorf("expanding `network_interface`: %+v", err)
		}

		updateProps.VirtualMachineProfile.NetworkProfile = &compute.VirtualMachineScaleSetUpdateNetworkProfile{
			NetworkInterfaceConfigurations: networkInterfaces,
		}

		healthProbeId := d.Get("health_probe_id").(string)
		if healthProbeId != "" {
			updateProps.VirtualMachineProfile.NetworkProfile.HealthProbe = &compute.APIEntityReference{
				ID: utils.String(healthProbeId),
			}
		}
	}

	if d.HasChange("boot_diagnostics") {
		updateInstances = true

		bootDiagnosticsRaw := d.Get("boot_diagnostics").([]interface{})
		updateProps.VirtualMachineProfile.DiagnosticsProfile = expandBootDiagnostics(bootDiagnosticsRaw)
	}

	if d.HasChange("do_not_run_extensions_on_overprovisioned_machines") {
		v := d.Get("do_not_run_extensions_on_overprovisioned_machines").(bool)
		updateProps.DoNotRunExtensionsOnOverprovisionedVMs = utils.Bool(v)
	}

	if d.HasChange("scale_in_policy") {
		scaleInPolicy := d.Get("scale_in_policy").(string)
		updateProps.ScaleInPolicy = &compute.ScaleInPolicy{
			Rules: &[]compute.VirtualMachineScaleSetScaleInRules{compute.VirtualMachineScaleSetScaleInRules(scaleInPolicy)},
		}
	}

	if d.HasChange("terminate_notification") {
		notificationRaw := d.Get("terminate_notification").([]interface{})
		updateProps.VirtualMachineProfile.ScheduledEventsProfile = ExpandVirtualMachineScaleSetScheduledEventsProfile(notificationRaw)
	}

	if d.HasChange("encryption_at_host_enabled") {
		if updateProps.VirtualMachineProfile.SecurityProfile == nil {
			updateProps.VirtualMachineProfile.SecurityProfile = &compute.SecurityProfile{}
		}
		updateProps.VirtualMachineProfile.SecurityProfile.EncryptionAtHost = utils.Bool(d.Get("encryption_at_host_enabled").(bool))
	}

	if d.HasChange("license_type") {
		license := d.Get("license_type").(string)
		if license == "" {
			// Only for create no specification is possible in the API. API does not allow empty string in update.
			// So removing attribute license_type from Terraform configuration if it was set to value other than 'None' would lead to an endless loop in apply.
			// To allow updating in this case set value explicitly to 'None'.
			license = "None"
		}
		updateProps.VirtualMachineProfile.LicenseType = &license
	}

	if d.HasChange("automatic_instance_repair") {
		automaticRepairsPolicyRaw := d.Get("automatic_instance_repair").([]interface{})
		automaticRepairsPolicy := ExpandVirtualMachineScaleSetAutomaticRepairsPolicy(automaticRepairsPolicyRaw)
		updateProps.AutomaticRepairsPolicy = automaticRepairsPolicy
	}

	if d.HasChange("plan") {
		planRaw := d.Get("plan").([]interface{})
		update.Plan = expandPlan(planRaw)
	}

	if d.HasChange("sku") || d.HasChange("instances") {
		// in-case ignore_changes is being used, since both fields are required
		// look up the current values and override them as needed
		sku := existing.Sku

		if d.HasChange("sku") {
			updateInstances = true

			sku.Name = utils.String(d.Get("sku").(string))
		}

		if d.HasChange("instances") {
			sku.Capacity = utils.Int64(int64(d.Get("instances").(int)))
		}

		update.Sku = sku
	}

	if d.HasChanges("extension") {
		updateInstances = true

		extensionProfile, _, err := expandVirtualMachineScaleSetExtensions(d.Get("extension").(*pluginsdk.Set).List())
		if err != nil {
			return err
		}
		updateProps.VirtualMachineProfile.ExtensionProfile = extensionProfile
	}

	if d.HasChange("tags") {
		update.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	update.VirtualMachineScaleSetUpdateProperties = &updateProps

	metaData := virtualMachineScaleSetUpdateMetaData{
		AutomaticOSUpgradeIsEnabled:  automaticOSUpgradeIsEnabled,
		CanRollInstancesWhenRequired: meta.(*clients.Client).Features.VirtualMachineScaleSet.RollInstancesWhenRequired,
		UpdateInstances:              updateInstances,
		Client:                       meta.(*clients.Client).Compute,
		Existing:                     existing,
		ID:                           id,
		OSType:                       compute.Windows,
	}

	if err := metaData.performUpdate(ctx, update); err != nil {
		return err
	}

	return resourceWindowsVirtualMachineScaleSetRead(d, meta)
}

func resourceWindowsVirtualMachineScaleSetRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualMachineScaleSetID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Windows Virtual Machine Scale Set %q was not found in Resource Group %q - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	var skuName *string
	var instances int
	if resp.Sku != nil {
		skuName = resp.Sku.Name
		if resp.Sku.Capacity != nil {
			instances = int(*resp.Sku.Capacity)
		}
	}
	d.Set("instances", instances)
	d.Set("sku", skuName)

	if err := d.Set("plan", flattenPlan(resp.Plan)); err != nil {
		return fmt.Errorf("setting `plan`: %+v", err)
	}

	if resp.VirtualMachineScaleSetProperties == nil {
		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): `properties` was nil", id.Name, id.ResourceGroup)
	}
	props := *resp.VirtualMachineScaleSetProperties

	if err := d.Set("additional_capabilities", FlattenVirtualMachineScaleSetAdditionalCapabilities(props.AdditionalCapabilities)); err != nil {
		return fmt.Errorf("setting `additional_capabilities`: %+v", props.AdditionalCapabilities)
	}

	if err := d.Set("automatic_instance_repair", FlattenVirtualMachineScaleSetAutomaticRepairsPolicy(props.AutomaticRepairsPolicy)); err != nil {
		return fmt.Errorf("setting `automatic_instance_repair`: %+v", err)
	}

	d.Set("do_not_run_extensions_on_overprovisioned_machines", props.DoNotRunExtensionsOnOverprovisionedVMs)
	d.Set("overprovision", props.Overprovision)
	d.Set("platform_fault_domain_count", props.PlatformFaultDomainCount)
	d.Set("single_placement_group", props.SinglePlacementGroup)
	d.Set("unique_id", props.UniqueID)

	var upgradeMode compute.UpgradeMode
	if policy := props.UpgradePolicy; policy != nil {
		upgradeMode = policy.Mode
		d.Set("upgrade_mode", string(policy.Mode))

		flattenedAutomatic := FlattenVirtualMachineScaleSetAutomaticOSUpgradePolicy(policy.AutomaticOSUpgradePolicy)
		if err := d.Set("automatic_os_upgrade_policy", flattenedAutomatic); err != nil {
			return fmt.Errorf("setting `automatic_os_upgrade_policy`: %+v", err)
		}
	}

	rule := string(compute.Default)
	if props.ScaleInPolicy != nil {
		if rules := props.ScaleInPolicy.Rules; rules != nil && len(*rules) > 0 {
			rule = string((*rules)[0])
		}
	}
	d.Set("scale_in_policy", rule)

	if profile := props.VirtualMachineProfile; profile != nil {
		if err := d.Set("boot_diagnostics", flattenBootDiagnostics(profile.DiagnosticsProfile)); err != nil {
			return fmt.Errorf("setting `boot_diagnostics`: %+v", err)
		}

		// defaulted since BillingProfile isn't returned if it's unset
		d.Set("license_type", profile.LicenseType)

		// the service just return empty when this is not assigned when provisioned
		// See discussion on https://github.com/Azure/azure-rest-api-specs/issues/10971
		if storageProfile := profile.StorageProfile; storageProfile != nil {
			if err := d.Set("os_disk", FlattenVirtualMachineScaleSetOSDisk(storageProfile.OsDisk)); err != nil {
				return fmt.Errorf("setting `os_disk`: %+v", err)
			}

			if err := d.Set("data_disk", FlattenVirtualMachineScaleSetDataDisk(storageProfile.DataDisks)); err != nil {
				return fmt.Errorf("setting `data_disk`: %+v", err)
			}

			if err := d.Set("source_image_reference", flattenSourceImageReference(storageProfile.ImageReference)); err != nil {
				return fmt.Errorf("setting `source_image_reference`: %+v", err)
			}

			var storageImageId string
			if storageProfile.ImageReference != nil && storageProfile.ImageReference.ID != nil {
				storageImageId = *storageProfile.ImageReference.ID
			}
			d.Set("source_image_id", storageImageId)
		}

		if osProfile := profile.OsProfile; osProfile != nil {
			// admin_password isn't returned, but it's a top level field so we can ignore it without consequence
			d.Set("admin_username", osProfile.AdminUsername)
			d.Set("computer_name_prefix", osProfile.ComputerNamePrefix)

			if err := d.Set("secret", flattenWindowsSecrets(osProfile.Secrets)); err != nil {
				return fmt.Errorf("setting `secret`: %+v", err)
			}

			if windows := osProfile.WindowsConfiguration; windows != nil {
				if err := d.Set("additional_unattend_content", flattenAdditionalUnattendContent(windows.AdditionalUnattendContent, d)); err != nil {
					return fmt.Errorf("setting `additional_unattend_content`: %+v", err)
				}

				enableAutomaticUpdates := false
				if windows.EnableAutomaticUpdates != nil {
					enableAutomaticUpdates = *windows.EnableAutomaticUpdates
				}

				// the API requires this is set to 'true' on submission (since it's now required for Windows VMSS's with
				// an Automatic Upgrade Mode configured) however it actually returns false from the API..
				// after a bunch of testing the least bad option appears to be not to set this if it's an Automatic Upgrade Mode
				if upgradeMode != compute.UpgradeModeAutomatic {
					d.Set("enable_automatic_updates", enableAutomaticUpdates)
				}

				d.Set("provision_vm_agent", windows.ProvisionVMAgent)
				d.Set("timezone", windows.TimeZone)

				if err := d.Set("winrm_listener", flattenWinRMListener(windows.WinRM)); err != nil {
					return fmt.Errorf("setting `winrm_listener`: %+v", err)
				}
			}
		}

		if nwProfile := profile.NetworkProfile; nwProfile != nil {
			flattenedNics := FlattenVirtualMachineScaleSetNetworkInterface(nwProfile.NetworkInterfaceConfigurations)
			if err := d.Set("network_interface", flattenedNics); err != nil {
				return fmt.Errorf("setting `network_interface`: %+v", err)
			}

			healthProbeId := ""
			if nwProfile.HealthProbe != nil && nwProfile.HealthProbe.ID != nil {
				healthProbeId = *nwProfile.HealthProbe.ID
			}
			d.Set("health_probe_id", healthProbeId)
		}

		if scheduleProfile := profile.ScheduledEventsProfile; scheduleProfile != nil {
			if err := d.Set("terminate_notification", FlattenVirtualMachineScaleSetScheduledEventsProfile(scheduleProfile)); err != nil {
				return fmt.Errorf("setting `terminate_notification`: %+v", err)
			}
		}

		extensionProfile, err := flattenVirtualMachineScaleSetExtensions(profile.ExtensionProfile, d)
		if err != nil {
			return fmt.Errorf("failed flattening `extension`: %+v", err)
		}
		d.Set("extension", extensionProfile)

		encryptionAtHostEnabled := false

		d.Set("encryption_at_host_enabled", encryptionAtHostEnabled)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceWindowsVirtualMachineScaleSetDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMScaleSetClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualMachineScaleSetID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil
		}

		return fmt.Errorf("retrieving Windows Virtual Machine Scale Set %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	// Sometimes VMSS's aren't fully deleted when the `Delete` call returns - as such we'll try to scale the cluster
	// to 0 nodes first, then delete the cluster - which should ensure there's no Network Interfaces kicking around
	// and work around this Azure API bug:
	// Original Error: Code="InUseSubnetCannotBeDeleted" Message="Subnet internal is in use by
	// /{nicResourceID}/|providers|Microsoft.Compute|virtualMachineScaleSets|acctestvmss-190923101253410278|virtualMachines|0|networkInterfaces|example/ipConfigurations/internal and cannot be deleted.
	// In order to delete the subnet, delete all the resources within the subnet. See aka.ms/deletesubnet.
	scaleToZeroOnDelete := meta.(*clients.Client).Features.VirtualMachineScaleSet.ScaleToZeroOnDelete
	if scaleToZeroOnDelete && resp.Sku != nil {
		resp.Sku.Capacity = utils.Int64(int64(0))

		log.Printf("[DEBUG] Scaling instances to 0 prior to deletion - this helps avoids networking issues within Azure")
		update := compute.VirtualMachineScaleSetUpdate{
			Sku: resp.Sku,
		}
		future, err := client.Update(ctx, id.ResourceGroup, id.Name, update)
		if err != nil {
			return fmt.Errorf("updating number of instances in Windows Virtual Machine Scale Set %q (Resource Group %q) to scale to 0: %+v", id.Name, id.ResourceGroup, err)
		}

		log.Printf("[DEBUG] Waiting for scaling of instances to 0 prior to deletion - this helps avoids networking issues within Azure")
		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("waiting for number of instances in Windows Virtual Machine Scale Set %q (Resource Group %q) to scale to 0: %+v", id.Name, id.ResourceGroup, err)
		}
		log.Printf("[DEBUG] Scaled instances to 0 prior to deletion - this helps avoids networking issues within Azure")
	} else {
		log.Printf("[DEBUG] Unable to scale instances to `0` since the `sku` block is nil - trying to delete anyway")
	}

	log.Printf("[DEBUG] Deleting Windows Virtual Machine Scale Set %q (Resource Group %q)..", id.Name, id.ResourceGroup)
	// @ArcturusZhang (mimicking from windows_virtual_machine_pluginsdk.go): sending `nil` here omits this value from being sent
	// which matches the previous behaviour - we're only splitting this out so it's clear why
	// TODO: support force deletion once it's out of Preview, if applicable
	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Windows Virtual Machine Scale Set %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	log.Printf("[DEBUG] Waiting for deletion of Windows Virtual Machine Scale Set %q (Resource Group %q)..", id.Name, id.ResourceGroup)
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Windows Virtual Machine Scale Set %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	log.Printf("[DEBUG] Deleted Windows Virtual Machine Scale Set %q (Resource Group %q).", id.Name, id.ResourceGroup)

	return nil
}
