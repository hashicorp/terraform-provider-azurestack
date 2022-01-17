package network

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Network"
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
		/*"azurestack_network_interface":                         dataSourceNetworkInterface(),
		"azurestack_network_security_group":                    dataSourceNetworkSecurityGroup(),
		"azurestack_public_ip":                                 dataSourcePublicIP(),
		"azurestack_public_ips":                                dataSourcePublicIPs(),
		"azurestack_route_table":                               dataSourceRouteTable(),
		"azurestack_subnet":                                    dataSourceSubnet(),
		"azurestack_virtual_network_gateway":                   dataSourceVirtualNetworkGateway(),
		"azurestack_virtual_network_gateway_connection":        dataSourceVirtualNetworkGatewayConnection(),
		"azurestack_virtual_network":                           dataSourceVirtualNetwork(),
		"azurestack_local_network_gateway":                     dataSourceLocalNetworkGateway(),*/
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		/*"azurestack_local_network_gateway":                    resourceLocalNetworkGateway(),
		"azurestack_public_ip":                                 resourcePublicIp(),
		"azurestack_network_security_group":                    resourceNetworkSecurityGroup(),
		"azurestack_network_security_rule":                     resourceNetworkSecurityRule(),
		"azurestack_route_table":                               resourceRouteTable(),
		"azurestack_route":                                     resourceRoute(),
		"azurestack_virtual_network_gateway_connection":        resourceVirtualNetworkGatewayConnection(),
		"azurestack_virtual_network_gateway":                   resourceVirtualNetworkGateway(),
		"azurestack_virtual_network":                           resourceVirtualNetwork(),*/
	}
}
