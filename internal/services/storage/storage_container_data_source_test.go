// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storage_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

type StorageContainerDataSource struct{}

func TestAccStorageContainerDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_storage_container", "test")

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: StorageContainerDataSource{}.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("container_access_type").HasValue("private"),
				check.That(data.ResourceName).Key("has_immutability_policy").HasValue("false"),
				check.That(data.ResourceName).Key("metadata.%").HasValue("2"),
				check.That(data.ResourceName).Key("metadata.k1").HasValue("v1"),
				check.That(data.ResourceName).Key("metadata.k2").HasValue("v2"),
			),
		},
	})
}

func (d StorageContainerDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%s"
  location = "%s"
}

resource "azurestack_storage_account" "test" {
  name                     = "acctestsadsc%s"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "containerdstest-%s"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
  metadata = {
    k1 = "v1"
    k2 = "v2"
  }
}

data "azurestack_storage_container" "test" {
  name                 = azurestack_storage_container.test.name
  storage_account_name = azurestack_storage_container.test.storage_account_name
}
`, data.RandomString, data.Locations.Primary, data.RandomString, data.RandomString)
}
