package azurestack

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/helpers/authentication"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	var p *schema.Provider
	p = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"arm_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENDPOINT", ""),
			},
			"subscription_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
			},

			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
			},

			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
			},

			"skip_credentials_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_CREDENTIALS_VALIDATION", false),
			},

			"skip_provider_registration": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_PROVIDER_REGISTRATION", false),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"azurestack_client_config":          dataSourceArmClientConfig(),
			"azurestack_network_interface":      dataSourceArmNetworkInterface(),
			"azurestack_network_security_group": dataSourceArmNetworkSecurityGroup(),
			"azurestack_resource_group":         dataSourceArmResourceGroup(),
			"azurestack_storage_account":        dataSourceArmStorageAccount(),
			"azurestack_virtual_network":        dataSourceArmVirtualNetwork(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"azurestack_availability_set":          resourceArmAvailabilitySet(),
			"azurestack_dns_zone":                  resourceArmDnsZone(),
			"azurestack_dns_a_record":              resourceArmDnsARecord(),
			"azurestack_network_interface":         resourceArmNetworkInterface(),
			"azurestack_network_security_group":    resourceArmNetworkSecurityGroup(),
			"azurestack_network_security_rule":     resourceArmNetworkSecurityRule(),
			"azurestack_local_network_gateway":     resourceArmLocalNetworkGateway(),
			"azurestack_lb":                        resourceArmLoadBalancer(),
			"azurestack_lb_backend_address_pool":   resourceArmLoadBalancerBackendAddressPool(),
			"azurestack_lb_nat_rule":               resourceArmLoadBalancerNatRule(),
			"azurestack_lb_probe":                  resourceArmLoadBalancerProbe(),
			"azurestack_lb_nat_pool":               resourceArmLoadBalancerNatPool(),
			"azurestack_public_ip":                 resourceArmPublicIp(),
			"azurestack_resource_group":            resourceArmResourceGroup(),
			"azurestack_route":                     resourceArmRoute(),
			"azurestack_route_table":               resourceArmRouteTable(),
			"azurestack_storage_account":           resourceArmStorageAccount(),
			"azurestack_storage_blob":              resourceArmStorageBlob(),
			"azurestack_storage_container":         resourceArmStorageContainer(),
			"azurestack_subnet":                    resourceArmSubnet(),
			"azurestack_virtual_network":           resourceArmVirtualNetwork(),
			"azurestack_virtual_machine":           resourceArmVirtualMachine(),
			"azurestack_virtual_machine_extension": resourceArmVirtualMachineExtensions(),
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		config := &authentication.Config{
			SubscriptionID:            d.Get("subscription_id").(string),
			ClientID:                  d.Get("client_id").(string),
			ClientSecret:              d.Get("client_secret").(string),
			TenantID:                  d.Get("tenant_id").(string),
			Environment:               "AZURESTACKCLOUD",
			SkipCredentialsValidation: d.Get("skip_credentials_validation").(bool),
			SkipProviderRegistration:  d.Get("skip_provider_registration").(bool),
			ARMEndpoint:               d.Get("arm_endpoint").(string),
		}

		if config.ARMEndpoint == "" {
			return nil, fmt.Errorf("The Azure Resource Manager endpoint must be specified either" +
				" via `arm_endpoint` in the Provider Block or the `ARM_ENDPOINT` Environment Variable.")
		}

		log.Printf("[DEBUG] Using Service Principal for Authentication")
		if err := config.ValidateServicePrincipal(); err != nil {
			return nil, err
		}

		client, err := getArmClient(config)
		if err != nil {
			return nil, err
		}

		client.StopContext = p.StopContext()

		// replaces the context between tests
		p.MetaReset = func() error {
			client.StopContext = p.StopContext()
			return nil
		}

		if !config.SkipCredentialsValidation {
			// List all the available providers and their registration state to avoid unnecessary
			// requests. This also lets us check if the provider credentials are correct.
			ctx := client.StopContext
			providerList, err := client.providersClient.List(ctx, nil, "")
			if err != nil {
				return nil, fmt.Errorf("Unable to list provider registration status, it is possible that this is due to invalid "+
					"credentials or the service principal does not have permission to use the Resource Manager API, Azure "+
					"error: %s", err)
			}

			if !config.SkipProviderRegistration {
				err = registerAzureResourceProvidersWithSubscription(ctx, providerList.Values(), client.providersClient)
				if err != nil {
					return nil, err
				}
			}
		}

		return client, nil
	}
}

