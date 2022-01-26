package loadbalancer

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/loadbalancer/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func loadBalancerBackendAddressPoolDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: loadBalancerBackendAddressPoolDataSourceRead,
		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"loadbalancer_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"backend_ip_configurations": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"id": {
							Type:     pluginsdk.TypeString,
							Computed: true,
						},
					},
				},
			},

			"load_balancing_rules": {
				Type:     pluginsdk.TypeList,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func loadBalancerBackendAddressPoolDataSourceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LoadBalancer.LoadBalancerBackendAddressPoolsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewLoadBalancerBackendAddressPoolID(loadBalancerId.SubscriptionId, loadBalancerId.ResourceGroup, loadBalancerId.Name, name)

	resp, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, id.BackendAddressPoolName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Load Balancer Backend Address Pool %q was not found", id)
		}
		return fmt.Errorf("failed to retrieve Load Balancer Backend Address Pool %q: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	if props := resp.BackendAddressPoolPropertiesFormat; props != nil {
		var backendIPConfigurations []interface{}
		if beipConfigs := props.BackendIPConfigurations; beipConfigs != nil {
			for _, config := range *beipConfigs {
				ipConfig := make(map[string]interface{})
				if id := config.ID; id != nil {
					ipConfig["id"] = *id
					backendIPConfigurations = append(backendIPConfigurations, ipConfig)
				}
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
