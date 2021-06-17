package network

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourcePublicIP() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourcePublicIPRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"location": azure.SchemaLocationForDataSource(),

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"sku": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"allocation_method": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"ip_version": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"domain_name_label": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"idle_timeout_in_minutes": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"reverse_fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"ip_address": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"ip_tags": {
				Type:     pluginsdk.TypeMap,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"zones": azure.SchemaZonesComputed(),

			"tags": tags.Schema(),
		},
	}
}

func dataSourcePublicIPRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	resp, err := client.Get(ctx, resGroup, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: Public IP %q (Resource Group %q) was not found", name, resGroup)
		}
		return fmt.Errorf("Error making Read request on Azure public ip %s: %s", name, err)
	}

	d.SetId(*resp.ID)

	d.Set("zones", resp.Zones)

	// ensure values are at least set to "", d.Set() is a noop on a nil
	// there must be a better way...
	d.Set("location", "")
	d.Set("sku", "")
	d.Set("fqdn", "")
	d.Set("reverse_fqdn", "")
	d.Set("domain_name_label", "")
	d.Set("allocation_method", "")
	d.Set("ip_address", "")
	d.Set("ip_version", "")
	d.Set("idle_timeout_in_minutes", 0)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku", string(sku.Name))
	}

	if props := resp.PublicIPAddressPropertiesFormat; props != nil {
		if dnsSettings := props.DNSSettings; dnsSettings != nil {
			d.Set("fqdn", dnsSettings.Fqdn)
			d.Set("reverse_fqdn", dnsSettings.ReverseFqdn)
			d.Set("domain_name_label", dnsSettings.DomainNameLabel)
		}

		d.Set("allocation_method", string(props.PublicIPAllocationMethod))
		d.Set("ip_address", props.IPAddress)
		d.Set("ip_version", string(props.PublicIPAddressVersion))
		d.Set("idle_timeout_in_minutes", props.IdleTimeoutInMinutes)

		iptags := flattenPublicIpPropsIpTags(*props.IPTags)
		if iptags != nil {
			d.Set("ip_tags", iptags)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
