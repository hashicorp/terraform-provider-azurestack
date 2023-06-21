// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type PublicIPsResource struct{}

func TestAccPublicIPsDataSource_namePrefix(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_public_ips", "test")
	r := PublicIPsResource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.prefix(data),
		},
		{
			Config: r.prefixDataSource(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("public_ips.#").HasValue("2"),
				check.That(data.ResourceName).Key("public_ips.0.name").HasValue(fmt.Sprintf("acctestpipa%s-0", data.RandomString)),
			),
		},
	})
}

func TestAccDataSourcePublicIPs_allocationType(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_public_ips", "test")
	r := PublicIPsResource{}

	staticDataSourceName := "data.azurestack_public_ips.static"
	dynamicDataSourceName := "data.azurestack_public_ips.dynamic"

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.allocationType(data),
		},
		{
			Config: r.allocationTypeDataSources(data),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.TestCheckResourceAttr(staticDataSourceName, "public_ips.#", "3"),
				acceptance.TestCheckResourceAttr(staticDataSourceName, "public_ips.0.name", fmt.Sprintf("acctestpips%s-0", data.RandomString)),
				acceptance.TestCheckResourceAttr(dynamicDataSourceName, "public_ips.#", "4"),
				acceptance.TestCheckResourceAttr(dynamicDataSourceName, "public_ips.0.name", fmt.Sprintf("acctestpipd%s-0", data.RandomString)),
			),
		},
	})
}

func (PublicIPsResource) prefix(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  count                   = 2
  name                    = "acctestpipb%s-${count.index}"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  allocation_method       = "Static"
  idle_timeout_in_minutes = 30

  tags = {
    environment = "test"
  }
}

resource "azurestack_public_ip" "test2" {
  count                   = 2
  name                    = "acctestpipa%s-${count.index}"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  allocation_method       = "Static"
  idle_timeout_in_minutes = 30

  tags = {
    environment = "test"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}

func (r PublicIPsResource) prefixDataSource(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_public_ips" "test" {
  resource_group_name = azurestack_resource_group.test.name
  name_prefix         = "acctestpipa"
}
`, r.prefix(data))
}

func (PublicIPsResource) allocationType(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "dynamic" {
  count                   = 4
  name                    = "acctestpipd%s-${count.index}"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  allocation_method       = "Dynamic"
  idle_timeout_in_minutes = 30

  tags = {
    environment = "test"
  }
}

resource "azurestack_public_ip" "static" {
  count                   = 3
  name                    = "acctestpips%s-${count.index}"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  allocation_method       = "Static"
  idle_timeout_in_minutes = 30

  tags = {
    environment = "test"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}

func (r PublicIPsResource) allocationTypeDataSources(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_public_ips" "dynamic" {
  resource_group_name = azurestack_resource_group.test.name
  allocation_type     = "Dynamic"
}

data "azurestack_public_ips" "static" {
  resource_group_name = azurestack_resource_group.test.name
  allocation_type     = "Static"
}
`, r.allocationType(data))
}
