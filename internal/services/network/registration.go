// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		"azurestack_network_interface":                  networkInterfaceDataSource(),
		"azurestack_public_ip":                          publicIPDataSource(),
		"azurestack_public_ips":                         publicIPsDataSource(),
		"azurestack_route_table":                        routeTableDataSource(),
		"azurestack_subnet":                             subnetDataSource(),
		"azurestack_virtual_network":                    virtualNetworkDataSource(),
		"azurestack_network_security_group":             networkSecurityGroupDataSource(),
		"azurestack_virtual_network_gateway":            virtualNetworkGatewayDataSource(),
		"azurestack_virtual_network_gateway_connection": virtualNetworkGatewayConnectionDataSource(),
		"azurestack_local_network_gateway":              localNetworkGatewayDataSource(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_network_interface":                                  networkInterface(),
		"azurestack_public_ip":                                          publicIp(),
		"azurestack_route_table":                                        routeTable(),
		"azurestack_route":                                              resourceRoute(),
		"azurestack_subnet":                                             subnet(),
		"azurestack_virtual_network":                                    virtualNetwork(),
		"azurestack_network_security_group":                             networkSecurityGroup(),
		"azurestack_network_security_rule":                              networkSecurityRule(),
		"azurestack_virtual_network_gateway_connection":                 virtualNetworkGatewayConnection(),
		"azurestack_virtual_network_gateway":                            virtualNetworkGateway(),
		"azurestack_local_network_gateway":                              localNetworkGateway(),
		"azurestack_virtual_network_peering":                            virtualNetworkPeering(),
		"azurestack_network_interface_backend_address_pool_association": loadBalancerBackendAddressPoolAssociation(),
		"azurestack_subnet_network_security_group_association":          subnetNetworkSecurityGroupAssociation(),
		"azurestack_network_interface_security_group_association":       networkInterfaceSecurityGroupAssociation(),
	}
}
