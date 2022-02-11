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

type LoadBalancerProbe struct{}

func TestAccLoadBalancerProbe_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_probe", "test")
	r := LoadBalancerProbe{}

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

func TestAccLoadBalancerProbe_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_probe", "test")
	r := LoadBalancerProbe{}

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

func TestAccLoadBalancerProbe_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_probe", "test")
	r := LoadBalancerProbe{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccLoadBalancerProbe_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_probe", "test")
	data2 := acceptance.BuildTestData(t, "azurestack_lb_probe", "test2")
	r := LoadBalancerProbe{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multipleProbes(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("port").HasValue("80"),
			),
		},
		data.ImportStep(),
		data2.ImportStep(),
		{
			Config: r.multipleProbesUpdate(data, data2),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).ExistsInAzure(r),
				check.That(data2.ResourceName).Key("port").HasValue("8080"),
			),
		},
	})
}

func TestAccLoadBalancerProbe_updateProtocol(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb_probe", "test")
	r := LoadBalancerProbe{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.updateProtocolBefore(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("protocol").HasValue("Http"),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateProtocolAfter(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("protocol").HasValue("Tcp"),
			),
		},
	})
}

func (r LoadBalancerProbe) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancerProbeID(state.ID)
	if err != nil {
		return nil, err
	}

	lb, err := client.LoadBalancer.LoadBalancersClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		if utils.ResponseWasNotFound(lb.Response) {
			return nil, fmt.Errorf("Load Balancer %q (resource group %q) not found for Probe %q", id.LoadBalancerName, id.ResourceGroup, id.ProbeName)
		}
		return nil, fmt.Errorf("failed reading Load Balancer %q (resource group %q) for Probe %q", id.LoadBalancerName, id.ResourceGroup, id.ProbeName)
	}
	props := lb.LoadBalancerPropertiesFormat
	if props == nil || props.Probes == nil || len(*props.Probes) == 0 {
		return nil, fmt.Errorf("Probe %q not found in Load Balancer %q (resource group %q)", id.ProbeName, id.LoadBalancerName, id.ResourceGroup)
	}

	found := false
	for _, v := range *props.Probes {
		if v.Name != nil && *v.Name == id.ProbeName {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Probe %q not found in Load Balancer %q (resource group %q)", id.ProbeName, id.LoadBalancerName, id.ResourceGroup)
	}
	return pointer.FromBool(found), nil
}

func (r LoadBalancerProbe) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.LoadBalancerProbeID(state.ID)
	if err != nil {
		return nil, err
	}

	lb, err := client.LoadBalancer.LoadBalancersClient.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Load Balancer %q (Resource Group %q)", id.LoadBalancerName, id.ResourceGroup)
	}
	if lb.LoadBalancerPropertiesFormat == nil {
		return nil, fmt.Errorf("`properties` was nil")
	}
	if lb.LoadBalancerPropertiesFormat.Probes == nil {
		return nil, fmt.Errorf("`properties.Probes` was nil")
	}

	probes := make([]network.Probe, 0)
	for _, probe := range *lb.LoadBalancerPropertiesFormat.Probes {
		if probe.Name == nil || *probe.Name == id.ProbeName {
			continue
		}

		probes = append(probes, probe)
	}
	lb.LoadBalancerPropertiesFormat.Probes = &probes

	future, err := client.LoadBalancer.LoadBalancersClient.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, lb)
	if err != nil {
		return nil, fmt.Errorf("updating Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.LoadBalancer.LoadBalancersClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for update of Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	return pointer.FromBool(true), nil
}

func (r LoadBalancerProbe) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}

resource "azurestack_lb_probe" "test" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  port                = 22
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r LoadBalancerProbe) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_lb_probe" "import" {
  name                = azurestack_lb_probe.test.name
  loadbalancer_id     = azurestack_lb_probe.test.loadbalancer_id
  resource_group_name = azurestack_lb_probe.test.resource_group_name
  port                = 22
}
`, template)
}

func (r LoadBalancerProbe) multipleProbes(data, data2 acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}

resource "azurestack_lb_probe" "test" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  port                = 22
}

resource "azurestack_lb_probe" "test2" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  port                = 80
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data2.RandomInteger)
}

func (r LoadBalancerProbe) multipleProbesUpdate(data, data2 acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}

resource "azurestack_lb_probe" "test" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  port                = 22
}

resource "azurestack_lb_probe" "test2" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  port                = 8080
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data2.RandomInteger)
}

func (r LoadBalancerProbe) updateProtocolBefore(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}

resource "azurestack_lb_probe" "test" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  protocol            = "Http"
  request_path        = "/"
  port                = 80
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r LoadBalancerProbe) updateProtocolAfter(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}

resource "azurestack_lb_probe" "test" {
  resource_group_name = azurestack_resource_group.test.name
  loadbalancer_id     = azurestack_lb.test.id
  name                = "probe-%d"
  protocol            = "Tcp"
  port                = 80
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
