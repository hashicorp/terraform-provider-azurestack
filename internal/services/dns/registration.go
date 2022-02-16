package dns

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/sdk"
)

var (
	_ sdk.TypedServiceRegistration   = Registration{}
	_ sdk.UntypedServiceRegistration = Registration{}
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "DNS"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"DNS",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_dns_zone": dnsZoneDataSource(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_dns_a_record":     dnsARecord(),
		"azurestack_dns_aaaa_record":  dnsAAAARecord(),
		"azurestack_dns_cname_record": dnsCNameRecord(),
		"azurestack_dns_mx_record":    dnsMxRecord(),
		"azurestack_dns_ns_record":    dnsNsRecord(),
		"azurestack_dns_ptr_record":   dnsPtrRecord(),
		"azurestack_dns_srv_record":   dnsSrvRecord(),
		"azurestack_dns_txt_record":   dnsTxtRecord(),
		"azurestack_dns_zone":         dnsZone(),
	}
}

// DataSources returns a list of Data Sources supported by this Service
func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}

// Resources returns a list of Resources supported by this Service
func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{}
}
