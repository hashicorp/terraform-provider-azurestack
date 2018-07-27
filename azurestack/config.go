package azurestack

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/storage/mgmt/storage"
	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2016-04-01/dns"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	mainStorage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/authentication"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/utils"
)

// ArmClient contains the handles to all the specific Azure Resource Manager
// resource classes' respective clients.
type ArmClient struct {
	clientId                 string
	tenantId                 string
	subscriptionId           string
	usingServicePrincipal    bool
	environment              azure.Environment
	skipProviderRegistration bool

	StopContext context.Context

	// Authentication
	servicePrincipalsClient graphrbac.ServicePrincipalsClient

	// Compute
	availSetClient    compute.AvailabilitySetsClient
	vmExtensionClient compute.VirtualMachineExtensionsClient

	// DNS
	dnsClient   dns.RecordSetsClient
	zonesClient dns.ZonesClient

	// Networking
	ifaceClient   network.InterfacesClient
	secRuleClient network.SecurityRulesClient

	// Resources
	providersClient resources.ProvidersClient
	resourcesClient resources.Client

	vmClient             compute.VirtualMachinesClient
	vmScaleSetClient     compute.VirtualMachineScaleSetsClient
	storageServiceClient storage.AccountsClient

	vnetClient     network.VirtualNetworksClient
	secGroupClient network.SecurityGroupsClient
	publicIPClient network.PublicIPAddressesClient
	subnetClient   network.SubnetsClient
	nicClient      network.InterfacesClient

	resourceGroupsClient resources.GroupsClient
}

func (c *ArmClient) configureClient(client *autorest.Client, auth autorest.Authorizer) {
	setUserAgent(client)
	client.Authorizer = auth
	client.Sender = autorest.CreateSender(withRequestLogging())
	client.SkipResourceProviderRegistration = c.skipProviderRegistration
	client.PollingDuration = 60 * time.Minute
}

func withRequestLogging() autorest.SendDecorator {
	return func(s autorest.Sender) autorest.Sender {
		return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			// dump request to wire format
			if dump, err := httputil.DumpRequestOut(r, true); err == nil {
				log.Printf("[DEBUG] azurestack Request: \n%s\n", dump)
			} else {
				// fallback to basic message
				log.Printf("[DEBUG] azurestack Request: %s to %s\n", r.Method, r.URL)
			}

			resp, err := s.Do(r)
			if resp != nil {
				// dump response to wire format
				if dump, err := httputil.DumpResponse(resp, true); err == nil {
					log.Printf("[DEBUG] azurestack Response for %s: \n%s\n", r.URL, dump)
				} else {
					// fallback to basic message
					log.Printf("[DEBUG] azurestack Response: %s for %s\n", resp.Status, r.URL)
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

func getAuthorizationToken(c *authentication.Config, oauthConfig *adal.OAuthConfig, endpoint string) (*autorest.BearerAuthorizer, error) {
	// TODO: support Azure CLI auth / Service Principal Certificate auth
	spt, err := adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, endpoint)
	if err != nil {
		return nil, err
	}

	auth := autorest.NewBearerAuthorizer(spt)
	return auth, nil
}

// getArmClient is a helper method which returns a fully instantiated
// *ArmClient based on the Config's current settings.
func getArmClient(c *authentication.Config) (*ArmClient, error) {

	env, err := azure.EnvironmentFromURL(c.ARMEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Cannot load environment, reason: %s", err)
	}

	// client declarations:
	client := ArmClient{
		clientId:                 c.ClientID,
		tenantId:                 c.TenantID,
		subscriptionId:           c.SubscriptionID,
		environment:              env,
		usingServicePrincipal:    c.ClientSecret != "",
		skipProviderRegistration: c.SkipProviderRegistration,
	}

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		return nil, err
	}

	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, fmt.Errorf("Unable to configure OAuthConfig for tenant %s", c.TenantID)
	}

	sender := autorest.CreateSender(withRequestLogging())

	// Resource Manager endpoints
	endpoint := env.ResourceManagerEndpoint

	// Instead of the same enpoint use token audience to get the correct token.
	auth, err := getAuthorizationToken(c, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, err
	}

	// Graph Endpoints
	graphEndpoint := env.GraphEndpoint
	graphAuth, err := getAuthorizationToken(c, oauthConfig, graphEndpoint)
	if err != nil {
		return nil, err
	}

	client.registerAuthentication(endpoint, graphEndpoint, c.SubscriptionID, c.TenantID, auth, graphAuth, sender)
	client.registerComputeClients(endpoint, c.SubscriptionID, auth, sender)
	client.registerDNSClients(endpoint, c.SubscriptionID, auth, sender)
	client.registerNetworkingClients(endpoint, c.SubscriptionID, auth, sender)
	client.registerResourcesClients(endpoint, c.SubscriptionID, auth)
	client.registerStorageClients(endpoint, c.SubscriptionID, auth)

	return &client, nil
}

func (c *ArmClient) registerComputeClients(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	availabilitySetsClient := compute.NewAvailabilitySetsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&availabilitySetsClient.Client, auth)
	c.availSetClient = availabilitySetsClient

	extensionsClient := compute.NewVirtualMachineExtensionsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&extensionsClient.Client, auth)
	c.vmExtensionClient = extensionsClient

	scaleSetsClient := compute.NewVirtualMachineScaleSetsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&scaleSetsClient.Client, auth)
	c.vmScaleSetClient = scaleSetsClient

	virtualMachinesClient := compute.NewVirtualMachinesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&virtualMachinesClient.Client, auth)
	c.vmClient = virtualMachinesClient
}

