package network_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
)

type PublicIPResource struct{}

func TestAccPublicIpStatic_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static_basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_address").Exists(),
				check.That(data.ResourceName).Key("allocation_method").HasValue("Static"),
				check.That(data.ResourceName).Key("ip_version").HasValue("IPv4"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static_basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_address").Exists(),
				check.That(data.ResourceName).Key("allocation_method").HasValue("Static"),
				check.That(data.ResourceName).Key("ip_version").HasValue("IPv4"),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_public_ip"),
		},
	})
}

func TestAccPublicIpStatic_basic_withDNSLabel(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}
	dnl := fmt.Sprintf("acctestdnl-%d", data.RandomInteger)

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic_withDNSLabel(data, dnl),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_address").Exists(),
				check.That(data.ResourceName).Key("allocation_method").HasValue("Static"),
				check.That(data.ResourceName).Key("domain_name_label").HasValue(dnl),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_standard_withIPv6(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.standard_withIPVersion(data, "IPv6"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_version").HasValue("IPv6"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpDynamic_basic_withIPv6(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}
	ipVersion := "Ipv6"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.dynamic_basic_withIPVersion(data, ipVersion),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_version").HasValue("IPv6"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_basic_defaultsToIPv4(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static_basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_version").HasValue("IPv4"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_basic_withIPv4(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}
	ipVersion := "IPv4"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static_basic_withIPVersion(data, ipVersion),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_version").HasValue("IPv4"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_standard(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.standard(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.static_basic,
			TestResource: r,
		}),
	})
}

func TestAccPublicIpStatic_idleTimeout(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.idleTimeout(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("idle_timeout_in_minutes").HasValue("30"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("Production"),
				check.That(data.ResourceName).Key("tags.cost_center").HasValue("MSFT"),
			),
		},
		{
			Config: r.withTagsUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.environment").HasValue("staging"),
			),
		},
	})
}

func TestAccPublicIpStatic_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static_basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("domain_name_label").HasValue(fmt.Sprintf("acctest-%d", data.RandomInteger)),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpDynamic_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.dynamic_basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPublicIpStatic_importIdError(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.static_basic(data),
		},
		{
			ResourceName:      data.ResourceName,
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateId:     fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/publicIPAdresses/acctestpublicip-%d", os.Getenv("ARM_SUBSCRIPTION_ID"), data.RandomInteger, data.RandomInteger),
			ExpectError:       regexp.MustCompile("Error: parsing Resource ID"),
		},
	})
}

func TestAccPublicIpStatic_canLabelBe63(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_public_ip", "test")
	r := PublicIPResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.canLabelBe63(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("ip_address").Exists(),
				check.That(data.ResourceName).Key("allocation_method").HasValue("Static"),
			),
		},
		data.ImportStep(),
	})
}

func (t PublicIPResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.PublicIpAddressID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Network.PublicIPsClient.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		return nil, fmt.Errorf("reading Public IP %s: %+v", *id, err)
	}

	return pointer.FromBool(resp.ID != nil), nil
}

func (PublicIPResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.PublicIpAddressID(state.ID)
	if err != nil {
		return nil, err
	}

	future, err := client.Network.PublicIPsClient.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return nil, fmt.Errorf("deleting Public IP %q: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Network.PublicIPsClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for Deletion of Public IP %s: %+v", *id, err)
	}

	return pointer.FromBool(true), nil
}

func (PublicIPResource) static_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r PublicIPResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurestack_public_ip" "import" {
  name                = azurestack_public_ip.test.name
  location            = azurestack_public_ip.test.location
  resource_group_name = azurestack_public_ip.test.resource_group_name
  allocation_method   = azurestack_public_ip.test.allocation_method
}
`, r.static_basic(data))
}

func (PublicIPResource) basic_withDNSLabel(data acceptance.TestData, dnsNameLabel string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  domain_name_label   = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, dnsNameLabel)
}

func (PublicIPResource) static_basic_withIPVersion(data acceptance.TestData, ipVersion string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  ip_version          = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, ipVersion)
}

func (PublicIPResource) standard(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  sku                 = "Standard"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (PublicIPResource) standard_withIPVersion(data acceptance.TestData, ipVersion string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  ip_version          = "%s"
  sku                 = "Standard"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, ipVersion)
}

func (PublicIPResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"
  domain_name_label   = "acctest-%d"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (PublicIPResource) idleTimeout(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                    = "acctestpublicip-%d"
  location                = azurestack_resource_group.test.location
  resource_group_name     = azurestack_resource_group.test.name
  allocation_method       = "Static"
  idle_timeout_in_minutes = 30
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (PublicIPResource) dynamic_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (PublicIPResource) dynamic_basic_withIPVersion(data acceptance.TestData, ipVersion string) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"

  ip_version = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, ipVersion)
}

func (PublicIPResource) withTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (PublicIPResource) withTagsUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Static"

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (PublicIPResource) canLabelBe63(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_public_ip" "test" {
  name                = "acctestpublicip-%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  allocation_method = "Static"
  domain_name_label = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomStringOfLength(63))
}
