package loadbalancer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type LoadBalancerRule struct{}

func TestAccLoadBalancerRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLoadBalancerRule_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLoadBalancerRule_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLoadBalancerRule_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccLoadBalancerRule_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

// https://github.com/hashicorp/terraform/issues/9424
func TestAccLoadBalancerRule_inconsistentReads(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	r := LoadBalancerRule{}
	p := LoadBalancerProbe{}
	b := LoadBalancerBackendAddressPool{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.inconsistentRead(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That("azurestack_lb_probe.test").ExistsInAzure(p),
				check.That("azurestack_lb_backend_address_pool.test").ExistsInAzure(b),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLoadBalancerRule_updateMultipleRules(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	data2 := acceptance.BuildTestData(t, "azurestack_lb_rule", "test2")
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multipleRules(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("frontend_port").HasValue("3390"),
				check.That(data2.ResourceName).Key("backend_port").HasValue("3390"),
			),
		},
		data.ImportStep(),
		data2.ImportStep(),
		{
			Config: r.multipleRulesUpdate(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("frontend_port").HasValue("3391"),
				check.That(data2.ResourceName).Key("backend_port").HasValue("3391"),
			),
		},
		data.ImportStep(),
		data2.ImportStep(),
	})
}

func TestAccLoadBalancerRule_vmssBackendPoolUpdateRemoveLBRule(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_rule", "test")
	lbRuleName := fmt.Sprintf("LbRule-%s", data.RandomString)
	r := LoadBalancerRule{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vmssBackendPool(data, lbRuleName, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.vmssBackendPoolUpdate(data, lbRuleName, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.vmssBackendPoolWithoutLBRule(data, "Standard"),
		},
	})
}

func (r LoadBalancerRule) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancingRuleID(state.ID)
	if err != nil {
		return nil, err
	}

	rule, err := client.LoadBalancer.LoadBalancingRulesClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(rule.Response) {
			return pointer.FromBool(false), nil
		}

		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}

	return pointer.FromBool(rule.ID != nil), nil
}

func (r LoadBalancerRule) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancingRuleID(state.ID)
	if err != nil {
		return nil, err
	}

	loadBalancer, err := client.LoadBalancer.LoadBalancersClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}
	if loadBalancer.LoadBalancerPropertiesFormat == nil {
		return nil, fmt.Errorf(`properties was nil`)
	}
	if loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules == nil {
		return nil, fmt.Errorf(`properties.LoadBalancingRules was nil`)
	}
	rules := make([]network.LoadBalancingRule, 0)
	for _, v := range *loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules {
		if v.Name == nil || *v.Name == id.Name {
			continue
		}

		rules = append(rules, v)
	}
	loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules = &rules

	future, err := client.LoadBalancer.LoadBalancersClient.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, loadBalancer)
	if err != nil {
		return nil, fmt.Errorf("updating Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.LoadBalancer.LoadBalancersClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for update of Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	return pointer.FromBool(true), nil
}

func (r LoadBalancerRule) template(data acceptance.TestData, sku string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%[1]d"
  location = "%[2]s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%[1]d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "%[3]s"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%[1]d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%[1]d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, sku)
}

func (r LoadBalancerRule) basic(data acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurestack_lb_rule" "test" {
  name                           = "LbRule-%s"
  resource_group_name            = azurestack_resource_group.test.name
  loadbalancer_id                = azurestack_lb.test.id
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
}
`, template, data.RandomStringOfLength(8))
}

func (r LoadBalancerRule) complete(data acceptance.TestData) string {
	template := r.template(data, "Standard")
	return fmt.Sprintf(`
%s

resource "azurestack_lb_rule" "test" {
  name                = "LbRule-%s"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"

  protocol      = "Tcp"
  frontend_port = 3389
  backend_port  = 3389

  disable_outbound_snat   = true
  enable_floating_ip      = true
  idle_timeout_in_minutes = 10

  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomStringOfLength(8))
}

