package loadbalancer_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type LoadBalancer struct{}

func TestAccLoadBalancer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb", "test")
	r := LoadBalancer{}

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

func TestAccLoadBalancer_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb", "test")
	r := LoadBalancer{}

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

func TestAccLoadBalancer_frontEndConfig(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb", "test")
	r := LoadBalancer{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.frontEndConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("frontend_ip_configuration.#").HasValue("2"),
			),
		},
		data.ImportStep(),
		{
			Config: r.frontEndConfigRemovalWithIP(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("frontend_ip_configuration.#").HasValue("1"),
			),
		},
		data.ImportStep(),
		{
			Config: r.frontEndConfigRemoval(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("frontend_ip_configuration.#").HasValue("1"),
			),
		},
	})
}

func TestAccLoadBalancer_tags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb", "test")
	r := LoadBalancer{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
				check.That(data.ResourceName).Key("tags.Environment").HasValue("production"),
				check.That(data.ResourceName).Key("tags.Purpose").HasValue("AcceptanceTests"),
			),
		},
		data.ImportStep(),
		{
			Config: r.updatedTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.Purpose").HasValue("AcceptanceTests"),
			),
		},
	})
}

func TestAccLoadBalancer_emptyPrivateIP(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb", "test")
	r := LoadBalancer{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.emptyPrivateIPAddress(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("frontend_ip_configuration.0.private_ip_address").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLoadBalancer_privateIP(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_lb", "test")
	r := LoadBalancer{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.privateIPAddress(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("frontend_ip_configuration.0.private_ip_address").Exists(),
			),
		},
	})
}

func (r LoadBalancer) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	loadBalancerName := state.Attributes["name"]
	resourceGroup := state.Attributes["resource_group_name"]

	resp, err := client.LoadBalancer.LoadBalancersClient.Get(ctx, resourceGroup, loadBalancerName, "")
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("Bad: Load Balancer %q (resource group: %q) does not exist", loadBalancerName, resourceGroup)
		}

		return nil, fmt.Errorf("Bad: Get on loadBalancerClient: %+v", err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (r LoadBalancer) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_lb" "test" {
  name                = "acctest-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  tags = {
    Environment = "production"
    Purpose     = "AcceptanceTests"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r LoadBalancer) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_lb" "import" {
  name                = azurestack_lb.test.name
  location            = azurestack_lb.test.location
  resource_group_name = azurestack_lb.test.resource_group_name

  tags = {
    Environment = "production"
    Purpose     = "AcceptanceTests"
  }
}
`, template)
}

func (r LoadBalancer) updatedTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_lb" "test" {
  name                = "acctest-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  tags = {
    Purpose = "AcceptanceTests"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r LoadBalancer) frontEndConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_public_ip" "test1" {
  name                = "another-test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "acctest-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }

  frontend_ip_configuration {
    name                 = "two-%d"
    public_ip_address_id = azurestack_public_ip.test1.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r LoadBalancer) frontEndConfigRemovalWithIP(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_public_ip" "test1" {
  name                = "another-test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "acctest-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r LoadBalancer) frontEndConfigRemoval(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "test-ip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "test" {
  name                = "acctest-loadbalancer-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = azurestack_public_ip.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r LoadBalancer) emptyPrivateIPAddress(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location
  sku                 = "Basic"

  frontend_ip_configuration {
    name                          = "Internal"
    private_ip_address_allocation = "Dynamic"
    private_ip_address            = ""
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r LoadBalancer) privateIPAddress(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-lb-%d"
  location = "%s"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_lb" "test" {
  name                = "acctestlb-%d"
  resource_group_name = azurestack_resource_group.test.name
  location            = azurestack_resource_group.test.location

  frontend_ip_configuration {
    name                          = "Internal"
    private_ip_address_allocation = "Static"
    private_ip_address            = "10.0.2.7"
    subnet_id                     = azurestack_subnet.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
