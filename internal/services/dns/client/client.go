package client

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/dns/mgmt/dns"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
)

type Client struct {
	RecordSetsClient *dns.RecordSetsClient
	ZonesClient      *dns.ZonesClient
}

func NewClient(o *common.ClientOptions) *Client {
	RecordSetsClient := dns.NewRecordSetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&RecordSetsClient.Client, o.ResourceManagerAuthorizer)

	ZonesClient := dns.NewZonesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&ZonesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		RecordSetsClient: &RecordSetsClient,
		ZonesClient:      &ZonesClient,
	}
}
