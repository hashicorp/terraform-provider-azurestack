// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceproviders

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
)

func availableResourceProviders(ctx context.Context, client *resources.ProvidersClient) (*[]string, error) {
	providerNames := make([]string, 0)
	providers, err := client.ListComplete(ctx, nil, "")
	if err != nil {
		return nil, fmt.Errorf("listing Resource Providers: %+v", err)
	}
	for providers.NotDone() {
		provider := providers.Value()
		if provider.Namespace != nil {
			providerNames = append(providerNames, *provider.Namespace)
		}

		if err := providers.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return &providerNames, nil
}
