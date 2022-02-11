package compute

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Compute"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Compute",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_availability_set": availabilitySetDataSource(),
		"azurestack_managed_disk":     managedDiskDataSource(),
		"azurestack_platform_image":   platformImageDataSource(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurestack_availability_set":                     availabilitySet(),
		"azurestack_linux_virtual_machine":                linuxVirtualMachine(),
		"azurestack_managed_disk":                         managedDisk(),
		"azurestack_virtual_machine":                      virtualMachine(),
		"azurestack_virtual_machine_data_disk_attachment": virtualMachineDataDiskAttachment(),
		"azurestack_virtual_machine_extension":            virtualMachineExtension(),
		"azurestack_virtual_machine_scale_set":            virtualMachineScaleSet(),
		"azurestack_virtual_machine_scale_set_extension":  virtualMachineScaleSetExtension(),
		"azurestack_windows_virtual_machine":              windowsVirtualMachine(),
	}

	return resources
}