func (c *ArmClient) registerDNSClients(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	dn := dns.NewRecordSetsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&dn.Client, auth)
	c.dnsClient = dn

	zo := dns.NewZonesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&zo.Client, auth)
	c.zonesClient = zo
}

func (c *ArmClient) registerNetworkingClients(endpoint, subscriptionId string, auth autorest.Authorizer, sender autorest.Sender) {
	interfacesClient := network.NewInterfacesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&interfacesClient.Client, auth)
	c.ifaceClient = interfacesClient

	networksClient := network.NewVirtualNetworksClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&networksClient.Client, auth)
	c.vnetClient = networksClient

	publicIPAddressesClient := network.NewPublicIPAddressesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&publicIPAddressesClient.Client, auth)
	c.publicIPClient = publicIPAddressesClient

	securityGroupsClient := network.NewSecurityGroupsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&securityGroupsClient.Client, auth)
	c.secGroupClient = securityGroupsClient

	securityRulesClient := network.NewSecurityRulesClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&securityRulesClient.Client, auth)
	c.secRuleClient = securityRulesClient

	subnetsClient := network.NewSubnetsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&subnetsClient.Client, auth)
	c.subnetClient = subnetsClient
}

func (c *ArmClient) registerResourcesClients(endpoint, subscriptionId string, auth autorest.Authorizer) {
	resourcesClient := resources.NewClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&resourcesClient.Client, auth)
	c.resourcesClient = resourcesClient

	resourceGroupsClient := resources.NewGroupsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&resourceGroupsClient.Client, auth)
	c.resourceGroupsClient = resourceGroupsClient

	providersClient := resources.NewProvidersClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&providersClient.Client, auth)
	c.providersClient = providersClient
}

func (c *ArmClient) registerStorageClients(endpoint, subscriptionId string, auth autorest.Authorizer) {
	accountsClient := storage.NewAccountsClientWithBaseURI(endpoint, subscriptionId)
	c.configureClient(&accountsClient.Client, auth)
	c.storageServiceClient = accountsClient
}

var (
	storageKeyCacheMu sync.RWMutex
	storageKeyCache   = make(map[string]string)
)

func (armClient *ArmClient) getKeyForStorageAccount(ctx context.Context, resourceGroupName, storageAccountName string) (string, bool, error) {
	cacheIndex := resourceGroupName + "/" + storageAccountName
	storageKeyCacheMu.RLock()
	key, ok := storageKeyCache[cacheIndex]
	storageKeyCacheMu.RUnlock()

	if ok {
		return key, true, nil
	}

	storageKeyCacheMu.Lock()
	defer storageKeyCacheMu.Unlock()
	key, ok = storageKeyCache[cacheIndex]
	if !ok {
		accountKeys, err := armClient.storageServiceClient.ListKeys(ctx, resourceGroupName, storageAccountName)
		if utils.ResponseWasNotFound(accountKeys.Response) {
			return "", false, nil
		}
		if err != nil {
			// We assume this is a transient error rather than a 404 (which is caught above),  so assume the
			// account still exists.
			return "", true, fmt.Errorf("Error retrieving keys for storage account %q: %s", storageAccountName, err)
		}

		if accountKeys.Keys == nil {
			return "", false, fmt.Errorf("Nil key returned for storage account %q", storageAccountName)
		}

		keys := *accountKeys.Keys
		if len(keys) <= 0 {
			return "", false, fmt.Errorf("No keys returned for storage account %q", storageAccountName)
		}

		keyPtr := keys[0].Value
		if keyPtr == nil {
			return "", false, fmt.Errorf("The first key returned is nil for storage account %q", storageAccountName)
		}

		key = *keyPtr
		storageKeyCache[cacheIndex] = key
	}

	return key, true, nil
}

func (armClient *ArmClient) getBlobStorageClientForStorageAccount(ctx context.Context, resourceGroupName, storageAccountName string) (*mainStorage.BlobStorageClient, bool, error) {
	key, accountExists, err := armClient.getKeyForStorageAccount(ctx, resourceGroupName, storageAccountName)
	if err != nil {
		return nil, accountExists, err
	}
	if accountExists == false {
		return nil, false, nil
	}

	storageClient, err := mainStorage.NewClient(storageAccountName, key, armClient.environment.StorageEndpointSuffix,
		mainStorage.DefaultAPIVersion, true)
	if err != nil {
		return nil, true, fmt.Errorf("Error creating storage client for storage account %q: %s", storageAccountName, err)
	}

	blobClient := storageClient.GetBlobService()
	return &blobClient, true, nil
}

func (c *ArmClient) registerAuthentication(endpoint, graphEndpoint, subscriptionId, tenantId string, auth, graphAuth autorest.Authorizer, sender autorest.Sender) {
	servicePrincipalsClient := graphrbac.NewServicePrincipalsClientWithBaseURI(graphEndpoint, tenantId)
	setUserAgent(&servicePrincipalsClient.Client)
	servicePrincipalsClient.Authorizer = graphAuth
	servicePrincipalsClient.Sender = sender
	servicePrincipalsClient.SkipResourceProviderRegistration = c.skipProviderRegistration
	c.servicePrincipalsClient = servicePrincipalsClient
}
