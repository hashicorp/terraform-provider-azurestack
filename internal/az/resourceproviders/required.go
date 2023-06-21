// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceproviders

// RequiredResourceProviders returns all of the Resource Providers used by the AzureStack Provider
// whilst all may not be used by every user - the intention is that we determine which should be
// registered such that we can avoid obscure errors where Resource Providers aren't registered.
// new Resource Providers should be added to this list as they're used in the Provider
// (this is the approach used by Microsoft in their tooling)
func Required() map[string]struct{} {
	// NOTE: Resource Providers in this list are case sensitive
	return map[string]struct{}{
		"Microsoft.Authorization": {},
		"Microsoft.Compute":       {},
		"Microsoft.KeyVault":      {},
		"Microsoft.Network":       {},
		"Microsoft.Storage":       {},
	}
}
