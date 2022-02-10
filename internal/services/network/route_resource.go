package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/locks"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func resourceRoute() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: routeCreateUpdate,
		Read:   routeRead,
		Update: routeCreateUpdate,
		Delete: routeDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.RouteID(id)
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
				ValidateFunc: validate.RouteName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"route_table_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.RouteTableName,
			},

			"address_prefix": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"next_hop_type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.RouteNextHopTypeVirtualNetworkGateway),
					string(network.RouteNextHopTypeVnetLocal),
					string(network.RouteNextHopTypeInternet),
					string(network.RouteNextHopTypeVirtualAppliance),
					string(network.RouteNextHopTypeNone),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"next_hop_in_ip_address": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func routeCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RoutesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	addressPrefix := d.Get("address_prefix").(string)
	nextHopType := d.Get("next_hop_type").(string)

	id := parse.NewRouteID(subscriptionId, d.Get("resource_group_name").(string), d.Get("route_table_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.RouteTableName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_route", id.ID())
		}
	}

	locks.ByName(id.RouteTableName, routeTableResourceName)
	defer locks.UnlockByName(id.RouteTableName, routeTableResourceName)

	route := network.Route{
		Name: pointer.FromString(id.Name),
		RoutePropertiesFormat: &network.RoutePropertiesFormat{
			AddressPrefix: &addressPrefix,
			NextHopType:   network.RouteNextHopType(nextHopType),
		},
	}

	if v, ok := d.GetOk("next_hop_in_ip_address"); ok {
		route.RoutePropertiesFormat.NextHopIPAddress = pointer.FromString(v.(string))
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.RouteTableName, id.Name, route)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for create/update of %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this
	return routeRead(d, meta)
}

func routeRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RoutesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RouteID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.RouteTableName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("route_table_name", id.RouteTableName)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := resp.RoutePropertiesFormat; props != nil {
		d.Set("address_prefix", props.AddressPrefix)
		d.Set("next_hop_type", string(props.NextHopType))
		d.Set("next_hop_in_ip_address", props.NextHopIPAddress)
	}

	return nil
}

func routeDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RoutesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RouteID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.RouteTableName, routeTableResourceName)
	defer locks.UnlockByName(id.RouteTableName, routeTableResourceName)

	future, err := client.Delete(ctx, id.ResourceGroup, id.RouteTableName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}
