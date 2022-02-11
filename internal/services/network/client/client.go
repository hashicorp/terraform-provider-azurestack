package client

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
)

type Client struct {
	ApplicationSecurityGroupsClient *network.ApplicationSecurityGroupsClient
	InterfacesClient                *network.InterfacesClient
	LocalNetworkGatewaysClient      *network.LocalNetworkGatewaysClient
	PublicIPsClient                 *network.PublicIPAddressesClient
	RoutesClient                    *network.RoutesClient
	RouteTablesClient               *network.RouteTablesClient
	SecurityGroupClient             *network.SecurityGroupsClient
	SecurityRuleClient              *network.SecurityRulesClient
	SubnetsClient                   *network.SubnetsClient
	VnetGatewayConnectionsClient    *network.VirtualNetworkGatewayConnectionsClient
	VnetGatewayClient               *network.VirtualNetworkGatewaysClient
	VnetClient                      *network.VirtualNetworksClient
}

func NewClient(o *common.ClientOptions) *Client {
	ApplicationGatewaysClient := network.NewApplicationGatewaysClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ApplicationGatewaysClient.Client, o.ResourceManagerAuthorizer)

	ApplicationSecurityGroupsClient := network.NewApplicationSecurityGroupsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ApplicationSecurityGroupsClient.Client, o.ResourceManagerAuthorizer)

	ConnectionMonitorsClient := network.NewConnectionMonitorsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ConnectionMonitorsClient.Client, o.ResourceManagerAuthorizer)

	ExpressRouteAuthsClient := network.NewExpressRouteCircuitAuthorizationsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ExpressRouteAuthsClient.Client, o.ResourceManagerAuthorizer)

	ExpressRouteCircuitsClient := network.NewExpressRouteCircuitsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ExpressRouteCircuitsClient.Client, o.ResourceManagerAuthorizer)

	ExpressRoutePeeringsClient := network.NewExpressRouteCircuitPeeringsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ExpressRoutePeeringsClient.Client, o.ResourceManagerAuthorizer)

	InterfacesClient := network.NewInterfacesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&InterfacesClient.Client, o.ResourceManagerAuthorizer)

	LocalNetworkGatewaysClient := network.NewLocalNetworkGatewaysClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&LocalNetworkGatewaysClient.Client, o.ResourceManagerAuthorizer)

	VnetClient := network.NewVirtualNetworksClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&VnetClient.Client, o.ResourceManagerAuthorizer)

	PacketCapturesClient := network.NewPacketCapturesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&PacketCapturesClient.Client, o.ResourceManagerAuthorizer)

	VnetPeeringsClient := network.NewVirtualNetworkPeeringsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&VnetPeeringsClient.Client, o.ResourceManagerAuthorizer)

	PublicIPsClient := network.NewPublicIPAddressesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&PublicIPsClient.Client, o.ResourceManagerAuthorizer)

	RoutesClient := network.NewRoutesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&RoutesClient.Client, o.ResourceManagerAuthorizer)

	RouteFiltersClient := network.NewRouteFiltersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&RouteFiltersClient.Client, o.ResourceManagerAuthorizer)

	RouteTablesClient := network.NewRouteTablesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&RouteTablesClient.Client, o.ResourceManagerAuthorizer)

	SecurityGroupClient := network.NewSecurityGroupsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&SecurityGroupClient.Client, o.ResourceManagerAuthorizer)

	SecurityRuleClient := network.NewSecurityRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&SecurityRuleClient.Client, o.ResourceManagerAuthorizer)

	SubnetsClient := network.NewSubnetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&SubnetsClient.Client, o.ResourceManagerAuthorizer)

	VnetGatewayClient := network.NewVirtualNetworkGatewaysClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&VnetGatewayClient.Client, o.ResourceManagerAuthorizer)

	VnetGatewayConnectionsClient := network.NewVirtualNetworkGatewayConnectionsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&VnetGatewayConnectionsClient.Client, o.ResourceManagerAuthorizer)

	WatcherClient := network.NewWatchersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&WatcherClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		ApplicationSecurityGroupsClient: &ApplicationSecurityGroupsClient,
		InterfacesClient:                &InterfacesClient,
		LocalNetworkGatewaysClient:      &LocalNetworkGatewaysClient,
		PublicIPsClient:                 &PublicIPsClient,
		RoutesClient:                    &RoutesClient,
		RouteTablesClient:               &RouteTablesClient,
		SecurityGroupClient:             &SecurityGroupClient,
		SecurityRuleClient:              &SecurityRuleClient,
		SubnetsClient:                   &SubnetsClient,
		VnetGatewayConnectionsClient:    &VnetGatewayConnectionsClient,
		VnetGatewayClient:               &VnetGatewayClient,
		VnetClient:                      &VnetClient,
	}
}
