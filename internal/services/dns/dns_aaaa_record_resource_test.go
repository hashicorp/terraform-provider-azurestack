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

type DnsAAAARecordResource struct{}

func TestAccDnsAAAARecord_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}

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

func TestAccDnsAAAARecord_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurestack_dns_aaaa_record"),
		},
	})
}

func TestAccDnsAAAARecord_updateRecords(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}

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

func TestAccDnsAAAARecord_withTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}

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

func TestAccDnsAAAARecord_withAlias(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}
	targetResourceName := "azurestack_public_ip.test"
	targetResourceName2 := "azurestack_public_ip.test2"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.withAlias(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttrPair(data.ResourceName, "target_resource_id", targetResourceName, "id"),
			),
		},
		{
			Config: r.withAliasUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttrPair(data.ResourceName, "target_resource_id", targetResourceName2, "id"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDnsAAAARecord_RecordsToAlias(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}
	targetResourceName := "azurestack_public_ip.test"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.AliasToRecordsUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("records.#").HasValue("2"),
			),
		},
		{
			Config: r.AliasToRecords(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttrPair(data.ResourceName, "target_resource_id", targetResourceName, "id"),
				acceptance.TestCheckNoResourceAttr(data.ResourceName, "records"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDnsAaaaRecord_AliasToRecords(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}
	targetResourceName := "azurestack_public_ip.test"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.AliasToRecords(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				acceptance.TestCheckResourceAttrPair(data.ResourceName, "target_resource_id", targetResourceName, "id"),
			),
		},
		{
			Config: r.AliasToRecordsUpdate(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("records.#").HasValue("2"),
				check.That(data.ResourceName).Key("target_resource_id").HasValue(""),
			),
		},
		data.ImportStep(),
	})
}

func TestAccDnsAAAARecord_uncompressed(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurestack_dns_aaaa_record", "test")
	r := DnsAAAARecordResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.uncompressed(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("fqdn").Exists(),
			),
		},
		{
			Config: r.uncompressed(data), // just use the same for updating
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("records.#").HasValue("2"),
			),
		},
		data.ImportStep(),
	})
}

func (DnsAAAARecordResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.AaaaRecordID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Dns.RecordSetsClient.Get(ctx, id.ResourceGroup, id.DnszoneName, id.AAAAName, dns.AAAA)
	if err != nil {
		return nil, fmt.Errorf("retrieving DNS AAAA record %s (resource group: %s): %v", id.AAAAName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.RecordSetProperties != nil), nil
}

func (DnsAAAARecordResource) basic(data acceptance.TestData) string {
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

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  records             = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) requiresImport(data acceptance.TestData) string {
	template := DnsAAAARecordResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurestack_dns_aaaa_record" "import" {
  name                = azurestack_dns_aaaa_record.test.name
  resource_group_name = azurestack_dns_aaaa_record.test.resource_group_name
  zone_name           = azurestack_dns_aaaa_record.test.zone_name
  ttl                 = 300
  records             = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]
}
`, template)
}

func (DnsAAAARecordResource) updateRecords(data acceptance.TestData) string {
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

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  records             = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006", "::1"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) withTags(data acceptance.TestData) string {
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

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  records             = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]

  tags = {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) withTagsUpdate(data acceptance.TestData) string {
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

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  records             = ["2607:f8b0:4009:1803::1005", "2607:f8b0:4009:1803::1006"]

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) withAlias(data acceptance.TestData) string {
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

resource "azurestack_public_ip" "test" {
  name                = "mypublicip%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
  ip_version          = "IPv6"
}

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myaaaarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  target_resource_id  = azurestack_public_ip.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) withAliasUpdate(data acceptance.TestData) string {
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

resource "azurestack_public_ip" "test2" {
  name                = "mypublicip%d2"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
  ip_version          = "IPv6"
}

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myaaaarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  target_resource_id  = azurestack_public_ip.test2.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) AliasToRecords(data acceptance.TestData) string {
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

resource "azurestack_public_ip" "test" {
  name                = "mypublicip%d"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
  allocation_method   = "Dynamic"
  ip_version          = "IPv6"
}

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  target_resource_id  = azurestack_public_ip.test.id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) AliasToRecordsUpdate(data acceptance.TestData) string {
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

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  records             = ["3a62:353:8885:293c:a218:45cc:9ee9:4e27", "3a62:353:8885:293c:a218:45cc:9ee9:4e28"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (DnsAAAARecordResource) uncompressed(data acceptance.TestData) string {
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

resource "azurestack_dns_aaaa_record" "test" {
  name                = "myarecord%d"
  resource_group_name = azurestack_resource_group.test.name
  zone_name           = azurestack_dns_zone.test.name
  ttl                 = 300
  records             = ["2607:f8b0:4005:0800:0000:0000:0000:1003", "2201:1234:1234::1"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
