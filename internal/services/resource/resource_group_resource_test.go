// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type ResourceGroupResource struct{}

func TestAccResourceGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_resource_group", "test")
	testResource := ResourceGroupResource{}
	data.ResourceTest(t, testResource, []acceptance.TestStep{
		data.ApplyStep(testResource.basicConfig, testResource),
		data.ImportStep(),
	})
}

func TestAccResourceGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_resource_group", "test")
	testResource := ResourceGroupResource{}
	data.ResourceTest(t, testResource, []acceptance.TestStep{
		data.ApplyStep(testResource.basicConfig, testResource),
		data.RequiresImportErrorStep(testResource.requiresImportConfig),
	})
}

func TestAccResourceGroup_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_resource_group", "test")
	testResource := ResourceGroupResource{}
	data.ResourceTest(t, testResource, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       testResource.basicConfig,
			TestResource: testResource,
		}),
	})
}

func TestAccResourceGroup_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_resource_group", "test")
	testResource := ResourceGroupResource{}
	assert := check.That(data.ResourceName)
	data.ResourceTest(t, testResource, []acceptance.TestStep{
		{
			Config: testResource.withTagsConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				assert.ExistsInAzure(testResource),
				assert.Key("tags.%").HasValue("2"),
				assert.Key("tags.cost_center").HasValue("MSFT"),
				assert.Key("tags.environment").HasValue("Production"),
			),
		},
		data.ImportStep(),
		{
			Config: testResource.withTagsUpdatedConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				assert.ExistsInAzure(testResource),
				assert.Key("tags.%").HasValue("1"),
				assert.Key("tags.environment").HasValue("staging"),
			),
		},
		data.ImportStep(),
	})
}

/*
// todo put back in when we add vnets back in
func TestAccResourceGroup_withNestedItemsAndFeatureFlag(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_resource_group", "test")
	r := ResourceGroupResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withFeatureFlag(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				// since we don't want to track/destroy this resource for test purposes, we can create this here
				// it'll be cleaned up in the final step with the feature flag disabled, so this should be fine.
				data.CheckWithClient(r.createNetworkOutsideTerraform(fmt.Sprintf("acctestvnet-%d", data.RandomInteger))),
			),
		},
		data.ImportStep(),
		{
			// attempting to delete this with the vnet should error
			Config:      r.withFeatureFlag(data, true),
			Destroy:     true,
			ExpectError: regexp.MustCompile("This feature is intended to avoid the unintentional destruction"),
		},
		{
			// with the feature disabled we should delete the RG and the Network
			Config:  r.withFeatureFlag(data, false),
			Destroy: true,
		},
	})
}*/

func (t ResourceGroupResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	resourceGroup := state.Attributes["name"]

	groupsClient := client.Resource.GroupsClient
	deleteFuture, err := groupsClient.Delete(ctx, resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("deleting Resource Group %q: %+v", resourceGroup, err)
	}

	err = deleteFuture.WaitForCompletionRef(ctx, groupsClient.Client)
	if err != nil {
		return nil, fmt.Errorf("waiting for deletion of Resource Group %q: %+v", resourceGroup, err)
	}

	return pointer.FromBool(true), nil
}

func (t ResourceGroupResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	name := state.Attributes["name"]

	resp, err := client.Resource.GroupsClient.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("retrieving Resource Group %q: %+v", name, err)
	}

	return pointer.FromBool(resp.Properties != nil), nil
}

func (t ResourceGroupResource) basicConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (t ResourceGroupResource) requiresImportConfig(data acceptance.TestData) string {
	template := t.basicConfig(data)
	return fmt.Sprintf(`
%s

resource "azurestack_resource_group" "import" {
  name     = azurestack_resource_group.test.name
  location = azurestack_resource_group.test.location
}
`, template)
}

func (t ResourceGroupResource) withTagsConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (t ResourceGroupResource) withTagsUpdatedConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}
