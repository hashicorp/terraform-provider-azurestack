// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/zones"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func publicIPDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: publicIPDataSourceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"location": commonschema.LocationComputed(),

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

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

			"zones": zones.SchemaComputed(),

			"tags": tags.Schema(),
		},
	}
}

func publicIPDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewPublicIpAddressID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("making Read request on %s: %s", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

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

	d.Set("location", location.NormalizeNilable(resp.Location))

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
	}

	return tags.FlattenAndSet(d, resp.Tags)
}
