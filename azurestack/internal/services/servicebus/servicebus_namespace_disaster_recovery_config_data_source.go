package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/servicebus/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourceServiceBusNamespaceDisasterRecoveryConfig() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceServiceBusNamespaceDisasterRecoveryConfigRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"namespace_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"partner_namespace_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"alias_primary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"alias_secondary_connection_string": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_primary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"default_secondary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceServiceBusNamespaceDisasterRecoveryConfigRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.DisasterRecoveryConfigsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewNamespaceDisasterRecoveryConfigID(subscriptionId, d.Get("resource_group_name").(string), d.Get("namespace_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id.ResourceGroup, id.NamespaceName, id.DisasterRecoveryConfigName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.DisasterRecoveryConfigName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("namespace_name", id.NamespaceName)
	d.Set("partner_namespace_id", resp.ArmDisasterRecoveryProperties.PartnerNamespace)
	d.SetId(*resp.ID)

	keys, err := client.ListKeys(ctx, id.ResourceGroup, id.NamespaceName, id.DisasterRecoveryConfigName, serviceBusNamespaceDefaultAuthorizationRule)

	if err != nil {
		log.Printf("[WARN] listing default keys for %s: %+v", id, err)
	} else {
		d.Set("alias_primary_connection_string", keys.AliasPrimaryConnectionString)
		d.Set("alias_secondary_connection_string", keys.AliasSecondaryConnectionString)
		d.Set("default_primary_key", keys.PrimaryKey)
		d.Set("default_secondary_key", keys.SecondaryKey)
	}
	return nil
}
