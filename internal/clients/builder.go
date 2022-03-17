package clients

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"
	"github.com/hashicorp/terraform-provider-azurestack/internal/common"
	"github.com/hashicorp/terraform-provider-azurestack/internal/features"
)

type ClientBuilder struct {
	AuthConfig                  *authentication.Config
	DisableCorrelationRequestID bool
	CustomCorrelationRequestID  string
	SkipProviderRegistration    bool
	TerraformVersion            string
	Features                    features.UserFeatures
}

func Build(ctx context.Context, builder ClientBuilder) (*Client, error) {
	env, err := authentication.LoadEnvironmentFromUrl(builder.AuthConfig.CustomResourceManagerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to load stack encironment from endpoint %q: %+v", builder.AuthConfig.CustomResourceManagerEndpoint, err)
	}

	// client declarations:
	account, err := NewResourceManagerAccount(ctx, *builder.AuthConfig, *env, builder.SkipProviderRegistration)
	if err != nil {
		return nil, fmt.Errorf("building account: %+v", err)
	}

	client := Client{
		Account: account,
	}

	oauthConfig, err := builder.AuthConfig.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, fmt.Errorf("building OAuth Config: %+v", err)
	}

	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, fmt.Errorf("unable to configure OAuthConfig for tenant %s", builder.AuthConfig.TenantID)
	}

	sender := sender.BuildSender("Azurestack")

	// Resource Manager endpoints
	endpoint := env.ResourceManagerEndpoint
	auth, err := builder.AuthConfig.GetADALToken(ctx, sender, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization token for resource manager: %+v", err)
	}

	// Graph Endpoints
	graphEndpoint := env.GraphEndpoint
	graphAuth, err := builder.AuthConfig.GetADALToken(ctx, sender, oauthConfig, graphEndpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization token for graph endpoints: %+v", err)
	}

	// Storage Endpoints
	storageAuth, err := builder.AuthConfig.GetADALToken(ctx, sender, oauthConfig, endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization token for storage endpoints: %+v", err)
	}

	// Key Vault Endpoints
	keyVaultAuth := builder.AuthConfig.ADALBearerAuthorizerCallback(ctx, sender, oauthConfig)

	o := &common.ClientOptions{
		SubscriptionId:              builder.AuthConfig.SubscriptionID,
		TenantID:                    builder.AuthConfig.TenantID,
		TerraformVersion:            builder.TerraformVersion,
		GraphAuthorizer:             graphAuth,
		GraphEndpoint:               graphEndpoint,
		KeyVaultAuthorizer:          keyVaultAuth,
		ResourceManagerAuthorizer:   auth,
		ResourceManagerEndpoint:     endpoint,
		StorageAuthorizer:           storageAuth,
		SkipProviderReg:             builder.SkipProviderRegistration,
		DisableCorrelationRequestID: builder.DisableCorrelationRequestID,
		CustomCorrelationRequestID:  builder.CustomCorrelationRequestID,
		Environment:                 *env,
		TokenFunc: func(endpoint string) (autorest.Authorizer, error) {
			authorizer, err := builder.AuthConfig.GetADALToken(ctx, sender, oauthConfig, endpoint)
			if err != nil {
				return nil, fmt.Errorf("getting authorization token for endpoint %s: %+v", endpoint, err)
			}
			return authorizer, nil
		},
	}

	if err := client.Build(ctx, o); err != nil {
		return nil, fmt.Errorf("building Client: %+v", err)
	}

	/*if features.EnhancedValidationEnabled() {
		location.CacheSupportedLocations(ctx, env.ResourceManagerEndpoint)
		resourceproviders.CacheSupportedProviders(ctx, client.Resource.ProvidersClient)
	}*/

	return &client, nil
}
