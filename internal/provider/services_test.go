// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/sdk"
)

func TestTypedDataSourcesContainValidModelObjects(t *testing.T) {
	for _, service := range SupportedTypedServices() {
		t.Logf("Service %q..", service.Name())
		for _, resource := range service.DataSources() {
			t.Logf("- DataSources %q..", resource.ResourceType())
			obj := resource.ModelObject()
			if err := sdk.ValidateModelObject(obj); err != nil {
				t.Fatalf("validating model: %+v", err)
			}
		}
	}
}

func TestTypedResourcesContainValidModelObjects(t *testing.T) {
	for _, service := range SupportedTypedServices() {
		t.Logf("Service %q..", service.Name())
		for _, resource := range service.Resources() {
			t.Logf("- Resource %q..", resource.ResourceType())
			obj := resource.ModelObject()
			if err := sdk.ValidateModelObject(obj); err != nil {
				t.Fatalf("validating model: %+v", err)
			}
		}
	}
}