func registerProviderWithSubscription(ctx context.Context, providerName string, client resources.ProvidersClient) error {
	_, err := client.Register(ctx, providerName)
	if err != nil {
		return fmt.Errorf("Cannot register provider %s with Azure Resource Manager: %s.", providerName, err)
	}

	return nil
}

func determineAzureResourceProvidersToRegister(providerList []resources.Provider) map[string]struct{} {
	providers := map[string]struct{}{
		"Microsoft.Authorization": {},
		// "Microsoft.Automation":          {},
		// "Microsoft.Cache": {},
		// "Microsoft.Cdn":                 {},
		"Microsoft.Compute": {},
		// "Microsoft.ContainerInstance": {},
		// "Microsoft.ContainerRegistry":   {},
		// "Microsoft.ContainerService":    {},
		// "Microsoft.DBforMySQL":          {},
		// "Microsoft.DBforPostgreSQL":     {},
		// "Microsoft.DocumentDB":          {},
		// "Microsoft.EventGrid":           {},
		// "Microsoft.EventHub":            {},
		"Microsoft.KeyVault": {},
		// "microsoft.insights":            {},
		"Microsoft.Network": {},
		// "Microsoft.OperationalInsights": {},
		// "Microsoft.Resources":           {},
		// "Microsoft.Search":              {},
		// "Microsoft.ServiceBus":          {},
		// "Microsoft.Sql":                 {},
		"Microsoft.Storage": {},
	}

	// filter out any providers already registered
	for _, p := range providerList {
		if _, ok := providers[*p.Namespace]; !ok {
			continue
		}

		if strings.ToLower(*p.RegistrationState) == "registered" {
			log.Printf("[DEBUG] Skipping provider registration for namespace %s\n", *p.Namespace)
			delete(providers, *p.Namespace)
		}
	}

	return providers
}

// registerAzureResourceProvidersWithSubscription uses the providers client to register
// all Azure resource providers which the Terraform provider may require (regardless of
// whether they are actually used by the configuration or not). It was confirmed by Microsoft
// that this is the approach their own internal tools also take.
func registerAzureResourceProvidersWithSubscription(ctx context.Context, providerList []resources.Provider, client resources.ProvidersClient) error {
	providers := determineAzureResourceProvidersToRegister(providerList)

	var err error
	var wg sync.WaitGroup
	wg.Add(len(providers))

	for providerName := range providers {
		go func(p string) {
			defer wg.Done()
			log.Printf("[DEBUG] Registering provider with namespace %s\n", p)
			if innerErr := registerProviderWithSubscription(ctx, p, client); err != nil {
				err = innerErr
			}
		}(providerName)
	}

	wg.Wait()

	return err
}

// armMutexKV is the instance of MutexKV for ARM resources
var armMutexKV = mutexkv.NewMutexKV()

// Resource group names can be capitalised, but we store them in lowercase.
// Use a custom diff function to avoid creation of new resources.
func resourceAzureStackResourceGroupNameDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return strings.ToLower(old) == strings.ToLower(new)
}

// ignoreCaseDiffSuppressFunc is a DiffSuppressFunc from helper/schema that is
// used to ignore any case-changes in a return value.
func ignoreCaseDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return strings.ToLower(old) == strings.ToLower(new)
}

// ignoreCaseStateFunc is a StateFunc from helper/schema that converts the
// supplied value to lower before saving to state for consistency.
func ignoreCaseStateFunc(val interface{}) string {
	return strings.ToLower(val.(string))
}

func userDataStateFunc(v interface{}) string {
	switch s := v.(type) {
	case string:
		s = base64Encode(s)
		hash := sha1.Sum([]byte(s))
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

// base64Encode encodes data if the input isn't already encoded using
// base64.StdEncoding.EncodeToString. If the input is already base64 encoded,
// return the original input unchanged.
func base64Encode(data string) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return data
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func isBase64Encoded(data string) bool {
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}
