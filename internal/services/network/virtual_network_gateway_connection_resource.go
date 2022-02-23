package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceid"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func virtualNetworkGatewayConnection() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: virtualNetworkGatewayConnectionCreateUpdate,
		Read:   virtualNetworkGatewayConnectionRead,
		Update: virtualNetworkGatewayConnectionCreateUpdate,
		Delete: virtualNetworkGatewayConnectionDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.NetworkGatewayConnectionID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"location": commonschema.Location(),

			"type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.ExpressRoute),
					string(network.IPsec),
					string(network.Vnet2Vnet),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"virtual_network_gateway_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: resourceid.ValidateResourceID,
			},

			"authorization_key": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"express_route_circuit_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
			},

			"peer_virtual_network_gateway_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
			},

			"local_network_gateway_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
			},

			"enable_bgp": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"use_policy_based_traffic_selectors": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"routing_weight": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 32000),
			},

			"shared_key": {
				Type:      pluginsdk.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"ipsec_policy": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"dh_group": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.DHGroup1),
								string(network.DHGroup14),
								string(network.DHGroup2),
								string(network.DHGroup2048),
								string(network.DHGroup24),
								string(network.ECP256),
								string(network.ECP384),
								string(network.None),
							}, true),
						},

						"ike_encryption": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.AES128),
								string(network.AES192),
								string(network.AES256),
								string(network.DES),
								string(network.DES3),
							}, true),
						},

						"ike_integrity": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.IkeIntegrityMD5),
								string(network.IkeIntegritySHA1),
								string(network.IkeIntegritySHA256),
								string(network.IkeIntegritySHA384),
							}, true),
						},

						"ipsec_encryption": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.IpsecEncryptionAES128),
								string(network.IpsecEncryptionAES192),
								string(network.IpsecEncryptionAES256),
								string(network.IpsecEncryptionDES),
								string(network.IpsecEncryptionDES3),
								string(network.IpsecEncryptionGCMAES128),
								string(network.IpsecEncryptionGCMAES192),
								string(network.IpsecEncryptionGCMAES256),
								string(network.IpsecEncryptionNone),
							}, true),
						},

						"ipsec_integrity": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.IpsecIntegrityGCMAES128),
								string(network.IpsecIntegrityGCMAES192),
								string(network.IpsecIntegrityGCMAES256),
								string(network.IpsecIntegrityMD5),
								string(network.IpsecIntegritySHA1),
								string(network.IpsecIntegritySHA256),
							}, true),
						},

						"pfs_group": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.PfsGroupECP256),
								string(network.PfsGroupECP384),
								string(network.PfsGroupNone),
								string(network.PfsGroupPFS1),
								string(network.PfsGroupPFS2),
								string(network.PfsGroupPFS2048),
								string(network.PfsGroupPFS24),
							}, true),
						},

						"sa_datasize": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(1024),
						},

						"sa_lifetime": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(300),
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func virtualNetworkGatewayConnectionCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayConnectionsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for azurestack Virtual Network Gateway Connection creation.")

	id := parse.NewNetworkGatewayConnectionID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ConnectionName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurestack_virtual_network_gateway_connection", *existing.ID)
		}
	}

	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	properties, err := getVirtualNetworkGatewayConnectionProperties(d)
	if err != nil {
		return err
	}

	connection := network.VirtualNetworkGatewayConnection{
		Name:     &id.ConnectionName,
		Location: &location,
		Tags:     tags.Expand(t),
		VirtualNetworkGatewayConnectionPropertiesFormat: properties,
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ConnectionName, connection)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	if properties.SharedKey != nil && !d.IsNewResource() {
		future, err := client.SetSharedKey(ctx, id.ResourceGroup, id.ConnectionName, network.ConnectionSharedKey{
			Value: properties.SharedKey,
		})
		if err != nil {
			return fmt.Errorf("updating Shared Key for %s: %+v", id, err)
		}
		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for updating Shared Key for %s: %+v", id, err)
		}
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return virtualNetworkGatewayConnectionRead(d, meta)
}

func virtualNetworkGatewayConnectionRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayConnectionsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NetworkGatewayConnectionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ConnectionName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on %s: %+v", id, err)
	}

	conn := *resp.VirtualNetworkGatewayConnectionPropertiesFormat

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if string(conn.ConnectionType) != "" {
		d.Set("type", string(conn.ConnectionType))
	}

	if conn.VirtualNetworkGateway1 != nil {
		d.Set("virtual_network_gateway_id", conn.VirtualNetworkGateway1.ID)
	}

	if conn.AuthorizationKey != nil {
		d.Set("authorization_key", conn.AuthorizationKey)
	}

	if conn.Peer != nil {
		d.Set("express_route_circuit_id", conn.Peer.ID)
	}

	if conn.VirtualNetworkGateway2 != nil {
		d.Set("peer_virtual_network_gateway_id", conn.VirtualNetworkGateway2.ID)
	}

	if conn.LocalNetworkGateway2 != nil {
		d.Set("local_network_gateway_id", conn.LocalNetworkGateway2.ID)
	}

	if conn.EnableBgp != nil {
		d.Set("enable_bgp", conn.EnableBgp)
	}

	if conn.UsePolicyBasedTrafficSelectors != nil {
		d.Set("use_policy_based_traffic_selectors", conn.UsePolicyBasedTrafficSelectors)
	}

	if conn.RoutingWeight != nil {
		d.Set("routing_weight", conn.RoutingWeight)
	}

	if conn.SharedKey != nil {
		d.Set("shared_key", conn.SharedKey)
	}

	if conn.IpsecPolicies != nil {
		ipsecPolicies := flattenVirtualNetworkGatewayConnectionIpsecPolicies(conn.IpsecPolicies)

		if err := d.Set("ipsec_policy", ipsecPolicies); err != nil {
			return fmt.Errorf("setting `ipsec_policy`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func virtualNetworkGatewayConnectionDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayConnectionsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NetworkGatewayConnectionID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.ConnectionName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", id, err)
	}

	return nil
}

func getVirtualNetworkGatewayConnectionProperties(d *pluginsdk.ResourceData) (*network.VirtualNetworkGatewayConnectionPropertiesFormat, error) {
	connectionType := network.VirtualNetworkGatewayConnectionType(d.Get("type").(string))

	props := &network.VirtualNetworkGatewayConnectionPropertiesFormat{
		ConnectionType:                 connectionType,
		EnableBgp:                      pointer.FromBool(d.Get("enable_bgp").(bool)),
		UsePolicyBasedTrafficSelectors: pointer.FromBool(d.Get("use_policy_based_traffic_selectors").(bool)),
	}

	if v, ok := d.GetOk("virtual_network_gateway_id"); ok {
		virtualNetworkGatewayId := v.(string)

		gwid, err := parse.VirtualNetworkGatewayID(virtualNetworkGatewayId)
		if err != nil {
			return nil, err
		}

		props.VirtualNetworkGateway1 = &network.VirtualNetworkGateway{
			ID:   &virtualNetworkGatewayId,
			Name: &gwid.Name,
			VirtualNetworkGatewayPropertiesFormat: &network.VirtualNetworkGatewayPropertiesFormat{
				IPConfigurations: &[]network.VirtualNetworkGatewayIPConfiguration{},
			},
		}
	}

	if v, ok := d.GetOk("authorization_key"); ok {
		authorizationKey := v.(string)
		props.AuthorizationKey = &authorizationKey
	}

	if v, ok := d.GetOk("express_route_circuit_id"); ok {
		expressRouteCircuitId := v.(string)
		props.Peer = &network.SubResource{
			ID: &expressRouteCircuitId,
		}
	}

	if v, ok := d.GetOk("peer_virtual_network_gateway_id"); ok {
		gwid, err := parse.VirtualNetworkGatewayID(v.(string))
		if err != nil {
			return nil, err
		}
		props.VirtualNetworkGateway2 = &network.VirtualNetworkGateway{
			ID:   pointer.FromString(gwid.ID()),
			Name: &gwid.Name,
			VirtualNetworkGatewayPropertiesFormat: &network.VirtualNetworkGatewayPropertiesFormat{
				IPConfigurations: &[]network.VirtualNetworkGatewayIPConfiguration{},
			},
		}
	}

	if v, ok := d.GetOk("local_network_gateway_id"); ok {
		localNetworkGatewayId := v.(string)
		id, err := parse.LocalNetworkGatewayID(localNetworkGatewayId)
		if err != nil {
			return nil, err
		}

		props.LocalNetworkGateway2 = &network.LocalNetworkGateway{
			ID:   &localNetworkGatewayId,
			Name: &id.Name,
			LocalNetworkGatewayPropertiesFormat: &network.LocalNetworkGatewayPropertiesFormat{
				LocalNetworkAddressSpace: &network.AddressSpace{},
			},
		}
	}

	if v, ok := d.GetOk("routing_weight"); ok {
		routingWeight := int32(v.(int))
		props.RoutingWeight = &routingWeight
	}

	if v, ok := d.GetOk("shared_key"); ok {
		props.SharedKey = pointer.FromString(v.(string))
	}

	if v, ok := d.GetOk("ipsec_policy"); ok {
		props.IpsecPolicies = expandVirtualNetworkGatewayConnectionIpsecPolicies(v.([]interface{}))
	}

	if props.ConnectionType == network.ExpressRoute {
		if props.Peer == nil || props.Peer.ID == nil {
			return nil, fmt.Errorf("`express_route_circuit_id` must be specified when `type` is set to `ExpressRoute`")
		}
	}

	if props.ConnectionType == network.IPsec {
		if props.LocalNetworkGateway2 == nil || props.LocalNetworkGateway2.ID == nil {
			return nil, fmt.Errorf("`local_network_gateway_id` must be specified when `type` is set to `IPsec`")
		}
	}

	if props.ConnectionType == network.Vnet2Vnet {
		if props.VirtualNetworkGateway2 == nil || props.VirtualNetworkGateway2.ID == nil {
			return nil, fmt.Errorf("`peer_virtual_network_gateway_id` must be specified when `type` is set to `Vnet2Vnet`")
		}
	}

	return props, nil
}

