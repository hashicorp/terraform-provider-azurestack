package network

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceid"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func virtualNetworkGateway() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: virtualNetworkGatewayCreateUpdate,
		Read:   virtualNetworkGatewayRead,
		Update: virtualNetworkGatewayCreateUpdate,
		Delete: virtualNetworkGatewayDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VirtualNetworkGatewayID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
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
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.VirtualNetworkGatewayTypeExpressRoute),
					string(network.VirtualNetworkGatewayTypeVpn),
				}, true),
			},

			"vpn_type": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          string(network.RouteBased),
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.RouteBased),
					string(network.PolicyBased),
				}, true),
			},

			"enable_bgp": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"active_active": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"sku": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				// This validator checks for all possible values for the SKU regardless of the attributes vpn_type and
				// type. For a validation which depends on the attributes vpn_type and type, refer to the special case
				// validators validateVirtualNetworkGatewayPolicyBasedVpnSku, validateVirtualNetworkGatewayRouteBasedVpnSku
				// and validateVirtualNetworkGatewayExpressRouteSku.
				ValidateFunc: validation.Any(
					validateVirtualNetworkGatewayPolicyBasedVpnSku(),
					validateVirtualNetworkGatewayRouteBasedVpnSkuGeneration1(),
					validateVirtualNetworkGatewayRouteBasedVpnSkuGeneration2(),
					validateVirtualNetworkGatewayExpressRouteSku(),
				),
			},

			"ip_configuration": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 3,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							// Azure Management API requires a name but does not generate a name if the field is missing
							// The name "vnetGatewayConfig" is used when creating a virtual network gateway via the
							// Azure portal.
							Default: "vnetGatewayConfig",
						},

						"private_ip_address_allocation": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.Static),
								string(network.Dynamic),
							}, false),
							Default: string(network.Dynamic),
						},

						"subnet_id": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							ValidateFunc:     validate.IsGatewaySubnet,
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"public_ip_address_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
						},
					},
				},
			},

			"vpn_client_configuration": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"address_space": {
							Type:     pluginsdk.TypeList,
							Required: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
							},
						},

						"root_certificate": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:     pluginsdk.TypeString,
										Required: true,
									},
									"public_cert_data": {
										Type:     pluginsdk.TypeString,
										Required: true,
									},
								},
							},
							Set: hashVirtualNetworkGatewayRootCert,
						},

						"revoked_certificate": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Resource{
								Schema: map[string]*pluginsdk.Schema{
									"name": {
										Type:     pluginsdk.TypeString,
										Required: true,
									},
									"thumbprint": {
										Type:     pluginsdk.TypeString,
										Required: true,
									},
								},
							},
							Set: hashVirtualNetworkGatewayRevokedCert,
						},

						"radius_server_address": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
							RequiredWith: []string{"vpn_client_configuration.0.radius_server_secret"},
						},

						"radius_server_secret": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							RequiredWith: []string{"vpn_client_configuration.0.radius_server_address"},
						},

						"vpn_client_protocols": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type: pluginsdk.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									string(network.IkeV2),
									string(network.SSTP),
								}, true),
							},
						},
					},
				},
			},

			"bgp_settings": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"asn": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							AtLeastOneOf: []string{"bgp_settings.0.asn", "bgp_settings.0.peering_address", "bgp_settings.0.peer_weight"},
						},

						"peering_address": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							Deprecated:   "Deprecated in favor of `bgp_settings.0.peering_addresses.0.default_addresses.0`",
							AtLeastOneOf: []string{"bgp_settings.0.asn", "bgp_settings.0.peering_address", "bgp_settings.0.peer_weight"},
						},

						"peer_weight": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							AtLeastOneOf: []string{"bgp_settings.0.asn", "bgp_settings.0.peering_address", "bgp_settings.0.peer_weight"},
						},
					},
				},
			},

			"default_local_network_gateway_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
			},

			"tags": tags.Schema(),
		},
	}
}

func virtualNetworkGatewayCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	defer cancel()

	log.Printf("[INFO] preparing arguments for azurestack Virtual Network Gateway creation.")

	id := parse.NewVirtualNetworkGatewayID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_virtual_network_gateway", id.ID())
		}
	}

	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	properties, err := getVirtualNetworkGatewayProperties(d)
	if err != nil {
		return err
	}

	gateway := network.VirtualNetworkGateway{
		Name:                                  &id.Name,
		Location:                              &location,
		Tags:                                  tags.Expand(t),
		VirtualNetworkGatewayPropertiesFormat: properties,
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, gateway)
	if err != nil {
		return fmt.Errorf("Creating/Updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return virtualNetworkGatewayRead(d, meta)
}

func virtualNetworkGatewayRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkGatewayID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if gw := resp.VirtualNetworkGatewayPropertiesFormat; gw != nil {
		d.Set("type", string(gw.GatewayType))
		d.Set("enable_bgp", gw.EnableBgp)
		d.Set("active_active", gw.ActiveActive)

		if string(gw.VpnType) != "" {
			d.Set("vpn_type", string(gw.VpnType))
		}

		if gw.GatewayDefaultSite != nil {
			d.Set("default_local_network_gateway_id", gw.GatewayDefaultSite.ID)
		}

		if gw.Sku != nil {
			d.Set("sku", string(gw.Sku.Name))
		}

		if err := d.Set("ip_configuration", flattenVirtualNetworkGatewayIPConfigurations(gw.IPConfigurations)); err != nil {
			return fmt.Errorf("setting `ip_configuration`: %+v", err)
		}

		if err := d.Set("vpn_client_configuration", flattenVirtualNetworkGatewayVpnClientConfig(gw.VpnClientConfiguration)); err != nil {
			return fmt.Errorf("setting `vpn_client_configuration`: %+v", err)
		}

		bgpSettings := flattenVirtualNetworkGatewayBgpSettings(gw.BgpSettings)
		if err := d.Set("bgp_settings", bgpSettings); err != nil {
			return fmt.Errorf("setting `bgp_settings`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func virtualNetworkGatewayDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.VnetGatewayClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkGatewayID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

// NOTE: these methods are deprecated, but provided to ease compatibility for open PR's
// TODO remove this function
func evaluateSchemaValidateFunc(i interface{}, k string, validateFunc pluginsdk.SchemaValidateFunc) (bool, error) {
	_, errors := validateFunc(i, k)

	errorStrings := []string{}
	for _, e := range errors {
		errorStrings = append(errorStrings, e.Error())
	}

	if len(errors) > 0 {
		return false, fmt.Errorf(strings.Join(errorStrings, "\n"))
	}

	return true, nil
}

func getVirtualNetworkGatewayProperties(d *pluginsdk.ResourceData) (*network.VirtualNetworkGatewayPropertiesFormat, error) {
	gatewayType := network.VirtualNetworkGatewayType(d.Get("type").(string))
	vpnType := network.VpnType(d.Get("vpn_type").(string))
	enableBgp := d.Get("enable_bgp").(bool)
	activeActive := d.Get("active_active").(bool)

	props := &network.VirtualNetworkGatewayPropertiesFormat{
		GatewayType:      gatewayType,
		VpnType:          vpnType,
		EnableBgp:        &enableBgp,
		ActiveActive:     &activeActive,
		Sku:              expandVirtualNetworkGatewaySku(d),
		IPConfigurations: expandVirtualNetworkGatewayIPConfigurations(d),
	}

	if gatewayDefaultSiteID := d.Get("default_local_network_gateway_id").(string); gatewayDefaultSiteID != "" {
		props.GatewayDefaultSite = &network.SubResource{
			ID: &gatewayDefaultSiteID,
		}
	}

	if _, ok := d.GetOk("vpn_client_configuration"); ok {
		props.VpnClientConfiguration = expandVirtualNetworkGatewayVpnClientConfig(d)
	}

	if _, ok := d.GetOk("bgp_settings"); ok {
		bgpSettings, err := expandVirtualNetworkGatewayBgpSettings(d)
		if err != nil {
			return nil, err
		}
		props.BgpSettings = bgpSettings
	}

	// Sku validation for policy-based VPN gateways
	if props.GatewayType == network.VirtualNetworkGatewayTypeVpn && props.VpnType == network.PolicyBased {
		if ok, err := evaluateSchemaValidateFunc(string(props.Sku.Name), "sku", validateVirtualNetworkGatewayPolicyBasedVpnSku()); !ok {
			return nil, err
		}
	}

	// Sku validation for ExpressRoute gateways
	if props.GatewayType == network.VirtualNetworkGatewayTypeExpressRoute {
		if ok, err := evaluateSchemaValidateFunc(string(props.Sku.Name), "sku", validateVirtualNetworkGatewayExpressRouteSku()); !ok {
			return nil, err
		}
	}

	return props, nil
}

func expandVirtualNetworkGatewayBgpSettings(d *pluginsdk.ResourceData) (*network.BgpSettings, error) {
	bgpSets := d.Get("bgp_settings").([]interface{})
	if len(bgpSets) == 0 {
		return nil, fmt.Errorf("bgp_settings is nil")
	}

	bgp := bgpSets[0].(map[string]interface{})

	asn := int64(bgp["asn"].(int))
	peeringAddress := bgp["peering_address"].(string)
	peerWeight := int32(bgp["peer_weight"].(int))

	return &network.BgpSettings{
		Asn:               &asn,
		BgpPeeringAddress: &peeringAddress,
		PeerWeight:        &peerWeight,
	}, nil
}

func expandVirtualNetworkGatewayIPConfigurations(d *pluginsdk.ResourceData) *[]network.VirtualNetworkGatewayIPConfiguration {
	configs := d.Get("ip_configuration").([]interface{})
	ipConfigs := make([]network.VirtualNetworkGatewayIPConfiguration, 0, len(configs))

	for _, c := range configs {
		conf := c.(map[string]interface{})

		name := conf["name"].(string)
		privateIPAllocMethod := network.IPAllocationMethod(conf["private_ip_address_allocation"].(string))

		props := &network.VirtualNetworkGatewayIPConfigurationPropertiesFormat{
			PrivateIPAllocationMethod: privateIPAllocMethod,
		}

		if subnetID := conf["subnet_id"].(string); subnetID != "" {
			props.Subnet = &network.SubResource{
				ID: &subnetID,
			}
		}

		if publicIP := conf["public_ip_address_id"].(string); publicIP != "" {
			props.PublicIPAddress = &network.SubResource{
				ID: &publicIP,
			}
		}

		ipConfig := network.VirtualNetworkGatewayIPConfiguration{
			Name: &name,
			VirtualNetworkGatewayIPConfigurationPropertiesFormat: props,
		}

		ipConfigs = append(ipConfigs, ipConfig)
	}

	return &ipConfigs
}

func expandVirtualNetworkGatewayVpnClientConfig(d *pluginsdk.ResourceData) *network.VpnClientConfiguration {
	configSets := d.Get("vpn_client_configuration").([]interface{})
	conf := configSets[0].(map[string]interface{})

	confAddresses := conf["address_space"].([]interface{})
	addresses := make([]string, 0, len(confAddresses))
	for _, addr := range confAddresses {
		addresses = append(addresses, addr.(string))
	}

	var rootCerts []network.VpnClientRootCertificate
	for _, rootCertSet := range conf["root_certificate"].(*pluginsdk.Set).List() {
		rootCert := rootCertSet.(map[string]interface{})
		name := rootCert["name"].(string)
		publicCertData := rootCert["public_cert_data"].(string)
		r := network.VpnClientRootCertificate{
			Name: &name,
			VpnClientRootCertificatePropertiesFormat: &network.VpnClientRootCertificatePropertiesFormat{
				PublicCertData: &publicCertData,
			},
		}
		rootCerts = append(rootCerts, r)
	}

	var revokedCerts []network.VpnClientRevokedCertificate
	for _, revokedCertSet := range conf["revoked_certificate"].(*pluginsdk.Set).List() {
		revokedCert := revokedCertSet.(map[string]interface{})
		name := revokedCert["name"].(string)
		thumbprint := revokedCert["thumbprint"].(string)
		r := network.VpnClientRevokedCertificate{
			Name: &name,
			VpnClientRevokedCertificatePropertiesFormat: &network.VpnClientRevokedCertificatePropertiesFormat{
				Thumbprint: &thumbprint,
			},
		}
		revokedCerts = append(revokedCerts, r)
	}

	var vpnClientProtocols []network.VpnClientProtocol
	for _, vpnClientProtocol := range conf["vpn_client_protocols"].(*pluginsdk.Set).List() {
		p := network.VpnClientProtocol(vpnClientProtocol.(string))
		vpnClientProtocols = append(vpnClientProtocols, p)
	}

	confRadiusServerAddress := conf["radius_server_address"].(string)
	confRadiusServerSecret := conf["radius_server_secret"].(string)

	return &network.VpnClientConfiguration{
		VpnClientAddressPool: &network.AddressSpace{
			AddressPrefixes: &addresses,
		},
		VpnClientRootCertificates:    &rootCerts,
		VpnClientRevokedCertificates: &revokedCerts,
		VpnClientProtocols:           &vpnClientProtocols,
		RadiusServerAddress:          &confRadiusServerAddress,
		RadiusServerSecret:           &confRadiusServerSecret,
	}
}

func expandVirtualNetworkGatewaySku(d *pluginsdk.ResourceData) *network.VirtualNetworkGatewaySku {
	sku := d.Get("sku").(string)

	return &network.VirtualNetworkGatewaySku{
		Name: network.VirtualNetworkGatewaySkuName(sku),
		Tier: network.VirtualNetworkGatewaySkuTier(sku),
	}
}

func flattenVirtualNetworkGatewayBgpSettings(settings *network.BgpSettings) []interface{} {
	output := make([]interface{}, 0)

	if settings != nil {
		flat := make(map[string]interface{})

		if asn := settings.Asn; asn != nil {
			flat["asn"] = int(*asn)
		}
		if address := settings.BgpPeeringAddress; address != nil {
			flat["peering_address"] = *address
		}
		if weight := settings.PeerWeight; weight != nil {
			flat["peer_weight"] = int(*weight)
		}

		output = append(output, flat)
	}

	return output
}

func flattenVirtualNetworkGatewayIPConfigurations(ipConfigs *[]network.VirtualNetworkGatewayIPConfiguration) []interface{} {
	flat := make([]interface{}, 0)

	if ipConfigs != nil {
		for _, cfg := range *ipConfigs {
			props := cfg.VirtualNetworkGatewayIPConfigurationPropertiesFormat
			v := make(map[string]interface{})

			if name := cfg.Name; name != nil {
				v["name"] = *name
			}
			v["private_ip_address_allocation"] = string(props.PrivateIPAllocationMethod)

			if subnet := props.Subnet; subnet != nil {
				if id := subnet.ID; id != nil {
					v["subnet_id"] = *id
				}
			}

			if pip := props.PublicIPAddress; pip != nil {
				if id := pip.ID; id != nil {
					v["public_ip_address_id"] = *id
				}
			}

			flat = append(flat, v)
		}
	}

	return flat
}

func flattenVirtualNetworkGatewayVpnClientConfig(cfg *network.VpnClientConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}
	flat := make(map[string]interface{})

	if pool := cfg.VpnClientAddressPool; pool != nil {
		flat["address_space"] = utils.FlattenStringSlice(pool.AddressPrefixes)
	} else {
		flat["address_space"] = []interface{}{}
	}

	rootCerts := make([]interface{}, 0)
	if certs := cfg.VpnClientRootCertificates; certs != nil {
		for _, cert := range *certs {
			v := map[string]interface{}{
				"name":             *cert.Name,
				"public_cert_data": *cert.VpnClientRootCertificatePropertiesFormat.PublicCertData,
			}
			rootCerts = append(rootCerts, v)
		}
	}
	flat["root_certificate"] = pluginsdk.NewSet(hashVirtualNetworkGatewayRootCert, rootCerts)

	revokedCerts := make([]interface{}, 0)
	if certs := cfg.VpnClientRevokedCertificates; certs != nil {
		for _, cert := range *certs {
			v := map[string]interface{}{
				"name":       *cert.Name,
				"thumbprint": *cert.VpnClientRevokedCertificatePropertiesFormat.Thumbprint,
			}
			revokedCerts = append(revokedCerts, v)
		}
	}
	flat["revoked_certificate"] = pluginsdk.NewSet(hashVirtualNetworkGatewayRevokedCert, revokedCerts)

	vpnClientProtocols := &pluginsdk.Set{F: pluginsdk.HashString}
	if vpnProtocols := cfg.VpnClientProtocols; vpnProtocols != nil {
		for _, protocol := range *vpnProtocols {
			vpnClientProtocols.Add(string(protocol))
		}
	}
	flat["vpn_client_protocols"] = vpnClientProtocols

	if v := cfg.RadiusServerAddress; v != nil {
		flat["radius_server_address"] = *v
	}

	if v := cfg.RadiusServerSecret; v != nil {
		flat["radius_server_secret"] = *v
	}

	return []interface{}{flat}
}

func hashVirtualNetworkGatewayRootCert(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["public_cert_data"].(string)))

	return pluginsdk.HashString(buf.String())
}

func hashVirtualNetworkGatewayRevokedCert(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["thumbprint"].(string)))

	return pluginsdk.HashString(buf.String())
}

