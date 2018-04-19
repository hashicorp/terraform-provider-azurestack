package azurestack

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/storage/mgmt/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/authentication"
)

// ArmClient contains the handles to all the specific Azure Resource Manager
// resource classes' respective clients.
type ArmClient struct {
	clientId       string
	tenantId       string
	subscriptionId string
	location       string
	armEndpoint    string
	environment    azure.Environment

	vmClient             compute.VirtualMachinesClient
	storageServiceClient storage.AccountsClient

	vnetClient     network.VirtualNetworksClient
	secGroupClient network.SecurityGroupsClient
	publicIPClient network.PublicIPAddressesClient
	subnetClient   network.SubnetsClient
	nicClient      network.InterfacesClient

	resourceGroupsClient resources.GroupsClient

	StopContext context.Context
}

// getArmClient is a helper method which returns a fully instantiated
// *ArmClient based on the Config's current settings.
func getArmClient(c *authentication.Config) (*ArmClient, error) {
	// detect cloud from environment
	env, envErr := azure.EnvironmentFromURL(c.ARMEndpoint)
	if envErr != nil {
		// try again with wrapped value to support readable values like german instead of AZUREGERMANCLOUD
		wrapped := fmt.Sprintf("AZURE%sCLOUD", c.Environment)
		var innerErr error
		if env, innerErr = azure.EnvironmentFromName(wrapped); innerErr != nil {
			return nil, envErr
		}
	}

	// client declarations:
	client := ArmClient{
		clientId:       c.ClientID,
		tenantId:       c.TenantID,
		subscriptionId: c.SubscriptionID,
		armEndpoint:    c.ARMEndpoint,
		environment:    env,
	}

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return nil, err
	}

	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, fmt.Errorf("Unable to configure OAuthConfig for tenant %s", c.TenantID)
	}

	endpoint := env.ResourceManagerEndpoint

	sender := autorest.CreateSender(withRequestLogging())

	auth, err := getAuthorizationToken(c, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, fmt.Errorf("Unable to create token, reason %s", err)
	}

	client.registerComputeClient(endpoint, c.SubscriptionID, auth, sender)
	client.registerStorageClient(endpoint, c.SubscriptionID, auth, sender)
	client.registerNetworks(endpoint, c.SubscriptionID, auth, sender)
	client.registerGroupsClient(endpoint, c.SubscriptionID, auth, sender)

	client.StopContext = context.Background()

	return &client, nil
}

func getAuthorizationToken(c *authentication.Config, oauthConfig *adal.OAuthConfig, endpoint string) (*autorest.BearerAuthorizer, error) {
	spt, err := adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, endpoint)
	if err != nil {
		return nil, err
	}
	token := autorest.NewBearerAuthorizer(spt)

	return token, nil
}

func (c *ArmClient) registerComputeClient(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	vmClient := compute.NewVirtualMachinesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&vmClient.Client, auth)
	c.vmClient = vmClient
}

func (c *ArmClient) registerStorageClient(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	storageAccountsClient := storage.NewAccountsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&storageAccountsClient.Client, auth)
	c.storageServiceClient = storageAccountsClient
}

func (c *ArmClient) registerNetworks(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	vnetClient := network.NewVirtualNetworksClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&vnetClient.Client, auth)
	c.vnetClient = vnetClient

	nsgClient := network.NewSecurityGroupsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&nsgClient.Client, auth)
	c.secGroupClient = nsgClient

	ipClient := network.NewPublicIPAddressesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&ipClient.Client, auth)
	c.publicIPClient = ipClient

	subnetsClient := network.NewSubnetsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&subnetsClient.Client, auth)
	c.subnetClient = subnetsClient

	nicClient := network.NewInterfacesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&nicClient.Client, auth)
	c.nicClient = nicClient
}

func (c *ArmClient) configureClient(client *autorest.Client, auth autorest.Authorizer) {
	setUserAgent(client)
	client.Authorizer = auth
	client.Sender = autorest.CreateSender(withRequestLogging())
	client.PollingDuration = 60 * time.Minute
}

func (c *ArmClient) registerGroupsClient(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	groupsClient := resources.NewGroupsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&groupsClient.Client, auth)
	c.resourceGroupsClient = groupsClient
}

func withRequestLogging() autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			// dump request to wire format
			if dump, err := httputil.DumpRequestOut(r, true); err == nil {
				log.Printf("[DEBUG] AzureRM Request: \n%s\n", dump)
			} else {
				// fallback to basic message
				log.Printf("[DEBUG] AzureRM Request: %s to %s\n", r.Method, r.URL)
			}

			resp, err := s.Do(r)
			if resp != nil {
				// dump response to wire format
				if dump, err := httputil.DumpResponse(resp, true); err == nil {
					log.Printf("[DEBUG] AzureRM Response for %s: \n%s\n", r.URL, dump)
				} else {
					// fallback to basic message
					log.Printf("[DEBUG] AzureRM Response: %s for %s\n", resp.Status, r.URL)
				}
			} else {
				log.Printf("[DEBUG] Request to %s completed with no response", r.URL)
			}
			return resp, err
		})
	}
}

func setUserAgent(client *autorest.Client) {
	tfVersion := fmt.Sprintf("HashiCorp-Terraform-v%s", terraform.VersionString())

	// if the user agent already has a value append the Terraform user agent string
	if curUserAgent := client.UserAgent; curUserAgent != "" {
		client.UserAgent = fmt.Sprintf("%s;%s", curUserAgent, tfVersion)
	} else {
		client.UserAgent = tfVersion
	}

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		client.UserAgent = fmt.Sprintf("%s;%s", client.UserAgent, azureAgent)
	}
}
