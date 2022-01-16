package dns

import (
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
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
		// "azurestack_dns_zone": dataSourceDnsZone(), todo
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurestack_dns_a_record": dnsARecord(),
		/* dodo copy in
		"azurestack_dns_aaaa_record":  resourceDnsAAAARecord(),
		"azurestack_dns_caa_record":   resourceDnsCaaRecord(),
		"azurestack_dns_cname_record": resourceDnsCNameRecord(),
		"azurestack_dns_mx_record":    resourceDnsMxRecord(),
		"azurestack_dns_ns_record":    resourceDnsNsRecord(),
		"azurestack_dns_ptr_record":   resourceDnsPtrRecord(),
		"azurestack_dns_srv_record":   resourceDnsSrvRecord(),
		"azurestack_dns_txt_record":   resourceDnsTxtRecord(),*/
		"azurestack_dns_zone": resourceDnsZone(),
	}
}
