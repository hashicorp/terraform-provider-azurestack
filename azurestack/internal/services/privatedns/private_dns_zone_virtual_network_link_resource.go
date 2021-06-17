package privatedns

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/privatedns/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourcePrivateDnsZoneVirtualNetworkLink() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePrivateDnsZoneVirtualNetworkLinkCreateUpdate,
		Read:   resourcePrivateDnsZoneVirtualNetworkLinkRead,
		Update: resourcePrivateDnsZoneVirtualNetworkLinkCreateUpdate,
		Delete: resourcePrivateDnsZoneVirtualNetworkLinkDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.VirtualNetworkLinkID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		// TODO: these can become case-sensitive with a state migration
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				// TODO: make this case sensitive once the API's fixed https://github.com/Azure/azure-rest-api-specs/issues/10933
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"private_dns_zone_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"virtual_network_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"registration_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			// TODO: make this case sensitive once the API's fixed https://github.com/Azure/azure-rest-api-specs/issues/10933
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"tags": tags.Schema(),
		},
	}
}

func resourcePrivateDnsZoneVirtualNetworkLinkCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).PrivateDns.VirtualNetworkLinksClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	vNetID := d.Get("virtual_network_id").(string)
	registrationEnabled := d.Get("registration_enabled").(bool)

	resourceId := parse.NewVirtualNetworkLinkID(subscriptionId, d.Get("resource_group_name").(string), d.Get("private_dns_zone_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceId.ResourceGroup, resourceId.PrivateDnsZoneName, resourceId.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Virtual Network Link %q (Private DNS Zone %q / Resource Group %q): %s", resourceId.Name, resourceId.PrivateDnsZoneName, resourceId.ResourceGroup, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_private_dns_zone_virtual_network_link", resourceId.ID())
		}
	}

	location := "global"
	t := d.Get("tags").(map[string]interface{})

	parameters := privatedns.VirtualNetworkLink{
		Location: &location,
		Tags:     tags.Expand(t),
		VirtualNetworkLinkProperties: &privatedns.VirtualNetworkLinkProperties{
			VirtualNetwork: &privatedns.SubResource{
				ID: &vNetID,
			},
			RegistrationEnabled: &registrationEnabled,
		},
	}

	etag := ""
	ifNoneMatch := "" // set to empty to allow updates to records after creation

	future, err := client.CreateOrUpdate(ctx, resourceId.ResourceGroup, resourceId.PrivateDnsZoneName, resourceId.Name, parameters, etag, ifNoneMatch)
	if err != nil {
		return fmt.Errorf("creating/updating Virtual Network Link %q (Private DNS Zone %q / Resource Group %q): %+v", resourceId.Name, resourceId.PrivateDnsZoneName, resourceId.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for Virtual Network Link %q (Private DNS Zone %q / Resource Group %q) to become available: %+v", resourceId.Name, resourceId.PrivateDnsZoneName, resourceId.ResourceGroup, err)
	}

	d.SetId(resourceId.ID())
	return resourcePrivateDnsZoneVirtualNetworkLinkRead(d, meta)
}

func resourcePrivateDnsZoneVirtualNetworkLinkRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).PrivateDns.VirtualNetworkLinksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkLinkID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.PrivateDnsZoneName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Virtual Network Link %q (Private DNS Zone %q / Resource Group %q): %+v", id.Name, id.PrivateDnsZoneName, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("private_dns_zone_name", id.PrivateDnsZoneName)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := resp.VirtualNetworkLinkProperties; props != nil {
		d.Set("registration_enabled", props.RegistrationEnabled)

		if network := props.VirtualNetwork; network != nil {
			d.Set("virtual_network_id", network.ID)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourcePrivateDnsZoneVirtualNetworkLinkDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).PrivateDns.VirtualNetworkLinksClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.VirtualNetworkLinkID(d.Id())
	if err != nil {
		return err
	}

	etag := ""
	if future, err := client.Delete(ctx, id.ResourceGroup, id.PrivateDnsZoneName, id.Name, etag); err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("deleting Virtual Network Link %q (Private DNS Zone %q / Resource Group %q): %+v", id.Name, id.PrivateDnsZoneName, id.ResourceGroup, err)
	}

	// whilst the Delete above returns a Future, the Azure API's broken such that even though it's marked as "gone"
	// it's still kicking around - so we have to poll until this is actually gone
	log.Printf("[DEBUG] Waiting for Virtual Network Link %q (Private DNS Zone %q / Resource Group %q) to be deleted", id.Name, id.PrivateDnsZoneName, id.ResourceGroup)
	stateConf := &pluginsdk.StateChangeConf{
		Pending: []string{"Available"},
		Target:  []string{"NotFound"},
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking to see if Virtual Network Link %q (Private DNS Zone %q / Resource Group %q) is still available", id.Name, id.PrivateDnsZoneName, id.ResourceGroup)
			resp, err := client.Get(ctx, id.ResourceGroup, id.PrivateDnsZoneName, id.Name)
			if err != nil {
				if utils.ResponseWasNotFound(resp.Response) {
					log.Printf("[DEBUG] Virtual Network Link %q (Private DNS Zone %q / Resource Group %q) was not found", id.Name, id.PrivateDnsZoneName, id.ResourceGroup)
					return "NotFound", "NotFound", nil
				}

				return "", "error", err
			}

			log.Printf("[DEBUG] Virtual Network Link %q (Private DNS Zone %q / Resource Group %q) still exists", id.Name, id.PrivateDnsZoneName, id.ResourceGroup)
			return "Available", "Available", nil
		},
		Delay:                     30 * time.Second,
		PollInterval:              10 * time.Second,
		ContinuousTargetOccurence: 10,
		Timeout:                   d.Timeout(pluginsdk.TimeoutDelete),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("waiting for deletion of Virtual Network Link %q (Private DNS Zone %q / Resource Group %q): %+v", id.Name, id.PrivateDnsZoneName, id.ResourceGroup, err)
	}

	return nil
}
