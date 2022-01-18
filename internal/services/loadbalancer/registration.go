package loadbalancer

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/sdk"
)

var _ sdk.UntypedServiceRegistration = Registration{}

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Load Balancer"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Load Balancer",
	}
}

// DataSources returns a list of Data Sources supported by this Service
func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_lb":                      loadBalancerDataSource(),
		"azurestack_lb_backend_address_pool": loadBalancerBackendAddressPoolDataSource(),
		"azurestack_lb_rule":                 loadBalancerRuleDataSource(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_lb_backend_address_pool": loadBalancerBackendAddressPool(),
		"azurestack_lb_nat_pool":             loadBalancerNatPool(),
		"azurestack_lb_nat_rule":             loadBalancerNatRule(),
		"azurestack_lb_probe":                loadBalancerProbe(),
		"azurestack_lb_rule":                 loadBalancerRule(),
		"azurestack_lb":                      loadBalancer(),
	}
}
