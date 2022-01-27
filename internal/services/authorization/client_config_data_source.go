package authorization

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
)

func clientConfigDataSource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: clientConfigRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"client_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tenant_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"subscription_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"object_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"service_principal_application_id": {
				Type:     schema.TypeString,
				Computed: true,
				// todo should we deprecate this?
			},

			"service_principal_object_id": {
				Type:     schema.TypeString,
				Computed: true,
				// todo should we deprecate this?
			},
		},
	}
}

func clientConfigRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client)
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	if client.Account.AuthenticatedAsAServicePrincipal {
		spClient := client.Authorization.ServicePrincipalsClient
		// Application & Service Principal is 1:1 per tenant. Since we know the appId (client_id)
		// here, we can query for the Service Principal whose appId matches.
		filter := fmt.Sprintf("appId eq '%s'", client.Account.ClientId)
		listResult, listErr := spClient.List(ctx, filter)

		if listErr != nil {
			return fmt.Errorf("listing Service Principals: %#v", listErr)
		}

		if listResult.Values() == nil || len(listResult.Values()) != 1 {
			return fmt.Errorf("Unexpected Service Principal query result: %#v", listResult.Values())
		}

		servicePrincipal := &(listResult.Values())[0]
		if principal := servicePrincipal; principal != nil {
			d.Set("service_principal_application_id", principal.AppID)
			d.Set("service_principal_object_id", principal.ObjectID)
		}
	}

	d.SetId(time.Now().UTC().String())
	d.Set("client_id", client.Account.ClientId)
	d.Set("object_id", client.Account.ObjectId)
	d.Set("subscription_id", client.Account.SubscriptionId)
	d.Set("tenant_id", client.Account.TenantId)

	return nil
}
