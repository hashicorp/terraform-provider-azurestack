package loadbalancer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/resourceid"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/zones"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/state"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func loadBalancer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: loadBalancerCreateUpdate,
		Read:   loadBalancerRead,
		Update: loadBalancerCreateUpdate,
		Delete: loadBalancerDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.LoadBalancerID(id)
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
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"frontend_ip_configuration": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"name": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"availability_zone": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							// Default:  "Zone-Redundant",
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"No-Zone",
								"1",
								"2",
								"3",
								"Zone-Redundant",
							}, false),
						},

						"subnet_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
						},

						"private_ip_address": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.Any(
								validation.IsIPAddress,
								validation.StringIsEmpty,
							),
						},

						"public_ip_address_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: resourceid.ValidateResourceIDOrEmpty,
						},

						"private_ip_address_allocation": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(network.Dynamic),
								string(network.Static),
							}, true),
							StateFunc:        state.IgnoreCase,
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"load_balancer_rules": {
							Type:     pluginsdk.TypeSet,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							Set: pluginsdk.HashString,
						},

						"inbound_nat_rules": {
							Type:     pluginsdk.TypeSet,
							Computed: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
							Set: pluginsdk.HashString,
						},

						// TODO - 3.0 make Computed only
						"zones": {
							Type:       pluginsdk.TypeList,
							Optional:   true,
							Computed:   true,
							Deprecated: "This property has been deprecated in favour of `availability_zone` due to a breaking behavioural change in Azure: https://azure.microsoft.com/en-us/updates/zone-behavior-change/",
							MaxItems:   1,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeString,
								ValidateFunc: validation.StringIsNotEmpty,
							},
						},

						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"private_ip_address": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
			"private_ip_addresses": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},

			"tags": tags.Schema(),
		},

		CustomizeDiff: pluginsdk.CustomizeDiffShim(func(ctx context.Context, d *pluginsdk.ResourceDiff, v interface{}) error {
			if ok := d.HasChange("frontend_ip_configuration"); ok {
				configs := d.Get("frontend_ip_configuration").([]interface{})

				for index := range configs {
					if d.HasChange(fmt.Sprintf("frontend_ip_configuration.%d.availability_zone", index)) && !d.HasChange(fmt.Sprintf("frontend_ip_configuration.%d.name", index)) {
						return fmt.Errorf("in place change of the `frontend_ip_configuration.%[1]d.availability_zone` is not allowed. It is allowed to do this while also changing `frontend_ip_configuration.%[1]d.name`", index)
					}

					// TODO - Remove in 3.0
					if d.HasChange(fmt.Sprintf("frontend_ip_configuration.%d.zones", index)) && !d.HasChange(fmt.Sprintf("frontend_ip_configuration.%d.name", index)) {
						return fmt.Errorf("in place change of the `frontend_ip_configuration.%[1]d.zones` is not allowed. It is allowed to do this while also changing `frontend_ip_configuration.%[1]d.name`", index)
					}
				}
			}

			return nil
		}),
	}
}

func loadBalancerCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM Load Balancer creation.")

	id := parse.NewLoadBalancerID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_lb", id.ID())
		}
	}

	properties := network.LoadBalancerPropertiesFormat{}

	if _, ok := d.GetOk("frontend_ip_configuration"); ok {
		frontendIPConfigurations, err := expandazurestackLoadBalancerFrontendIpConfigurations(d)
		if err != nil {
			return err
		}
		properties.FrontendIPConfigurations = frontendIPConfigurations
	}

	loadBalancer := network.LoadBalancer{
		Name:                         utils.String(id.Name),
		Location:                     utils.String(location.Normalize(d.Get("location").(string))),
		Tags:                         tags.Expand(d.Get("tags").(map[string]interface{})),
		LoadBalancerPropertiesFormat: &properties,
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, loadBalancer)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %s: %+v", id, err)
	}

	d.SetId(id.ID())

	return loadBalancerRead(d, meta)
}

func loadBalancerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))

	if props := resp.LoadBalancerPropertiesFormat; props != nil {
		if feipConfigs := props.FrontendIPConfigurations; feipConfigs != nil {
			if err := d.Set("frontend_ip_configuration", flattenLoadBalancerFrontendIpConfiguration(feipConfigs)); err != nil {
				return fmt.Errorf("flattening `frontend_ip_configuration`: %+v", err)
			}

			privateIpAddress := ""
			privateIpAddresses := make([]string, 0)
			for _, config := range *feipConfigs {
				if feipProps := config.FrontendIPConfigurationPropertiesFormat; feipProps != nil {
					if ip := feipProps.PrivateIPAddress; ip != nil {
						if privateIpAddress == "" {
							privateIpAddress = *feipProps.PrivateIPAddress
						}

						privateIpAddresses = append(privateIpAddresses, *feipProps.PrivateIPAddress)
					}
				}
			}

			d.Set("private_ip_address", privateIpAddress)
			d.Set("private_ip_addresses", privateIpAddresses)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func loadBalancerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return nil
}

func expandazurestackLoadBalancerFrontendIpConfigurations(d *pluginsdk.ResourceData) (*[]network.FrontendIPConfiguration, error) {
	configs := d.Get("frontend_ip_configuration").([]interface{})
	frontEndConfigs := make([]network.FrontendIPConfiguration, 0, len(configs))
	sku := d.Get("sku").(string)

	for index, configRaw := range configs {
		data := configRaw.(map[string]interface{})

		privateIpAllocationMethod := data["private_ip_address_allocation"].(string)
		properties := network.FrontendIPConfigurationPropertiesFormat{
			PrivateIPAllocationMethod: network.IPAllocationMethod(privateIpAllocationMethod),
		}

		if v := data["private_ip_address"].(string); v != "" {
			properties.PrivateIPAddress = &v
		}

		subnetSet := false
		if v := data["public_ip_address_id"].(string); v != "" {
			properties.PublicIPAddress = &network.PublicIPAddress{
				ID: &v,
			}
		}

		if v := data["subnet_id"].(string); v != "" {
			subnetSet = true
			properties.Subnet = &network.Subnet{
				ID: &v,
			}
		}

		name := data["name"].(string)
		// TODO - get zone list for each location by Resource API, instead of hardcode
		z := &[]string{"1", "2"}
		zonesSet := false
		// TODO - Remove in 3.0
		if deprecatedZonesRaw, ok := d.GetOk(fmt.Sprintf("frontend_ip_configuration.%d.zones", index)); ok {
			zonesSet = true
			deprecatedZones := zones.ExpandZones(deprecatedZonesRaw.([]interface{}))
			if deprecatedZones != nil {
				z = deprecatedZones
			}
		}

		if availabilityZones, ok := d.GetOk(fmt.Sprintf("frontend_ip_configuration.%d.availability_zone", index)); ok {
			zonesSet = true
			switch availabilityZones.(string) {
			case "1", "2", "3":
				z = &[]string{availabilityZones.(string)}
			case "Zone-Redundant":
				z = &[]string{"1", "2"}
			case "No-Zone":
				z = &[]string{}
			}
		}
		if strings.EqualFold(sku, string(network.LoadBalancerSkuNameBasic)) {
			if zonesSet && len(*z) > 0 {
				return nil, fmt.Errorf("Availability Zones are not available on the `Basic` SKU")
			}
			z = &[]string{}
		} else if !subnetSet {
			if zonesSet && len(*z) > 0 {
				return nil, fmt.Errorf("Networking supports zones only for frontendIpconfigurations which reference a subnet.")
			}
			z = &[]string{}
		}
		frontEndConfig := network.FrontendIPConfiguration{
			Name:                                    &name,
			FrontendIPConfigurationPropertiesFormat: &properties,
			Zones:                                   z,
		}

		frontEndConfigs = append(frontEndConfigs, frontEndConfig)
	}

	return &frontEndConfigs, nil
}

func flattenLoadBalancerFrontendIpConfiguration(ipConfigs *[]network.FrontendIPConfiguration) []interface{} {
	result := make([]interface{}, 0)
	if ipConfigs == nil {
		return result
	}

	for _, config := range *ipConfigs {
		ipConfig := make(map[string]interface{})

		if config.Name != nil {
			ipConfig["name"] = *config.Name
		}

		if config.ID != nil {
			ipConfig["id"] = *config.ID
		}

		availabilityZones := "No-Zone"
		zonesDeprecated := make([]string, 0)
		if config.Zones != nil {
			if len(*config.Zones) > 1 {
				availabilityZones = "Zone-Redundant"
			}
			if len(*config.Zones) == 1 {
				zones := *config.Zones
				availabilityZones = zones[0]
				zonesDeprecated = zones
			}
		}
		ipConfig["availability_zone"] = availabilityZones
		ipConfig["zones"] = zonesDeprecated

		if props := config.FrontendIPConfigurationPropertiesFormat; props != nil {
			ipConfig["private_ip_address_allocation"] = string(props.PrivateIPAllocationMethod)

			if subnet := props.Subnet; subnet != nil {
				ipConfig["subnet_id"] = *subnet.ID
			}

			if pip := props.PrivateIPAddress; pip != nil {
				ipConfig["private_ip_address"] = *pip
			}

			if pip := props.PublicIPAddress; pip != nil {
				ipConfig["public_ip_address_id"] = *pip.ID
			}

			loadBalancingRules := make([]interface{}, 0)
			if rules := props.LoadBalancingRules; rules != nil {
				for _, rule := range *rules {
					loadBalancingRules = append(loadBalancingRules, *rule.ID)
				}
			}
			ipConfig["load_balancer_rules"] = pluginsdk.NewSet(pluginsdk.HashString, loadBalancingRules)

			inboundNatRules := make([]interface{}, 0)
			if rules := props.InboundNatRules; rules != nil {
				for _, rule := range *rules {
					inboundNatRules = append(inboundNatRules, *rule.ID)
				}
			}
			ipConfig["inbound_nat_rules"] = pluginsdk.NewSet(pluginsdk.HashString, inboundNatRules)
		}

		result = append(result, ipConfig)
	}
	return result
}
