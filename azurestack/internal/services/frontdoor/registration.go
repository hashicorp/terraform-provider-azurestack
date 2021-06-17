package frontdoor

import (
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

// TODO: move this to the network one

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "FrontDoor"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Network",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_frontdoor":                            resourceFrontDoor(),
		"azurerm_frontdoor_firewall_policy":            resourceFrontDoorFirewallPolicy(),
		"azurerm_frontdoor_custom_https_configuration": resourceFrontDoorCustomHttpsConfiguration(),
	}
}
