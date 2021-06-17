package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-11-01/network"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/identity"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/location"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

type applicationGatewayDataSourceIdentity = identity.UserAssigned

func dataSourceApplicationGateway() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceApplicationGatewayRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"location": azure.SchemaLocationForDataSource(),

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"identity": applicationGatewayDataSourceIdentity{}.SchemaDataSource(),

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceApplicationGatewayRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.ApplicationGatewaysClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewApplicationGatewayID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("%s was not found", id)
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("location", location.NormalizeNilable(resp.Location))

	identity := flattenApplicationGatewayDataSourceIdentity(resp.Identity)
	flattenedIdentity := applicationGatewayDataSourceIdentity{}.Flatten(identity)
	if err = d.Set("identity", flattenedIdentity); err != nil {
		return err
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func flattenApplicationGatewayDataSourceIdentity(input *network.ManagedServiceIdentity) *identity.ExpandedConfig {
	var config *identity.ExpandedConfig
	if input != nil {
		config = &identity.ExpandedConfig{
			Type:        string(input.Type),
			PrincipalId: input.PrincipalID,
			TenantId:    input.TenantID,
		}
	}
	return config
}
