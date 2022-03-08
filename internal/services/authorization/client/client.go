package client

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
)

type Client struct {
	RoleAssignmentsClient   *authorization.RoleAssignmentsClient
	RoleDefinitionsClient   *authorization.RoleDefinitionsClient
	ServicePrincipalsClient *graphrbac.ServicePrincipalsClient
}

func NewClient(o *common.ClientOptions) *Client {
	servicePrincipalsClient := graphrbac.NewServicePrincipalsClientWithBaseURI(o.GraphEndpoint, o.TenantID)
	o.ConfigureClient(&servicePrincipalsClient.Client, o.GraphAuthorizer)

	return &Client{
		ServicePrincipalsClient: &servicePrincipalsClient,
	}
}
