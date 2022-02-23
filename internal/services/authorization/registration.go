package authorization

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Authorization"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Authorization",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_client_config":   clientConfigDataSource(),
		"azurestack_role_definition": dataSourceArmRoleDefinition(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_role_assignment": resourceArmRoleAssignment(),
		"azurestack_role_definition": resourceArmRoleDefinition(),
	}
}
