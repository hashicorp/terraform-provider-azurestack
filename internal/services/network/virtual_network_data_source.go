// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func virtualNetworkDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: virtualNetworkDataSourceRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": commonschema.ResourceGroupNameForDataSource(),

			"location": commonschema.LocationComputed(),

			"address_space": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"dns_servers": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"guid": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"subnets": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"vnet_peerings": {
				Type:     pluginsdk.TypeMap,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func virtualNetworkDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewVirtualNetworkID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: %s was not found", id)
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.VirtualNetworkPropertiesFormat; props != nil {
		d.Set("guid", props.ResourceGUID)

		if as := props.AddressSpace; as != nil {
			if err := d.Set("address_space", utils.FlattenStringSlice(as.AddressPrefixes)); err != nil {
				return fmt.Errorf("setting `address_space`: %v", err)
			}
		}

		if options := props.DhcpOptions; options != nil {
			if err := d.Set("dns_servers", utils.FlattenStringSlice(options.DNSServers)); err != nil {
				return fmt.Errorf("setting `dns_servers`: %v", err)
			}
		}

		if err := d.Set("subnets", flattenVnetSubnetsNames(props.Subnets)); err != nil {
			return fmt.Errorf("setting `subnets`: %v", err)
		}

		if err := d.Set("vnet_peerings", flattenVnetPeerings(props.VirtualNetworkPeerings)); err != nil {
			return fmt.Errorf("setting `vnet_peerings`: %v", err)
		}
	}
	return nil
}

func flattenVnetSubnetsNames(input *[]network.Subnet) []interface{} {
	subnets := make([]interface{}, 0)

	if mysubnets := input; mysubnets != nil {
		for _, subnet := range *mysubnets {
			if v := subnet.Name; v != nil {
				subnets = append(subnets, *v)
			}
		}
	}
	return subnets
}

func flattenVnetPeerings(input *[]network.VirtualNetworkPeering) map[string]interface{} {
	output := make(map[string]interface{})

	if peerings := input; peerings != nil {
		for _, vnetpeering := range *peerings {
			if vnetpeering.Name == nil || vnetpeering.RemoteVirtualNetwork == nil || vnetpeering.RemoteVirtualNetwork.ID == nil {
				continue
			}

			key := *vnetpeering.Name
			value := *vnetpeering.RemoteVirtualNetwork.ID

			output[key] = value
		}
	}

	return output
}
