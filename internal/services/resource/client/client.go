package client

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
)

type Client struct {
	DeploymentsClient *resources.DeploymentsClient
	GroupsClient      *resources.GroupsClient
	ProvidersClient   *resources.ProvidersClient
	ResourcesClient   *resources.Client

	options *common.ClientOptions
}

func NewClient(o *common.ClientOptions) *Client {
	deploymentsClient := resources.NewDeploymentsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&deploymentsClient.Client, o.ResourceManagerAuthorizer)

	groupsClient := resources.NewGroupsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&groupsClient.Client, o.ResourceManagerAuthorizer)

	// this has to come from the Profile since this is shared with Stack
	providersClient := resources.NewProvidersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&providersClient.Client, o.ResourceManagerAuthorizer)

	resourcesClient := resources.NewClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&resourcesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		DeploymentsClient: &deploymentsClient,
		GroupsClient:      &groupsClient,
		ProvidersClient:   &providersClient,
		ResourcesClient:   &resourcesClient,

		options: o,
	}
}
