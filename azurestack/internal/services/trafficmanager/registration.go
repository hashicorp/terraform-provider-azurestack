package trafficmanager

import (
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Traffic Manager"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Network",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_traffic_manager_geographical_location": dataSourceArmTrafficManagerGeographicalLocation(),
		"azurerm_traffic_manager_profile":               dataSourceArmTrafficManagerProfile(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_traffic_manager_endpoint": resourceArmTrafficManagerEndpoint(),
		"azurerm_traffic_manager_profile":  resourceArmTrafficManagerProfile(),
	}
}
