package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/appplatform/mgmt/2020-11-01-preview/appplatform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/common"
)

type Client struct {
	AppsClient               *appplatform.AppsClient
	BindingsClient           *appplatform.BindingsClient
	CertificatesClient       *appplatform.CertificatesClient
	ConfigServersClient      *appplatform.ConfigServersClient
	CustomDomainsClient      *appplatform.CustomDomainsClient
	MonitoringSettingsClient *appplatform.MonitoringSettingsClient
	DeploymentsClient        *appplatform.DeploymentsClient
	ServicesClient           *appplatform.ServicesClient
}

func NewClient(o *common.ClientOptions) *Client {
	appsClient := appplatform.NewAppsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&appsClient.Client, o.ResourceManagerAuthorizer)

	bindingsClient := appplatform.NewBindingsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&bindingsClient.Client, o.ResourceManagerAuthorizer)

	certificatesClient := appplatform.NewCertificatesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&certificatesClient.Client, o.ResourceManagerAuthorizer)

	configServersClient := appplatform.NewConfigServersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&configServersClient.Client, o.ResourceManagerAuthorizer)

	customDomainsClient := appplatform.NewCustomDomainsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&customDomainsClient.Client, o.ResourceManagerAuthorizer)

	deploymentsClient := appplatform.NewDeploymentsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&deploymentsClient.Client, o.ResourceManagerAuthorizer)

	monitoringSettingsClient := appplatform.NewMonitoringSettingsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&monitoringSettingsClient.Client, o.ResourceManagerAuthorizer)

	servicesClient := appplatform.NewServicesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&servicesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		AppsClient:               &appsClient,
		BindingsClient:           &bindingsClient,
		CertificatesClient:       &certificatesClient,
		ConfigServersClient:      &configServersClient,
		CustomDomainsClient:      &customDomainsClient,
		DeploymentsClient:        &deploymentsClient,
		MonitoringSettingsClient: &monitoringSettingsClient,
		ServicesClient:           &servicesClient,
	}
}
