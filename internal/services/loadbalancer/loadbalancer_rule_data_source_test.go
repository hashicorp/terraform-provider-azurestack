// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
)

func TestAccLoadBalancerRuleDataSource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.basicDataSource(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
				check.That(data.ResourceName).Key("frontend_ip_configuration_name").Exists(),
				check.That(data.ResourceName).Key("protocol").Exists(),
				check.That(data.ResourceName).Key("frontend_port").Exists(),
				check.That(data.ResourceName).Key("backend_port").Exists(),
			),
		},
	})
}

func TestAccLoadBalancerRuleDataSource_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.completeDataSource(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("id").Exists(),
				check.That(data.ResourceName).Key("frontend_ip_configuration_name").Exists(),
				check.That(data.ResourceName).Key("protocol").Exists(),
				check.That(data.ResourceName).Key("frontend_port").Exists(),
				check.That(data.ResourceName).Key("backend_port").Exists(),
				check.That(data.ResourceName).Key("backend_address_pool_id").Exists(),
				check.That(data.ResourceName).Key("probe_id").Exists(),
				check.That(data.ResourceName).Key("enable_floating_ip").Exists(),
				check.That(data.ResourceName).Key("disable_outbound_snat").Exists(),
				check.That(data.ResourceName).Key("idle_timeout_in_minutes").Exists(),
				check.That(data.ResourceName).Key("load_distribution").Exists(),
			),
		},
	})
}

func (r LoadBalancerRule) basicDataSource(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

data "azurestack_lb_rule" "test" {
  name                = azurestack_lb_rule.test.name
  resource_group_name = azurestack_lb_rule.test.resource_group_name
  loadbalancer_id     = azurestack_lb_rule.test.loadbalancer_id
}
`, template)
}

func (r LoadBalancerRule) completeDataSource(data acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s
resource "azurestack_lb_backend_address_pool" "test" {
  name                = "LbPool-%s"
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
}

resource "azurestack_lb_probe" "test" {
  name                = "LbProbe-%s"
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  protocol            = "Tcp"
  port                = 443
}

resource "azurestack_lb_rule" "test" {
  name                = "LbRule-%s"
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id

  protocol      = "Tcp"
  frontend_port = 3389
  backend_port  = 3389

  enable_floating_ip      = true
  idle_timeout_in_minutes = 10

  backend_address_pool_id = azurestack_lb_backend_address_pool.test.id
  probe_id                = azurestack_lb_probe.test.id

  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}

data "azurestack_lb_rule" "test" {
  name                = azurestack_lb_rule.test.name
  resource_group_name = azurestack_lb_rule.test.resource_group_name
  loadbalancer_id     = azurestack_lb_rule.test.loadbalancer_id
}
`, template, data.RandomStringOfLength(8), data.RandomStringOfLength(8), data.RandomStringOfLength(8))
}
