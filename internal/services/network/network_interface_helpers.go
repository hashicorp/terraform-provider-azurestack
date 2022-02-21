package network

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
)

type networkInterfaceUpdateInformation struct {
	applicationGatewayBackendAddressPoolIDs []string
	applicationSecurityGroupIDs             []string
	loadBalancerBackendAddressPoolIDs       []string
	loadBalancerInboundNatRuleIDs           []string
	networkSecurityGroupID                  string
}

func parseFieldsFromNetworkInterface(input network.InterfacePropertiesFormat) networkInterfaceUpdateInformation {
	networkSecurityGroupId := ""
	if input.NetworkSecurityGroup != nil && input.NetworkSecurityGroup.ID != nil {
		networkSecurityGroupId = *input.NetworkSecurityGroup.ID
	}

	mapToSlice := func(input map[string]struct{}) []string {
		output := make([]string, 0)

		for id := range input {
			output = append(output, id)
		}

		return output
	}

	applicationSecurityGroupIds := make(map[string]struct{})
	applicationGatewayBackendAddressPoolIds := make(map[string]struct{})
	loadBalancerBackendAddressPoolIds := make(map[string]struct{})
	loadBalancerInboundNatRuleIds := make(map[string]struct{})

	if input.IPConfigurations != nil {
		for _, v := range *input.IPConfigurations {
			if v.InterfaceIPConfigurationPropertiesFormat == nil {
				continue
			}

			props := *v.InterfaceIPConfigurationPropertiesFormat
			if props.ApplicationSecurityGroups != nil {
				for _, asg := range *props.ApplicationSecurityGroups {
					if asg.ID != nil {
						applicationSecurityGroupIds[*asg.ID] = struct{}{}
					}
				}
			}

			if props.ApplicationGatewayBackendAddressPools != nil {
				for _, pool := range *props.ApplicationGatewayBackendAddressPools {
					if pool.ID != nil {
						applicationGatewayBackendAddressPoolIds[*pool.ID] = struct{}{}
					}
				}
			}

			if props.LoadBalancerBackendAddressPools != nil {
				for _, pool := range *props.LoadBalancerBackendAddressPools {
					if pool.ID != nil {
						loadBalancerBackendAddressPoolIds[*pool.ID] = struct{}{}
					}
				}
			}

			if props.LoadBalancerInboundNatRules != nil {
				for _, rule := range *props.LoadBalancerInboundNatRules {
					if rule.ID != nil {
						loadBalancerInboundNatRuleIds[*rule.ID] = struct{}{}
					}
				}
			}
		}
	}

	return networkInterfaceUpdateInformation{
		applicationGatewayBackendAddressPoolIDs: mapToSlice(applicationGatewayBackendAddressPoolIds),
		applicationSecurityGroupIDs:             mapToSlice(applicationSecurityGroupIds),
		loadBalancerBackendAddressPoolIDs:       mapToSlice(loadBalancerBackendAddressPoolIds),
		loadBalancerInboundNatRuleIDs:           mapToSlice(loadBalancerInboundNatRuleIds),
		networkSecurityGroupID:                  networkSecurityGroupId,
	}
}

func mapFieldsToNetworkInterface(input *[]network.InterfaceIPConfiguration, info networkInterfaceUpdateInformation) *[]network.InterfaceIPConfiguration {
	output := input

	applicationSecurityGroups := make([]network.ApplicationSecurityGroup, 0)
	for _, id := range info.applicationSecurityGroupIDs {
		applicationSecurityGroups = append(applicationSecurityGroups, network.ApplicationSecurityGroup{
			ID: pointer.FromString(id),
		})
	}

	applicationGatewayBackendAddressPools := make([]network.ApplicationGatewayBackendAddressPool, 0)
	for _, id := range info.applicationGatewayBackendAddressPoolIDs {
		applicationGatewayBackendAddressPools = append(applicationGatewayBackendAddressPools, network.ApplicationGatewayBackendAddressPool{
			ID: pointer.FromString(id),
		})
	}

	loadBalancerBackendAddressPools := make([]network.BackendAddressPool, 0)
	for _, id := range info.loadBalancerBackendAddressPoolIDs {
		loadBalancerBackendAddressPools = append(loadBalancerBackendAddressPools, network.BackendAddressPool{
			ID: pointer.FromString(id),
		})
	}

	loadBalancerInboundNatRules := make([]network.InboundNatRule, 0)
	for _, id := range info.loadBalancerInboundNatRuleIDs {
		loadBalancerInboundNatRules = append(loadBalancerInboundNatRules, network.InboundNatRule{
			ID: pointer.FromString(id),
		})
	}

	for _, config := range *output {
		if config.InterfaceIPConfigurationPropertiesFormat == nil {
			continue
		}

		if config.InterfaceIPConfigurationPropertiesFormat.PrivateIPAddressVersion != network.IPv4 {
			continue
		}

		config.ApplicationSecurityGroups = &applicationSecurityGroups
		config.ApplicationGatewayBackendAddressPools = &applicationGatewayBackendAddressPools
		config.LoadBalancerBackendAddressPools = &loadBalancerBackendAddressPools
		config.LoadBalancerInboundNatRules = &loadBalancerInboundNatRules
	}

	return output
}

func FindNetworkInterfaceIPConfiguration(input *[]network.InterfaceIPConfiguration, name string) *network.InterfaceIPConfiguration {
	if input == nil {
		return nil
	}

	for _, v := range *input {
		if v.Name == nil {
			continue
		}

		if *v.Name == name {
			return &v
		}
	}

	return nil
}

func updateNetworkInterfaceIPConfiguration(config network.InterfaceIPConfiguration, configs *[]network.InterfaceIPConfiguration) *[]network.InterfaceIPConfiguration {
	output := make([]network.InterfaceIPConfiguration, 0)
	if configs == nil {
		return &output
	}

	for _, v := range *configs {
		if v.Name == nil {
			continue
		}

		if *v.Name != *config.Name {
			output = append(output, v)
		} else {
			output = append(output, config)
		}
	}

	return &output
}
