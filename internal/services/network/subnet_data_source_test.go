// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type SubnetDataSource struct{}

func TestAccSubnetDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_subnet", "test")
	r := SubnetDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
				check.That(data.ResourceName).Key("resource_group_name").Exists(),
				check.That(data.ResourceName).Key("virtual_network_name").Exists(),
				check.That(data.ResourceName).Key("address_prefix").Exists(),
				check.That(data.ResourceName).Key("network_security_group_id").HasValue(""),
				check.That(data.ResourceName).Key("route_table_id").HasValue(""),
			),
		},
	})
}

func (r SubnetDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.0.0/24"
}

data "azurestack_subnet" "test" {
  name                 = azurestack_subnet.test.name
  virtual_network_name = azurestack_subnet.test.virtual_network_name
  resource_group_name  = azurestack_subnet.test.resource_group_name
}
`, r.template(data))
}

func (SubnetDataSource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest%d-vn"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  address_space       = ["10.0.0.0/16"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
