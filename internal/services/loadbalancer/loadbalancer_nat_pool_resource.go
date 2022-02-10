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
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/state"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func loadBalancerNatPool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: loadBalancerNatPoolCreateUpdate,
		Read:   loadBalancerNatPoolRead,
		Update: loadBalancerNatPoolCreateUpdate,
		Delete: loadBalancerNatPoolDelete,

		Importer: loadBalancerSubResourceImporter(func(input string) (*parse.LoadBalancerId, error) {
			id, err := parse.LoadBalancerInboundNatPoolID(input)
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
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"loadbalancer_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"protocol": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				StateFunc:        state.IgnoreCase,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.TransportProtocolAll),
					string(network.TransportProtocolTCP),
					string(network.TransportProtocolUDP),
				}, true),
			},

			"frontend_port_start": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},

			"frontend_port_end": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},

			"backend_port": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
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
		},
	}
}

func loadBalancerNatPoolCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return fmt.Errorf("parsing Load Balancer Name and Group: %+v", err)
	}

	id := parse.NewLoadBalancerInboundNatPoolID(subscriptionId, loadBalancerId.ResourceGroup, loadBalancerId.Name, d.Get("name").(string))

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
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}
	newNatPool, err := expandazurestackLoadBalancerNatPool(d, &loadBalancer)
	if err != nil {
		return fmt.Errorf("expanding NAT Pool: %+v", err)
	}

	natPools := append(*loadBalancer.LoadBalancerPropertiesFormat.InboundNatPools, *newNatPool)

	existingNatPool, existingNatPoolIndex, exists := FindLoadBalancerNatPoolByName(&loadBalancer, id.InboundNatPoolName)
	if exists {
		if id.InboundNatPoolName == *existingNatPool.Name {
			if d.IsNewResource() {
				return tf.ImportAsExistsError("azurestack_lb_nat_pool", *existingNatPool.ID)
			}

			// this pool is being updated/reapplied remove old copy from the slice
			natPools = append(natPools[:existingNatPoolIndex], natPools[existingNatPoolIndex+1:]...)
		}
	}

	loadBalancer.LoadBalancerPropertiesFormat.InboundNatPools = &natPools

	future, err := client.CreateOrUpdate(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, loadBalancer)
	if err != nil {
		return fmt.Errorf("updating Load Balancer %q (Resource Group %q) for Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the update of Load Balancer %q (Resource Group %q) for Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return loadBalancerNatPoolRead(d, meta)
}

func loadBalancerNatPoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerInboundNatPoolID(d.Id())
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
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}

	config, _, exists := FindLoadBalancerNatPoolByName(&loadBalancer, id.InboundNatPoolName)
	if !exists {
		d.SetId("")
		log.Printf("[INFO] Load Balancer Nat Pool %q not found. Removing from state", id.InboundNatPoolName)
		return nil
	}

	d.Set("name", config.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := config.InboundNatPoolPropertiesFormat; props != nil {
		backendPort := 0
		if props.BackendPort != nil {
			backendPort = int(*props.BackendPort)
		}
		d.Set("backend_port", backendPort)

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
		d.Set("frontend_ip_configuration_id", frontendIPConfigID)
		d.Set("frontend_ip_configuration_name", frontendIPConfigName)

		frontendPortRangeEnd := 0
		if props.FrontendPortRangeEnd != nil {
			frontendPortRangeEnd = int(*props.FrontendPortRangeEnd)
		}
		d.Set("frontend_port_end", frontendPortRangeEnd)

		frontendPortRangeStart := 0
		if props.FrontendPortRangeStart != nil {
			frontendPortRangeStart = int(*props.FrontendPortRangeStart)
		}
		d.Set("frontend_port_start", frontendPortRangeStart)
		d.Set("protocol", string(props.Protocol))
	}

	return nil
}

func loadBalancerNatPoolDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerInboundNatPoolID(d.Id())
	if err != nil {
		return err
	}

	loadBalancerId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)
	loadBalancerID := loadBalancerId.ID()
	locks.ByID(loadBalancerID)
	defer locks.UnlockByID(loadBalancerID)

	loadBalancer, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(loadBalancer.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for deletion of Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}

	_, index, exists := FindLoadBalancerNatPoolByName(&loadBalancer, id.InboundNatPoolName)
	if !exists {
		return nil
	}

	natPools := *loadBalancer.LoadBalancerPropertiesFormat.InboundNatPools
	natPools = append(natPools[:index], natPools[index+1:]...)
	loadBalancer.LoadBalancerPropertiesFormat.InboundNatPools = &natPools

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, loadBalancer)
	if err != nil {
		return fmt.Errorf("updating Load Balancer %q (Resource Group %q) for Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for update of the Load Balancer %q (Resource Group %q) for Nat Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.InboundNatPoolName, err)
	}

	return nil
}

func expandazurestackLoadBalancerNatPool(d *pluginsdk.ResourceData, lb *network.LoadBalancer) (*network.InboundNatPool, error) {
	properties := network.InboundNatPoolPropertiesFormat{
		Protocol:               network.TransportProtocol(d.Get("protocol").(string)),
		FrontendPortRangeStart: utils.Int32(int32(d.Get("frontend_port_start").(int))),
		FrontendPortRangeEnd:   utils.Int32(int32(d.Get("frontend_port_end").(int))),
		BackendPort:            utils.Int32(int32(d.Get("backend_port").(int))),
	}

	if v := d.Get("frontend_ip_configuration_name").(string); v != "" {
		rule, exists := FindLoadBalancerFrontEndIpConfigurationByName(lb, v)
		if !exists {
			return nil, fmt.Errorf("[ERROR] Cannot find FrontEnd IP Configuration with the name %s", v)
		}

		properties.FrontendIPConfiguration = &network.SubResource{
			ID: rule.ID,
		}
	}

	return &network.InboundNatPool{
		Name:                           pointer.FromString(d.Get("name").(string)),
		InboundNatPoolPropertiesFormat: &properties,
	}, nil
}