func validateVirtualNetworkGatewayPolicyBasedVpnSku() pluginsdk.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		string(network.VirtualNetworkGatewaySkuTierBasic),
	}, true)
}

func validateVirtualNetworkGatewayRouteBasedVpnSkuGeneration1() pluginsdk.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		string(network.VirtualNetworkGatewaySkuTierBasic),
		string(network.VirtualNetworkGatewaySkuTierStandard),
		string(network.VirtualNetworkGatewaySkuTierHighPerformance),
		string(network.VirtualNetworkGatewaySkuNameVpnGw1),
		string(network.VirtualNetworkGatewaySkuNameVpnGw2),
		string(network.VirtualNetworkGatewaySkuNameVpnGw3),
	}, true)
}

func validateVirtualNetworkGatewayRouteBasedVpnSkuGeneration2() pluginsdk.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		string(network.VirtualNetworkGatewaySkuNameVpnGw2),
		string(network.VirtualNetworkGatewaySkuNameVpnGw3),
	}, true)
}

func validateVirtualNetworkGatewayExpressRouteSku() pluginsdk.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		string(network.VirtualNetworkGatewaySkuTierStandard),
		string(network.VirtualNetworkGatewaySkuTierHighPerformance),
		string(network.VirtualNetworkGatewaySkuTierUltraPerformance),
	}, true)
}