func expandVirtualNetworkGatewayConnectionIpsecPolicies(schemaIpsecPolicies []interface{}) *[]network.IpsecPolicy {
	ipsecPolicies := make([]network.IpsecPolicy, 0, len(schemaIpsecPolicies))

	for _, d := range schemaIpsecPolicies {
		schemaIpsecPolicy := d.(map[string]interface{})
		ipsecPolicy := &network.IpsecPolicy{}

		if dhGroup, ok := schemaIpsecPolicy["dh_group"].(string); ok && dhGroup != "" {
			ipsecPolicy.DhGroup = network.DhGroup(dhGroup)
		}

		if ikeEncryption, ok := schemaIpsecPolicy["ike_encryption"].(string); ok && ikeEncryption != "" {
			ipsecPolicy.IkeEncryption = network.IkeEncryption(ikeEncryption)
		}

		if ikeIntegrity, ok := schemaIpsecPolicy["ike_integrity"].(string); ok && ikeIntegrity != "" {
			ipsecPolicy.IkeIntegrity = network.IkeIntegrity(ikeIntegrity)
		}

		if ipsecEncryption, ok := schemaIpsecPolicy["ipsec_encryption"].(string); ok && ipsecEncryption != "" {
			ipsecPolicy.IpsecEncryption = network.IpsecEncryption(ipsecEncryption)
		}

		if ipsecIntegrity, ok := schemaIpsecPolicy["ipsec_integrity"].(string); ok && ipsecIntegrity != "" {
			ipsecPolicy.IpsecIntegrity = network.IpsecIntegrity(ipsecIntegrity)
		}

		if pfsGroup, ok := schemaIpsecPolicy["pfs_group"].(string); ok && pfsGroup != "" {
			ipsecPolicy.PfsGroup = network.PfsGroup(pfsGroup)
		}

		if v, ok := schemaIpsecPolicy["sa_datasize"].(int); ok {
			saDatasize := int32(v)
			ipsecPolicy.SaDataSizeKilobytes = &saDatasize
		}

		if v, ok := schemaIpsecPolicy["sa_lifetime"].(int); ok {
			saLifetime := int32(v)
			ipsecPolicy.SaLifeTimeSeconds = &saLifetime
		}

		ipsecPolicies = append(ipsecPolicies, *ipsecPolicy)
	}

	return &ipsecPolicies
}

func flattenVirtualNetworkGatewayConnectionIpsecPolicies(ipsecPolicies *[]network.IpsecPolicy) []interface{} {
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
