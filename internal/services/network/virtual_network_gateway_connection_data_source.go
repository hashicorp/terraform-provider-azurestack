package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func virtualNetworkGatewayConnectionDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: virtualNetworkGatewayConnectionDataSourceRead,

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

			"type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"shared_key": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"authorization_key": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"enable_bgp": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"ingress_bytes_transferred": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"egress_bytes_transferred": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"use_policy_based_traffic_selectors": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"routing_weight": {
				Type:     pluginsdk.TypeInt,
				Computed: true,
			},

			"virtual_network_gateway_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"peer_virtual_network_gateway_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"local_network_gateway_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"express_route_circuit_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"resource_guid": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"ipsec_policy": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"sa_lifetime": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},
						"sa_datasize": {
							Type:     pluginsdk.TypeInt,
							Computed: true,
						},
						"ipsec_encryption": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"ipsec_integrity": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"ike_encryption": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"ike_integrity": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"dh_group": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
						"pfs_group": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func virtualNetworkGatewayConnectionDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayConnectionsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewNetworkGatewayConnectionID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.ConnectionName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}

		return fmt.Errorf("making Read request on %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if resp.VirtualNetworkGatewayConnectionPropertiesFormat != nil {
		gwc := *resp.VirtualNetworkGatewayConnectionPropertiesFormat

		d.Set("shared_key", gwc.SharedKey)
		d.Set("authorization_key", gwc.AuthorizationKey)
		d.Set("enable_bgp", gwc.EnableBgp)
		d.Set("ingress_bytes_transferred", gwc.IngressBytesTransferred)
		d.Set("egress_bytes_transferred", gwc.EgressBytesTransferred)
		d.Set("use_policy_based_traffic_selectors", gwc.UsePolicyBasedTrafficSelectors)
		d.Set("type", string(gwc.ConnectionType))
		d.Set("routing_weight", gwc.RoutingWeight)

		if gwc.VirtualNetworkGateway1 != nil {
			d.Set("virtual_network_gateway_id", gwc.VirtualNetworkGateway1.ID)
		}

		if gwc.VirtualNetworkGateway2 != nil {
			d.Set("peer_virtual_network_gateway_id", gwc.VirtualNetworkGateway2.ID)
		}

		if gwc.LocalNetworkGateway2 != nil {
			d.Set("local_network_gateway_id", gwc.LocalNetworkGateway2.ID)
		}

		if gwc.Peer != nil {
			d.Set("express_route_circuit_id", gwc.Peer.ID)
		}

		d.Set("resource_guid", gwc.ResourceGUID)

		ipsecPoliciesSettingsFlat := flattenVirtualNetworkGatewayConnectionDataSourceIpsecPolicies(gwc.IpsecPolicies)
		if err := d.Set("ipsec_policy", ipsecPoliciesSettingsFlat); err != nil {
			return fmt.Errorf("setting `ipsec_policy`: %+v", err)
		}
	}

	return nil
}

func flattenVirtualNetworkGatewayConnectionDataSourceIpsecPolicies(ipsecPolicies *[]network.IpsecPolicy) []interface{} {
	schemaIpsecPolicies := make([]interface{}, 0)

	if ipsecPolicies != nil {
		for _, ipsecPolicy := range *ipsecPolicies {
			schemaIpsecPolicy := make(map[string]interface{})

			schemaIpsecPolicy["dh_group"] = string(ipsecPolicy.DhGroup)
			schemaIpsecPolicy["ike_encryption"] = string(ipsecPolicy.IkeEncryption)
			schemaIpsecPolicy["ike_integrity"] = string(ipsecPolicy.IkeIntegrity)
			schemaIpsecPolicy["ipsec_encryption"] = string(ipsecPolicy.IpsecEncryption)
			schemaIpsecPolicy["ipsec_integrity"] = string(ipsecPolicy.IpsecIntegrity)
			schemaIpsecPolicy["pfs_group"] = string(ipsecPolicy.PfsGroup)

			if ipsecPolicy.SaDataSizeKilobytes != nil {
				schemaIpsecPolicy["sa_datasize"] = int(*ipsecPolicy.SaDataSizeKilobytes)
			}

			if ipsecPolicy.SaLifeTimeSeconds != nil {
				schemaIpsecPolicy["sa_lifetime"] = int(*ipsecPolicy.SaLifeTimeSeconds)
			}

			schemaIpsecPolicies = append(schemaIpsecPolicies, schemaIpsecPolicy)
		}
	}

	return schemaIpsecPolicies
}
