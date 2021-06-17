package client

import (
	healthcare "github.com/Azure/azure-sdk-for-go/services/healthcareapis/mgmt/2020-03-30/healthcareapis"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/common"
)

type Client struct {
	HealthcareServiceClient *healthcare.ServicesClient
}

func NewClient(o *common.ClientOptions) *Client {
	HealthcareServiceClient := healthcare.NewServicesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&HealthcareServiceClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		HealthcareServiceClient: &HealthcareServiceClient,
	}
}
