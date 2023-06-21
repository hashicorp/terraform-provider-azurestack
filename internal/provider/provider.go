// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceproviders"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/sdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func AzureProvider() *schema.Provider {
	return azureProvider(false)
}

func TestAzureProvider() *schema.Provider {
	return azureProvider(true)
}

func azureProvider(supportLegacyTestSuite bool) *schema.Provider {
	// avoids this showing up in test output
	debugLog := func(f string, v ...interface{}) {
		if os.Getenv("TF_LOG") == "" {
			return
		}

		if os.Getenv("TF_ACC") != "" {
			return
		}

		log.Printf(f, v...)
	}

	dataSources := make(map[string]*schema.Resource)
	resources := make(map[string]*schema.Resource)

	// first handle the typed services
	for _, service := range SupportedTypedServices() {
		debugLog("[DEBUG] Registering Data Sources for %q..", service.Name())
		for _, ds := range service.DataSources() {
			key := ds.ResourceType()
			if existing := dataSources[key]; existing != nil {
				panic(fmt.Sprintf("An existing Data Source exists for %q", key))
			}

			wrapper := sdk.NewDataSourceWrapper(ds)
			dataSource, err := wrapper.DataSource()
			if err != nil {
				panic(fmt.Errorf("creating Wrapper for Data Source %q: %+v", key, err))
			}

			dataSources[key] = dataSource
		}

		debugLog("[DEBUG] Registering Resources for %q..", service.Name())
		for _, r := range service.Resources() {
			key := r.ResourceType()
			if existing := resources[key]; existing != nil {
				panic(fmt.Sprintf("An existing Resource exists for %q", key))
			}

			wrapper := sdk.NewResourceWrapper(r)
			resource, err := wrapper.Resource()
			if err != nil {
				panic(fmt.Errorf("creating Wrapper for Resource %q: %+v", key, err))
			}
			resources[key] = resource
		}
	}

	// then handle the untyped services
	for _, service := range SupportedUntypedServices() {
		debugLog("[DEBUG] Registering Data Sources for %q..", service.Name())
		for k, v := range service.SupportedDataSources() {
			if existing := dataSources[k]; existing != nil {
				panic(fmt.Sprintf("An existing Data Source exists for %q", k))
			}

			dataSources[k] = v
		}

		debugLog("[DEBUG] Registering Resources for %q..", service.Name())
		for k, v := range service.SupportedResources() {
			if existing := resources[k]; existing != nil {
				panic(fmt.Sprintf("An existing Resource exists for %q", k))
			}

			resources[k] = v
		}
	}

	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"subscription_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
				Description: "The Subscription ID which should be used.",
			},

			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
				Description: "The Client ID which should be used.",
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
				Description: "The Tenant ID which should be used.",
			},

			"metadata_host": {
				Type:        schema.TypeString,
				Optional:    true, // TODO: make Required when `arm_endpoint` is removed
				DefaultFunc: schema.EnvDefaultFunc("ARM_METADATA_HOST", ""),
				Description: "The Hostname which should be used for the Azure Metadata Service.",
			},

			"arm_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENDPOINT", ""),
				Description: "The Hostname which should be used for the Azure Metadata Service.",
				Deprecated:  "`arm_endpoint` is deprecated in favour of `metadata_host` and will be removed in version 1.0 of the AzureStack provider.",
			},

			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", ""),
				Description: "The Cloud Environment which should be used.",
			},

			"auxiliary_tenant_ids": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// Client Certificate specific fields
			"client_certificate_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PATH", ""),
				Description: "The path to the Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate.",
			},

			"client_certificate_password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PASSWORD", ""),
				Description: "The password associated with the Client Certificate. For use when authenticating as a Service Principal using a Client Certificate",
			},

			// Client Secret specific fields
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
				Description: "The Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret.",
			},

			// Managed Service Identity specific fields
			"use_msi": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", false),
				Description: "Allowed Managed Service Identity be used for Authentication.",
			},

			"msi_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_MSI_ENDPOINT", ""),
				Description: "The path to a custom endpoint for Managed Service Identity - in most circumstances this should be detected automatically. ",
			},

			"disable_correlation_request_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_DISABLE_CORRELATION_REQUEST_ID", false),
				Description: "This will disable the x-ms-correlation-request-id header.",
			},

			// Advanced feature flags
			"skip_provider_registration": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_PROVIDER_REGISTRATION", false),
				Description: "Should the AzureStack Provider skip registering all of the Resource Providers that it supports, if they're not already registered?",
			},

			"features": schemaFeatures(supportLegacyTestSuite),
		},

		DataSourcesMap: dataSources,
		ResourcesMap:   resources,
	}

	p.ConfigureContextFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var auxTenants []string
		if v, ok := d.Get("auxiliary_tenant_ids").([]interface{}); ok && len(v) > 0 {
			auxTenants = *utils.ExpandStringSlice(v)
		} else if v := os.Getenv("ARM_AUXILIARY_TENANT_IDS"); v != "" {
			auxTenants = strings.Split(v, ";")
		}

		if len(auxTenants) > 3 {
			return nil, diag.FromErr(fmt.Errorf("The provider only supports 3 auxiliary tenant IDs"))
		}

		var metadataHost string
		if v := d.Get("metadata_host").(string); v != "" {
			metadataHost = v
		} else if v := d.Get("arm_endpoint").(string); v != "" {
			metadataHost = v
		}

		u, err := url.Parse(metadataHost)
		if err != nil {
			return nil, diag.Errorf("parsing `metadata_host`: %v", err)
		}
		if u.Host != "" {
			metadataHost = u.Host
		}

		if len(metadataHost) == 0 {
			return nil, diag.Errorf("provider: `metadata_host` must be set")
		}

		builder := &authentication.Builder{
			SubscriptionID:     d.Get("subscription_id").(string),
			ClientID:           d.Get("client_id").(string),
			ClientSecret:       d.Get("client_secret").(string),
			TenantID:           d.Get("tenant_id").(string),
			Environment:        d.Get("environment").(string),
			MetadataHost:       metadataHost,
			AuxiliaryTenantIDs: auxTenants,
			MsiEndpoint:        d.Get("msi_endpoint").(string),
			ClientCertPassword: d.Get("client_certificate_password").(string),
			ClientCertPath:     d.Get("client_certificate_path").(string),

			// Feature Toggles
			SupportsClientCertAuth:   true,
			SupportsClientSecretAuth: true,
			// SupportsManagedServiceIdentity: d.Get("use_msi").(bool), todo supported in stack?
			SupportsAzureCliToken:    true,
			SupportsAuxiliaryTenants: len(auxTenants) > 0,

			// Doc Links
			ClientSecretDocsLink: "https://registry.terraform.io/providers/hashicorp/azurestack/latest/docs/guides/service_principal_client_secret",
		}

		config, err := builder.Build()
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("building Azurestack Client: %s", err))
		}

		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}

		skipProviderRegistration := d.Get("skip_provider_registration").(bool)
		clientBuilder := clients.ClientBuilder{
			AuthConfig:                  config,
			SkipProviderRegistration:    skipProviderRegistration,
			TerraformVersion:            terraformVersion,
			DisableCorrelationRequestID: d.Get("disable_correlation_request_id").(bool),

			// this field is intentionally not exposed in the provider block, since it's only used for
			// platform level tracing
			CustomCorrelationRequestID: os.Getenv("ARM_CORRELATION_REQUEST_ID"),
		}

		//lint:ignore SA1019 SDKv2 migration - staticcheck's own linter directives are currently being ignored under golanci-lint
		stopCtx, ok := schema.StopContext(ctx) //nolint:staticcheck
		if !ok {
			stopCtx = ctx
		}

		client, err := clients.Build(stopCtx, clientBuilder)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		client.StopContext = stopCtx

		if !skipProviderRegistration {
			// List all the available providers and their registration state to avoid unnecessary
			// requests. This also lets us check if the provider credentials are correct.
			providerList, err := client.Resource.ProvidersClient.List(ctx, nil, "")
			if err != nil {
				return nil, diag.FromErr(fmt.Errorf("Unable to list provider registration status, it is possible that this is due to invalid "+
					"credentials or the service principal does not have permission to use the Resource Manager API, Azure "+
					"error: %s", err))
			}

			availableResourceProviders := providerList.Values()
			requiredResourceProviders := resourceproviders.Required()

			if err := resourceproviders.EnsureRegistered(ctx, *client.Resource.ProvidersClient, availableResourceProviders, requiredResourceProviders); err != nil {
				return nil, diag.FromErr(fmt.Errorf(resourceProviderRegistrationErrorFmt, err))
			}
		}

		return client, nil
	}
}

const resourceProviderRegistrationErrorFmt = `Error ensuring Resource Providers are registered.

Terraform automatically attempts to register the Resource Providers it supports to
ensure it's able to provision resources.

If you don't have permission to register Resource Providers you may wish to use the
"skip_provider_registration" flag in the Provider block to disable this functionality.

Please note that if you opt out of Resource Provider Registration and Terraform tries
to provision a resource from a Resource Provider which is unregistered, then the errors
may appear misleading - for example:

> API version 2019-XX-XX was not found for Microsoft.Foo

Could indicate either that the Resource Provider "Microsoft.Foo" requires registration,
but this could also indicate that this Azure Region doesn't support this API version.

More information on the "skip_provider_registration" flag can be found here:
https://registry.terraform.io/providers/hashicorp/azurestack/latest/docs#skip_provider_registration

Original Error: %s`
