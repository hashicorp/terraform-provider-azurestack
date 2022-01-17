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

func TestAccPublicIPsDataSource_assigned(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurestack_public_ips", "test")
	r := PublicIPsResource{}

	attachedDataSourceName := "data.azurestack_public_ips.attached"
	attachedDataSourceName2 := "data.azurestack_public_ips.attached2"
	unattachedDataSourceName := "data.azurestack_public_ips.unattached"
	unattachedDataSourceName2 := "data.azurestack_public_ips.unattached2"
	unattachedAndAttachedDataSourceName := "data.azurestack_public_ips.unattached_and_attached"

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.attached(data),
		},
		{
			Config: r.attachedDataSource(data),
			Check: acceptance.ComposeTestCheckFunc(
				acceptance.TestCheckResourceAttr(attachedDataSourceName, "public_ips.#", "4"),
				acceptance.TestCheckResourceAttr(attachedDataSourceName, "public_ips.0.name", fmt.Sprintf("acctestpip%s-0", data.RandomString)),
				acceptance.TestCheckResourceAttr(attachedDataSourceName, "public_ips.3.name", fmt.Sprintf("acctestpip%s-3", data.RandomString)),
				acceptance.TestCheckResourceAttr(attachedDataSourceName2, "public_ips.#", "4"),
				acceptance.TestCheckResourceAttr(attachedDataSourceName2, "public_ips.0.name", fmt.Sprintf("acctestpip%s-0", data.RandomString)),
				acceptance.TestCheckResourceAttr(attachedDataSourceName2, "public_ips.3.name", fmt.Sprintf("acctestpip%s-3", data.RandomString)),
				acceptance.TestCheckResourceAttr(unattachedDataSourceName, "public_ips.#", "4"),
				acceptance.TestCheckResourceAttr(unattachedDataSourceName, "public_ips.0.name", fmt.Sprintf("acctestpip%s-4", data.RandomString)),
				acceptance.TestCheckResourceAttr(unattachedDataSourceName2, "public_ips.#", "4"),
				acceptance.TestCheckResourceAttr(unattachedDataSourceName2, "public_ips.0.name", fmt.Sprintf("acctestpip%s-4", data.RandomString)),
				acceptance.TestCheckResourceAttr(unattachedAndAttachedDataSourceName, "public_ips.#", "8"),
				acceptance.TestCheckResourceAttr(unattachedAndAttachedDataSourceName, "public_ips.0.name", fmt.Sprintf("acctestpip%s-0", data.RandomString)),
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

func (PublicIPsResource) attached(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  count                   = 8
  name                    = "acctestpip%s-${count.index}"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  allocation_method       = "Static"
  idle_timeout_in_minutes = 30
  sku                     = "Standard"

  tags = {
    environment = "test"
  }
}

resource "azurestack_lb" "test" {
  count               = 3
  name                = "acctestlb-${count.index}"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  sku                 = "Standard"

  frontend_ip_configuration {
    name                 = "frontend"
    public_ip_address_id = element(azurestack_public_ip.test.*.id, count.index)
  }
}

resource "azurestack_nat_gateway" "test" {
  name                    = "nat-Gateway"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
}

resource "azurestack_nat_gateway_public_ip_association" "test" {
  nat_gateway_id       = azurestack_nat_gateway.test.id
  public_ip_address_id = element(azurestack_public_ip.test.*.id, 3)
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func (r PublicIPsResource) attachedDataSource(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurestack_public_ips" "unattached" {
  resource_group_name = azurestack_resource_group.test.name
  attachment_status   = "Unattached"
}

data "azurestack_public_ips" "unattached2" {
  resource_group_name = azurestack_resource_group.test.name
  attached            = false
}

data "azurestack_public_ips" "unattached_and_attached" {
  resource_group_name = azurestack_resource_group.test.name
  attachment_status   = "All"
}

data "azurestack_public_ips" "attached" {
  resource_group_name = azurestack_resource_group.test.name
  attachment_status   = "Attached"
}

data "azurestack_public_ips" "attached2" {
  resource_group_name = azurestack_resource_group.test.name
  attached            = true
}
`, r.attached(data))
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
