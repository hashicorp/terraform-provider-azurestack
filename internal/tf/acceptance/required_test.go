// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acceptance

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceproviders"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
)

// since this depends on GetAuthConfig which lives in this package
// unfortunately this has to live in a different package to the other func

func TestAccEnsureRequiredResourceProvidersAreRegistered(t *testing.T) {
	config := GetAuthConfig(t)
	if config == nil {
		return
	}

	builder := clients.ClientBuilder{
		AuthConfig:                  config,
		TerraformVersion:            "0.0.0",
		DisableCorrelationRequestID: true,
		// this test intentionally checks all the RP's are registered - so this is intentional
		SkipProviderRegistration: true,
	}
	armClient, err := clients.Build(context.Background(), builder)
	if err != nil {
		t.Fatalf("Error building ARM Client: %+v", err)
	}

	client := armClient.Resource.ProvidersClient
	ctx := armClient.StopContext
	providerList, err := client.List(ctx, nil, "")
	if err != nil {
		t.Fatalf("Unable to list provider registration status, it is possible that this is due to invalid "+
			"credentials or the service principal does not have permission to use the Resource Manager API, Azure "+
			"error: %s", err)
	}

	availableResourceProviders := providerList.Values()
	requiredResourceProviders := resourceproviders.Required()
	err = resourceproviders.EnsureRegistered(ctx, *client, availableResourceProviders, requiredResourceProviders)
	if err != nil {
		t.Fatalf("Error registering Resource Providers: %+v", err)
	}

	// refresh the list now things have been re-registered
	providerList, err = client.List(ctx, nil, "")
	if err != nil {
		t.Fatalf("Unable to list provider registration status, it is possible that this is due to invalid "+
			"credentials or the service principal does not have permission to use the Resource Manager API, Azure "+
			"error: %s", err)
	}

	stillRequiringRegistration := resourceproviders.DetermineResourceProvidersRequiringRegistration(providerList.Values(), requiredResourceProviders)
	if len(stillRequiringRegistration) > 0 {
		t.Fatalf("'%d' Resource Providers are still Pending Registration: %s", len(stillRequiringRegistration), spew.Sprint(stillRequiringRegistration))
	}
}