func (r LoadBalancerRule) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_lb_rule" "import" {
  name                           = azurestack_lb_rule.test.name
  resource_group_name            = azurestack_lb_rule.test.resource_group_name
  loadbalancer_id                = azurestack_lb_rule.test.loadbalancer_id
  frontend_ip_configuration_name = azurestack_lb_rule.test.frontend_ip_configuration_name
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
}
`, template)
}

// https://github.com/hashicorp/terraform/issues/9424
func (r LoadBalancerRule) inconsistentRead(data acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "%d-address-pool"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
}

resource "azurestack_lb_probe" "test" {
  name                = "probe-%d"
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  protocol            = "Tcp"
  port                = 443
}

resource "azurestack_lb_rule" "test" {
  name                           = "LbRule-%s"
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomInteger, data.RandomInteger, data.RandomStringOfLength(8))
}

func (r LoadBalancerRule) multipleRules(data, data2 acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurestack_lb_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "LbRule-%s"
  protocol                       = "Udp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}

resource "azurestack_lb_rule" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "LbRule-%s"
  protocol                       = "Udp"
  frontend_port                  = 3390
  backend_port                   = 3390
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomStringOfLength(8), data2.RandomStringOfLength(8))
}

func (r LoadBalancerRule) multipleRulesUpdate(data, data2 acceptance.TestData) string {
	template := r.template(data, "Basic")
	return fmt.Sprintf(`
%s

resource "azurestack_lb_rule" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "LbRule-%s"
  protocol                       = "Udp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}

resource "azurestack_lb_rule" "test2" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "LbRule-%s"
  protocol                       = "Udp"
  frontend_port                  = 3391
  backend_port                   = 3391
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}
`, template, data.RandomStringOfLength(8), data2.RandomStringOfLength(8))
}

func (r LoadBalancerRule) vmssBackendPoolWithoutLBRule(data acceptance.TestData, sku string) string {
	template := r.template(data, sku)
	return fmt.Sprintf(`
%[1]s

resource "azurestack_lb_backend_address_pool" "test" {
  name                = "acctest-lb-BAP-%[2]d"
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
}

resource "azurestack_virtual_network" "test" {
  name                = "acctest-lb-vnet-%[2]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "acctest-lb-subnet-%[2]d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_linux_virtual_machine_scale_set" "test" {
  name                = "acctest-lb-vmss-%[2]d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Standard_F2"
  instances           = 1
  admin_username      = "adminuser"
  admin_password      = "P@ssword1234!"

  disable_password_authentication = false

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name                                  = "internal"
      primary                               = true
      subnet_id                             = azurestack_subnet.test.id
      load_balancer_backend_address_pool_id = azurestack_lb_backend_address_pool.test.id
    }
  }
}
`, template, data.RandomInteger)
}

func (r LoadBalancerRule) vmssBackendPool(data acceptance.TestData, lbRuleName, sku string) string {
	template := r.vmssBackendPoolWithoutLBRule(data, sku)
	return fmt.Sprintf(`
%s

resource "azurestack_lb_rule" "test" {
  resource_group_name            = azurestack_resource_group.test.name
  loadbalancer_id                = azurestack_lb.test.id
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  backend_address_pool_id        = azurestack_lb_backend_address_pool.test.id
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
}
`, template, lbRuleName)
}

func (r LoadBalancerRule) vmssBackendPoolUpdate(data acceptance.TestData, lbRuleName, sku string) string {
	template := r.vmssBackendPoolWithoutLBRule(data, sku)
	return fmt.Sprintf(`
%s
resource "azurestack_lb_rule" "test" {
  resource_group_name            = azurestack_resource_group.test.name
  loadbalancer_id                = azurestack_lb.test.id
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  backend_address_pool_id        = azurestack_lb_backend_address_pool.test.id
  frontend_ip_configuration_name = azurestack_lb.test.frontend_ip_configuration.0.name
  disable_outbound_snat          = false
}
`, template, lbRuleName)
}
