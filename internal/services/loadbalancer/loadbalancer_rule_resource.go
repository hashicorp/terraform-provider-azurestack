package loadbalancer

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/locks"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func loadBalancerRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceArmLoadBalancerRuleCreateUpdate,
		Read:   loadBalancerRuleRead,
		Update: resourceArmLoadBalancerRuleCreateUpdate,
		Delete: loadBalancerRuleDelete,

		Importer: loadBalancerSubResourceImporter(func(input string) (*parse.LoadBalancerId, error) {
			id, err := parse.LoadBalancingRuleID(input)
			if err != nil {
				return nil, err
			}

			lbId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)
			return &lbId, nil
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
				ValidateFunc: validate.RuleName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"loadbalancer_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"frontend_ip_configuration_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"frontend_ip_configuration_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"backend_address_pool_id": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},

			"protocol": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.TransportProtocolAll),
					string(network.TransportProtocolTCP),
					string(network.TransportProtocolUDP),
				}, true),
			},

			"frontend_port": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},

			"backend_port": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},

			"probe_id": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},

			"enable_floating_ip": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"disable_outbound_snat": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"idle_timeout_in_minutes": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(4, 30),
			},

			"load_distribution": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceArmLoadBalancerRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewLoadBalancingRuleID(subscriptionId, loadBalancerId.ResourceGroup, loadBalancerId.Name, d.Get("name").(string))

	loadBalancerID := loadBalancerId.ID()
	locks.ByID(loadBalancerID)
	defer locks.UnlockByID(loadBalancerID)

	loadBalancer, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			log.Printf("[INFO] Load Balancer %q not found. Removing from state", id.LoadBalancerName)
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.Name, err)
	}

	newLbRule, err := expandazurestackLoadBalancerRule(d, &loadBalancer)
	if err != nil {
		return fmt.Errorf("expanding Load Balancer Rule: %+v", err)
	}

	lbRules := append(*loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules, *newLbRule)

	existingRule, existingRuleIndex, exists := FindLoadBalancerRuleByName(&loadBalancer, id.Name)
	if exists {
		if id.Name == *existingRule.Name {
			if d.IsNewResource() {
				return tf.ImportAsExistsError("azurestack_lb_rule", *existingRule.ID)
			}

			// this rule is being updated/reapplied remove old copy from the slice
			lbRules = append(lbRules[:existingRuleIndex], lbRules[existingRuleIndex+1:]...)
		}
	}

	loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules = &lbRules

	future, err := client.CreateOrUpdate(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, loadBalancer)
	if err != nil {
		return fmt.Errorf("updating Loadbalancer %q (resource group %q) for Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.Name, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of Load Balancer %q (resource group %q) for Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.Name, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return loadBalancerRuleRead(d, meta)
}

func loadBalancerRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancingRuleID(d.Id())
	if err != nil {
		return err
	}

	loadBalancer, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			log.Printf("[INFO] Load Balancer %q not found. Removing from state", id.LoadBalancerName)
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.Name, err)
	}

	config, _, exists := FindLoadBalancerRuleByName(&loadBalancer, id.Name)
	if !exists {
		d.SetId("")
		log.Printf("[INFO] Load Balancer Rule %q not found. Removing from state", id.Name)
		return nil
	}

	d.Set("name", config.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := config.LoadBalancingRulePropertiesFormat; props != nil {
		d.Set("disable_outbound_snat", props.DisableOutboundSnat)
		d.Set("enable_floating_ip", props.EnableFloatingIP)
		d.Set("protocol", string(props.Protocol))

		backendPort := 0
		if props.BackendPort != nil {
			backendPort = int(*props.BackendPort)
		}
		d.Set("backend_port", backendPort)

		// The backendAddressPools is designed for Gateway LB, while the backendAddressPool is designed for other skus.
		// Thought currently the API returns both, but for the sake of stability, we do use different fields here depending on the LB sku.
		var (
			backendAddressPoolId string
		)

		if props.BackendAddressPool != nil && props.BackendAddressPool.ID != nil {
			backendAddressPoolId = *props.BackendAddressPool.ID
		}

		d.Set("backend_address_pool_id", backendAddressPoolId)

		frontendIPConfigName := ""
		frontendIPConfigID := ""
		if props.FrontendIPConfiguration != nil && props.FrontendIPConfiguration.ID != nil {
			feid, err := parse.LoadBalancerFrontendIpConfigurationID(*props.FrontendIPConfiguration.ID)
			if err != nil {
				return err
			}

			frontendIPConfigName = feid.FrontendIPConfigurationName
			frontendIPConfigID = feid.ID()
		}
		d.Set("frontend_ip_configuration_name", frontendIPConfigName)
		d.Set("frontend_ip_configuration_id", frontendIPConfigID)

		frontendPort := 0
		if props.FrontendPort != nil {
			frontendPort = int(*props.FrontendPort)
		}
		d.Set("frontend_port", frontendPort)

		idleTimeoutInMinutes := 0
		if props.IdleTimeoutInMinutes != nil {
			idleTimeoutInMinutes = int(*props.IdleTimeoutInMinutes)
		}
		d.Set("idle_timeout_in_minutes", idleTimeoutInMinutes)

		loadDistribution := ""
		if props.LoadDistribution != "" {
			loadDistribution = string(props.LoadDistribution)
		}
		d.Set("load_distribution", loadDistribution)

		probeId := ""
		if props.Probe != nil && props.Probe.ID != nil {
			probeId = *props.Probe.ID
		}
		d.Set("probe_id", probeId)
	}

	return nil
}

func loadBalancerRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancingRuleID(d.Id())
	if err != nil {
		return err
	}

	loadBalancerId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)
	loadBalancerIDRaw := loadBalancerId.ID()
	locks.ByID(loadBalancerIDRaw)
	defer locks.UnlockByID(loadBalancerIDRaw)

	loadBalancer, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Rule %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.Name, err)
	}

	_, index, exists := FindLoadBalancerRuleByName(&loadBalancer, d.Get("name").(string))
	if !exists {
		return nil
	}

	lbRules := *loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules
	lbRules = append(lbRules[:index], lbRules[index+1:]...)
	loadBalancer.LoadBalancerPropertiesFormat.LoadBalancingRules = &lbRules

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, loadBalancer)
	if err != nil {
		return fmt.Errorf("Creating/Updating Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	return nil
}

func expandazurestackLoadBalancerRule(d *pluginsdk.ResourceData, lb *network.LoadBalancer) (*network.LoadBalancingRule, error) {
	properties := network.LoadBalancingRulePropertiesFormat{
		Protocol:            network.TransportProtocol(d.Get("protocol").(string)),
		FrontendPort:        utils.Int32(int32(d.Get("frontend_port").(int))),
		BackendPort:         utils.Int32(int32(d.Get("backend_port").(int))),
		EnableFloatingIP:    pointer.FromBool(d.Get("enable_floating_ip").(bool)),
		DisableOutboundSnat: pointer.FromBool(d.Get("disable_outbound_snat").(bool)),
	}

	if v, ok := d.GetOk("idle_timeout_in_minutes"); ok {
		properties.IdleTimeoutInMinutes = utils.Int32(int32(v.(int)))
	}

	if v := d.Get("load_distribution").(string); v != "" {
		properties.LoadDistribution = network.LoadDistribution(v)
	}

	// TODO: ensure these ID's are consistent
	if v := d.Get("frontend_ip_configuration_name").(string); v != "" {
		rule, exists := FindLoadBalancerFrontEndIpConfigurationByName(lb, v)
		if !exists {
			return nil, fmt.Errorf("[ERROR] Cannot find FrontEnd IP Configuration with the name %s", v)
		}

		properties.FrontendIPConfiguration = &network.SubResource{
			ID: rule.ID,
		}
	}

	if v := d.Get("backend_address_pool_id").(string); v != "" {
		properties.BackendAddressPool = &network.SubResource{
			ID: &v,
		}
	}

	if v := d.Get("probe_id").(string); v != "" {
		properties.Probe = &network.SubResource{
			ID: &v,
		}
	}

	return &network.LoadBalancingRule{
		Name:                              pointer.FromString(d.Get("name").(string)),
		LoadBalancingRulePropertiesFormat: &properties,
	}, nil
}
