// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dns_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type azurestackDNSZoneDataSource struct{}

func TestAccDNSZoneDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_dns_zone", "test")
	r := azurestackDNSZoneDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func TestAccDNSZoneDataSource_tags(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_dns_zone", "test")
	r := azurestackDNSZoneDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.tags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.hello").HasValue("world"),
			),
		},
	})
}

func TestAccDNSZoneDataSource_withoutResourceGroupName(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_dns_zone", "test")
	r := azurestackDNSZoneDataSource{}
	// resource group of DNS zone is always small case
	resourceGroupName := fmt.Sprintf("acctestrg-%d", data.RandomInteger)

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.onlyName(data, resourceGroupName),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("resource_group_name").HasValue(resourceGroupName),
			),
		},
	})
}

func (azurestackDNSZoneDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name
}

data "azurestack_dns_zone" "test" {
  name                = azurestack_dns_zone.test.name
  resource_group_name = azurestack_dns_zone.test.resource_group_name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (azurestackDNSZoneDataSource) tags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name

  tags = {
    hello = "world"
  }
}

data "azurestack_dns_zone" "test" {
  name                = azurestack_dns_zone.test.name
  resource_group_name = azurestack_dns_zone.test.resource_group_name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (azurestackDNSZoneDataSource) onlyName(data acceptance.TestData, resourceGroupName string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "%s"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name
}

data "azurestack_dns_zone" "test" {
  name = azurestack_dns_zone.test.name
}
`, resourceGroupName, data.Locations.Primary, data.RandomInteger)
}
