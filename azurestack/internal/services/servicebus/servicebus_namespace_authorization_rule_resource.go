package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/servicebus/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/servicebus/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceServiceBusNamespaceAuthorizationRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusNamespaceAuthorizationRuleCreateUpdate,
		Read:   resourceServiceBusNamespaceAuthorizationRuleRead,
		Update: resourceServiceBusNamespaceAuthorizationRuleCreateUpdate,
		Delete: resourceServiceBusNamespaceAuthorizationRuleDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.NamespaceAuthorizationRuleID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		// function takes a schema map and adds the authorization rule properties to it
		Schema: authorizationRuleSchemaFrom(map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AuthorizationRuleName(),
			},

			"namespace_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NamespaceName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),
		}),

		CustomizeDiff: pluginsdk.CustomizeDiffShim(authorizationRuleCustomizeDiff),
	}
}

func resourceServiceBusNamespaceAuthorizationRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for ServiceBus Namespace Authorization Rule create/update.")
	resourceId := parse.NewNamespaceAuthorizationRuleID(subscriptionId, d.Get("resource_group_name").(string), d.Get("namespace_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.GetAuthorizationRule(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.AuthorizationRuleName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", resourceId, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_servicebus_namespace_authorization_rule", resourceId.ID())
		}
	}

	parameters := servicebus.SBAuthorizationRule{
		Name: utils.String(resourceId.AuthorizationRuleName),
		SBAuthorizationRuleProperties: &servicebus.SBAuthorizationRuleProperties{
			Rights: expandAuthorizationRuleRights(d),
		},
	}

	if _, err := client.CreateOrUpdateAuthorizationRule(ctx, resourceId.ResourceGroup, resourceId.NamespaceName, resourceId.AuthorizationRuleName, parameters); err != nil {
		return fmt.Errorf("creating/updating %s: %+v", resourceId, err)
	}

	d.SetId(resourceId.ID())

	return resourceServiceBusNamespaceAuthorizationRuleRead(d, meta)
}

func resourceServiceBusNamespaceAuthorizationRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NamespaceAuthorizationRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetAuthorizationRule(ctx, id.ResourceGroup, id.NamespaceName, id.AuthorizationRuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	keysResp, err := client.ListKeys(ctx, id.ResourceGroup, id.NamespaceName, id.AuthorizationRuleName)
	if err != nil {
		return fmt.Errorf("listing keys for %s: %+v", id, err)
	}

	d.Set("name", id.AuthorizationRuleName)
	d.Set("namespace_name", id.NamespaceName)
	d.Set("resource_group_name", id.ResourceGroup)

	if properties := resp.SBAuthorizationRuleProperties; properties != nil {
		listen, send, manage := flattenAuthorizationRuleRights(properties.Rights)
		d.Set("manage", manage)
		d.Set("listen", listen)
		d.Set("send", send)
	}

	d.Set("primary_key", keysResp.PrimaryKey)
	d.Set("primary_connection_string", keysResp.PrimaryConnectionString)
	d.Set("secondary_key", keysResp.SecondaryKey)
	d.Set("secondary_connection_string", keysResp.SecondaryConnectionString)

	return nil
}

func resourceServiceBusNamespaceAuthorizationRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.NamespacesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.NamespaceAuthorizationRuleID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.DeleteAuthorizationRule(ctx, id.ResourceGroup, id.NamespaceName, id.AuthorizationRuleName); err != nil {
		return fmt.Errorf("deleting %s: %+v", id, err)
	}

	return nil
}
