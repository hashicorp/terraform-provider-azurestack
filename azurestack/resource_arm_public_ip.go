package azurestack

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/network/mgmt/network"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/utils"
)

func resourceArmPublicIp() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmPublicIpCreate,
		Read:   resourceArmPublicIpRead,
		Update: resourceArmPublicIpCreate,
		Delete: resourceArmPublicIpDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				id, err := parseAzureResourceID(d.Id())
				if err != nil {
					return nil, err
				}
				name := id.Path["publicIPAddresses"]
				if name == "" {
					return nil, fmt.Errorf("Error parsing supplied resource id. Please check it and rerun:\n %s", d.Id())
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			// Not supported for 2017-03-09 profile
			// "zones": singleZonesSchema(),

			"public_ip_address_allocation": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validatePublicIpAllocation,
				StateFunc:        ignoreCaseStateFunc,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"idle_timeout_in_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(int)
					if value < 4 || value > 30 {
						errors = append(errors, fmt.Errorf(
							"The idle timeout must be between 4 and 30 minutes"))
					}
					return
				},
			},

			"domain_name_label": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validatePublicIpDomainNameLabel,
			},

			"reverse_fqdn": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Not supported for 2017-03-09 profile
			// "sku": {
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// 	Default:  "Basic",
			// 	ForceNew: true,
			// 	ValidateFunc: validation.StringInSlice([]string{
			// 		"Basic",
			// 		"Standard",
			// 	}, true),
			// 	DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			// },

			"tags": tagsSchema(),
		},
	}
}

func resourceArmPublicIpCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).publicIPClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureStack Public IP creation.")

	name := d.Get("name").(string)
	location := azureStackNormalizeLocation(d.Get("location").(string))
	resGroup := d.Get("resource_group_name").(string)
	// Not supported for 2017-03-09 profile
	// sku := network.PublicIPAddressSku{
	// 	Name: network.PublicIPAddressSkuName(d.Get("sku").(string)),
	// }
	tags := d.Get("tags").(map[string]interface{})

	// Not supported for 2017-03-09 profile
	// zones := expandZones(d.Get("zones").([]interface{}))

	ipAllocationMethod := network.IPAllocationMethod(d.Get("public_ip_address_allocation").(string))

	// Not supported for 2017-03-09 profile
	// if strings.ToLower(string(sku.Name)) == "standard" {
	// 	if strings.ToLower(string(ipAllocationMethod)) != "static" {
	// 		return fmt.Errorf("Static IP allocation must be used when creating Standard SKU public IP addresses.")
	// 	}
	// }

	properties := network.PublicIPAddressPropertiesFormat{
		PublicIPAllocationMethod: ipAllocationMethod,
	}

	dnl, hasDnl := d.GetOk("domain_name_label")
	rfqdn, hasRfqdn := d.GetOk("reverse_fqdn")

	if hasDnl || hasRfqdn {
		dnsSettings := network.PublicIPAddressDNSSettings{}

		if hasRfqdn {
			reverseFqdn := rfqdn.(string)
			dnsSettings.ReverseFqdn = &reverseFqdn
		}

		if hasDnl {
			domainNameLabel := dnl.(string)
			dnsSettings.DomainNameLabel = &domainNameLabel
		}

		properties.DNSSettings = &dnsSettings
	}

	if v, ok := d.GetOk("idle_timeout_in_minutes"); ok {
		idleTimeout := int32(v.(int))
		properties.IdleTimeoutInMinutes = &idleTimeout
	}

	publicIp := network.PublicIPAddress{
		Name:                            &name,
		Location:                        &location,
		PublicIPAddressPropertiesFormat: &properties,
		Tags: *expandTags(tags),
		// Not supported for 2017-03-09 profile
		// Sku:      &sku,
		// Zones: zones,
	}

	future, err := client.CreateOrUpdate(ctx, resGroup, name, publicIp)
	if err != nil {
		return fmt.Errorf("Error Creating/Updating Public IP %q (Resource Group %q): %+v", name, resGroup, err)
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return fmt.Errorf("Error waiting for completion of Public IP %q (Resource Group %q): %+v", name, resGroup, err)
	}

	read, err := client.Get(ctx, resGroup, name, "")
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Public IP %q (resource group %q) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmPublicIpRead(d, meta)
}

func resourceArmPublicIpRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).publicIPClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["publicIPAddresses"]

	resp, err := client.Get(ctx, resGroup, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Public IP %q (Resource Group %q): %+v", name, resGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azureStackNormalizeLocation(*location))
	}

	d.Set("public_ip_address_allocation", strings.ToLower(string(resp.PublicIPAddressPropertiesFormat.PublicIPAllocationMethod)))

	// if sku := resp.Sku; sku != nil {
	// 	d.Set("sku", string(sku.Name))
	// }
	// d.Set("zones", resp.Zones)

	if props := resp.PublicIPAddressPropertiesFormat; props != nil {
		d.Set("public_ip_address_allocation", strings.ToLower(string(props.PublicIPAllocationMethod)))

		if settings := props.DNSSettings; settings != nil {
			if fqdn := settings.Fqdn; fqdn != nil {
				d.Set("fqdn", fqdn)
			}
		}

		if ip := props.IPAddress; ip != nil {
			d.Set("ip_address", ip)
		}
	}

	flattenAndSetTags(d, &resp.Tags)

	return nil
}

func resourceArmPublicIpDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).publicIPClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["publicIPAddresses"]

	future, err := client.Delete(ctx, resGroup, name)
	if err != nil {
		return fmt.Errorf("Error deleting Public IP %q (Resource Group %q): %+v", name, resGroup, err)
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return fmt.Errorf("Error waiting for deletion of Public IP %q (Resource Group %q): %+v", name, resGroup, err)
	}

	return nil
}

func validatePublicIpAllocation(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	allocations := map[string]bool{
		"static":  true,
		"dynamic": true,
	}

	if !allocations[value] {
		errors = append(errors, fmt.Errorf("Public IP Allocation must be an accepted value: Static, Dynamic"))
	}
	return
}

func validatePublicIpDomainNameLabel(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}

	if len(value) > 61 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 61 characters: %q", k, value))
	}

	if len(value) == 0 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be an empty string: %q", k, value))
	}
	if regexp.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen: %q", k, value))
	}

	return
}
