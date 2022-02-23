package network

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func publicIp() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: publicIpCreateUpdate,
		Read:   publicIpRead,
		Update: publicIpCreateUpdate,
		Delete: publicIpDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.PublicIpAddressID(id)
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
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"public_ip_address_allocation": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ExactlyOneOf:     []string{"public_ip_address_allocation", "allocation_method"},
				Deprecated:       "the `public_ip_address_allocation` property has been renamed to `allocation_method`",
				ValidateFunc: validation.StringInSlice([]string{
					string(network.Dynamic),
					string(network.Static),
				}, true),
			},

			"allocation_method": {
				Type:         pluginsdk.TypeString,
				Optional:     true, // TODO required when deprecation completed
				Computed:     true,
				ExactlyOneOf: []string{"public_ip_address_allocation", "allocation_method"},
				ValidateFunc: validation.StringInSlice([]string{
					string(network.Static),
					string(network.Dynamic),
				}, false),
			},

			"ip_version": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Default:          string(network.IPv4),
				ForceNew:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.IPv4),
					string(network.IPv6),
				}, true),
			},

			"sku": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          string(network.PublicIPAddressSkuNameBasic),
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.PublicIPAddressSkuNameBasic),
					string(network.PublicIPAddressSkuNameStandard),
				}, true),
			},

			"idle_timeout_in_minutes": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      4,
				ValidateFunc: validation.IntBetween(4, 30),
			},

			"domain_name_label": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.PublicIpDomainNameLabel,
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"reverse_fqdn": {
				Type:     pluginsdk.TypeString,
				Optional: true,
			},

			"ip_address": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func publicIpCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for azurestack Public IP creation.")

	id := parse.NewPublicIpAddressID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurestack_public_ip", id.ID())
		}
	}

	location := location.Normalize(d.Get("location").(string))
	sku := d.Get("sku").(string)
	t := d.Get("tags").(map[string]interface{})

	idleTimeout := d.Get("idle_timeout_in_minutes").(int)
	ipVersion := network.IPVersion(d.Get("ip_version").(string))
	ipAllocationMethod := d.Get("allocation_method").(string)
	if ipAllocationMethod == "" {
		ipAllocationMethod = d.Get("public_ip_address_allocation").(string)
	}

	if strings.EqualFold(sku, "standard") {
		if !strings.EqualFold(ipAllocationMethod, "static") {
			return fmt.Errorf("Static IP allocation must be used when creating Standard SKU public IP addresses.")
		}
	}

	publicIp := network.PublicIPAddress{
		Name:     pointer.FromString(id.Name),
		Location: &location,
		Sku: &network.PublicIPAddressSku{
			Name: network.PublicIPAddressSkuName(sku),
		},
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.IPAllocationMethod(ipAllocationMethod),
			PublicIPAddressVersion:   ipVersion,
			IdleTimeoutInMinutes:     utils.Int32(int32(idleTimeout)),
		},
		Tags: tags.Expand(t),
	}

	dnl, dnlOk := d.GetOk("domain_name_label")
	rfqdn, rfqdnOk := d.GetOk("reverse_fqdn")

	if dnlOk || rfqdnOk {
		dnsSettings := network.PublicIPAddressDNSSettings{}

		if rfqdnOk {
			dnsSettings.ReverseFqdn = pointer.FromString(rfqdn.(string))
		}

		if dnlOk {
			dnsSettings.DomainNameLabel = pointer.FromString(dnl.(string))
		}

		publicIp.PublicIPAddressPropertiesFormat.DNSSettings = &dnsSettings
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, publicIp)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation/update of %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this
	return publicIpRead(d, meta)
}

func publicIpRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PublicIpAddressID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	d.Set("location", location.NormalizeNilable(resp.Location))

	if sku := resp.Sku; sku != nil {
		d.Set("sku", string(sku.Name))
	}

	if props := resp.PublicIPAddressPropertiesFormat; props != nil {
		d.Set("allocation_method", string(props.PublicIPAllocationMethod))
		d.Set("public_ip_address_allocation", string(props.PublicIPAllocationMethod))
		d.Set("ip_version", string(props.PublicIPAddressVersion))

		if settings := props.DNSSettings; settings != nil {
			d.Set("fqdn", settings.Fqdn)
			d.Set("reverse_fqdn", settings.ReverseFqdn)
			d.Set("domain_name_label", settings.DomainNameLabel)
		}

		d.Set("ip_address", props.IPAddress)
		d.Set("idle_timeout_in_minutes", props.IdleTimeoutInMinutes)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func publicIpDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.PublicIPsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.PublicIpAddressID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}
