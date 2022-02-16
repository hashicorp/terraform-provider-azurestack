package dns_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/dns/mgmt/dns"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/dns/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

type DnsNsRecordResource struct{}

func TestAccDnsNsRecord_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_ns_record", "test")
	r := DnsNsRecordResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("fqdn").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDnsNsRecord_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_ns_record", "test")
	r := DnsNsRecordResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_dns_ns_record"),
		},
	})
}

func TestAccDnsNsRecord_updateRecords(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_ns_record", "test")
	r := DnsNsRecordResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("records.#").HasValue("2"),
			),
		},
		{
			Config: r.updateRecords(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("records.#").HasValue("3"),
			),
		},
	})
}

func TestAccDnsNsRecord_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_ns_record", "test")
	r := DnsNsRecordResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withTags(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
			),
		},
		{
			Config: r.withTagsUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func (DnsNsRecordResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.NsRecordID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Dns.RecordSetsClient.Get(ctx, id.ResourceGroup, id.DnszoneName, id.NSName, dns.NS)
	if err != nil {
		return nil, fmt.Errorf("retrieving DNS NS record %s (resource group: %s): %v", id.NSName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.RecordSetProperties != nil), nil
}

func (DnsNsRecordResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_dns_ns_record" "test" {
  name                = "mynsrecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300

  records = ["ns1.contoso.com", "ns2.contoso.com"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsNsRecordResource) requiresImport(data acceptance.TestData) string {
	template := DnsNsRecordResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_dns_ns_record" "import" {
  name                = azurestack_dns_ns_record.test.name
  resource_group_name = azurestack_dns_ns_record.test.resource_group_name
  zone_name           = azurestack_dns_ns_record.test.zone_name
  ttl                 = 300

  records = ["ns1.contoso.com", "ns2.contoso.com"]
}
`, template)
}

func (DnsNsRecordResource) updateRecords(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_dns_ns_record" "test" {
  name                = "mynsrecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300

  records = ["ns1.contoso.com", "ns2.contoso.com", "ns3.contoso.com"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsNsRecordResource) withTags(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_dns_ns_record" "test" {
  name                = "mynsrecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300

  records = ["ns1.contoso.com", "ns2.contoso.com"]

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsNsRecordResource) withTagsUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurestack" {
  features {}
}

resource "azurestack_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurestack_dns_zone" "test" {
  name                = "acctestzone%d.com"
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_dns_ns_record" "test" {
  name                = "mynsrecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300

  records = ["ns1.contoso.com", "ns2.contoso.com"]

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
