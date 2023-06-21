// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccDataSourceLoadBalancer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_lb", "test")
	d := LoadBalancer{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: d.dataSourceBasic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("tags.Environment").HasValue("production"),
				check.That(data.ResourceName).Key("tags.Purpose").HasValue("AcceptanceTests"),
			),
		},
	})
}

func (r LoadBalancer) dataSourceBasic(data acceptance.TestData) string {
	resource := r.basic(data)
	return fmt.Sprintf(`
%s

data "azurestack_lb" "test" {
  name                = azurestack_lb.test.name
  resource_group_name = azurestack_lb.test.resource_group_name
}
`, resource)
}
