package loadbalancer

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/locks"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

var backendAddressPoolResourceName = "azurestack_lb_backend_address_pool"

func loadBalancerBackendAddressPool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: loadBalancerBackendAddressPoolCreateUpdate,
		Update: loadBalancerBackendAddressPoolCreateUpdate, // TODO: remove in 3.0 since all fields are ForceNew
		Read:   loadBalancerBackendAddressPoolRead,
		Delete: loadBalancerBackendAddressPoolDelete,

		Importer: loadBalancerSubResourceImporter(func(input string) (*parse.LoadBalancerId, error) {
			id, err := parse.LoadBalancerBackendAddressPoolID(input)
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

		Schema: func() map[string]*pluginsdk.Schema {
			s := map[string]*pluginsdk.Schema{
				"name": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				// TODO 3.0: remove this as it can be inferred from "loadbalancer_id"
				"resource_group_name": commonschema.ResourceGroupNameDeprecatedComputed(),

				"loadbalancer_id": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validate.LoadBalancerID,
				},

				"backend_ip_configurations": {
					Type:     pluginsdk.TypeList,
					Computed: true,
					Elem: &pluginsdk.Schema{
						Type: pluginsdk.TypeString,
					},
				},

				"load_balancing_rules": {
					Type:     pluginsdk.TypeList,
					Computed: true,
					Elem: &pluginsdk.Schema{
						Type: pluginsdk.TypeString,
					},
				},
			}

			return s
		}(),
	}
}

func loadBalancerBackendAddressPoolCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancerBackendAddressPoolsClient
	lbClient := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return fmt.Errorf("parsing Load Balancer Name and Group: %+v", err)
	}

	name := d.Get("name").(string)
	id := parse.NewLoadBalancerBackendAddressPoolID(loadBalancerId.SubscriptionId, loadBalancerId.ResourceGroup, loadBalancerId.Name, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Load Balancer Backend Address Pool %q: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_lb_backend_address_pool", id.ID())
		}
	}

	locks.ByName(name, backendAddressPoolResourceName)
	defer locks.UnlockByName(name, backendAddressPoolResourceName)

	locks.ByID(loadBalancerId.ID())
	defer locks.UnlockByID(loadBalancerId.ID())

	lb, err := lbClient.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(lb.Response) {
			return fmt.Errorf("Load Balancer %q for Backend Address Pool %q was not found", loadBalancerId, id)
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q for Backend Address Pool %q: %+v", loadBalancerId, id, err)
	}

	param := network.BackendAddressPool{
		Name: &id.BackendAddressPoolName,
	}

	// Insert this BAP and update the LB since the dedicated BAP endpoint doesn't work for the Basic sku.
	backendAddressPools := append(*lb.LoadBalancerPropertiesFormat.BackendAddressPools, param)
	_, existingPoolIndex, exists := FindLoadBalancerBackEndAddressPoolByName(&lb, id.BackendAddressPoolName)
	if exists {
		// this pool is being updated/reapplied remove the old copy from the slice
		backendAddressPools = append(backendAddressPools[:existingPoolIndex], backendAddressPools[existingPoolIndex+1:]...)
	}

	lb.LoadBalancerPropertiesFormat.BackendAddressPools = &backendAddressPools

	future, err := lbClient.CreateOrUpdate(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, lb)
	if err != nil {
		return fmt.Errorf("updating Load Balancer %q for Backend Address Pool %q: %+v", loadBalancerId, id, err)
	}

	if err = future.WaitForCompletionRef(ctx, lbClient.Client); err != nil {
		return fmt.Errorf("waiting for update of Load Balancer %q for Backend Address Pool %q: %+v", loadBalancerId, id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return loadBalancerBackendAddressPoolRead(d, meta)
}

func loadBalancerBackendAddressPoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancerBackendAddressPoolsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerBackendAddressPoolID(d.Id())
	if err != nil {
		return err
	}

	lbId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)

	resp, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			log.Printf("[INFO] Load Balancer Backend Address Pool %q not found - removing from state", id)
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer Backend Address Pool %q: %+v", id, err)
	}

	d.Set("name", id.BackendAddressPoolName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("loadbalancer_id", lbId.ID())

	if props := resp.BackendAddressPoolPropertiesFormat; props != nil {
		var backendIPConfigurations []string
		if configs := props.BackendIPConfigurations; configs != nil {
			for _, backendConfig := range *configs {
				if backendConfig.ID == nil {
					continue
				}
				backendIPConfigurations = append(backendIPConfigurations, *backendConfig.ID)
			}
		}
		if err := d.Set("backend_ip_configurations", backendIPConfigurations); err != nil {
			return fmt.Errorf("setting `backend_ip_configurations`: %v", err)
		}

		var loadBalancingRules []string
		if rules := props.LoadBalancingRules; rules != nil {
			for _, rule := range *rules {
				if rule.ID == nil {
					continue
				}
				loadBalancingRules = append(loadBalancingRules, *rule.ID)
			}
		}
		if err := d.Set("load_balancing_rules", loadBalancingRules); err != nil {
			return fmt.Errorf("setting `load_balancing_rules`: %v", err)
		}
	}

	return nil
}

func loadBalancerBackendAddressPoolDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	lbClient := meta.(*clients.Client).LoadBalancer.LoadBalancersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerBackendAddressPoolID(d.Id())
	if err != nil {
		return err
	}

	loadBalancerId := parse.NewLoadBalancerID(id.SubscriptionId, id.ResourceGroup, id.LoadBalancerName)
	loadBalancerID := loadBalancerId.ID()
	locks.ByID(loadBalancerID)
	defer locks.UnlockByID(loadBalancerID)

	locks.ByName(id.BackendAddressPoolName, backendAddressPoolResourceName)
	defer locks.UnlockByName(id.BackendAddressPoolName, backendAddressPoolResourceName)

	lb, err := lbClient.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(lb.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failed to retrieve Load Balancer %q (resource group %q) for Backend Address Pool %q: %+v", loadBalancerId.Name, loadBalancerId.ResourceGroup, id.BackendAddressPoolName, err)
	}

	_, index, exists := FindLoadBalancerBackEndAddressPoolByName(&lb, id.BackendAddressPoolName)
	if !exists {
		return nil
	}

	backEndPools := *lb.LoadBalancerPropertiesFormat.BackendAddressPools
	backEndPools = append(backEndPools[:index], backEndPools[index+1:]...)
	lb.LoadBalancerPropertiesFormat.BackendAddressPools = &backEndPools

	future, err := lbClient.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, lb)
	if err != nil {
		return fmt.Errorf("updating Load Balancer %q (resource group %q) to remove Backend Address Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.BackendAddressPoolName, err)
	}

	if err = future.WaitForCompletionRef(ctx, lbClient.Client); err != nil {
		return fmt.Errorf("waiting for update of Load Balancer %q (resource group %q) for Backend Address Pool %q: %+v", id.LoadBalancerName, id.ResourceGroup, id.BackendAddressPoolName, err)
	}

	return nil
}
