// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/network/mgmt/network"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/network/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func networkSecurityRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: networkSecurityRuleCreateUpdate,
		Read:   networkSecurityRuleRead,
		Update: networkSecurityRuleCreateUpdate,
		Delete: networkSecurityRuleDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.SecurityRuleID(id)
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
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"network_security_group_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 140),
			},

			"protocol": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityRuleProtocolAsterisk),
					string(network.SecurityRuleProtocolTCP),
					string(network.SecurityRuleProtocolUDP),
					// string(network.SecurityRuleProtocolIcmp),
					// string(network.SecurityRuleProtocolAh),
					// string(network.SecurityRuleProtocolEsp),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"source_port_range": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source_port_ranges"},
			},

			"source_port_ranges": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"source_port_range"},
			},

			"destination_port_range": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"destination_port_ranges"},
			},

			"destination_port_ranges": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"destination_port_range"},
			},

			"source_address_prefix": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source_address_prefixes"},
			},

			"source_address_prefixes": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"source_address_prefix"},
			},

			"destination_address_prefix": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ConflictsWith: []string{"destination_address_prefixes"},
			},

			"destination_address_prefixes": {
				Type:          pluginsdk.TypeSet,
				Optional:      true,
				Elem:          &pluginsdk.Schema{Type: pluginsdk.TypeString},
				Set:           pluginsdk.HashString,
				ConflictsWith: []string{"destination_address_prefix"},
			},

			"access": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityRuleAccessAllow),
					string(network.SecurityRuleAccessDeny),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"priority": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(100, 4096),
			},

			"direction": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.SecurityRuleDirectionInbound),
					string(network.SecurityRuleDirectionOutbound),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},
		},
	}
}

func networkSecurityRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityRuleClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewSecurityRuleID(subscriptionId, d.Get("resource_group_name").(string), d.Get("network_security_group_name").(string), d.Get("name").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurestack_network_security_rule", *existing.ID)
		}
	}

	sourcePortRange := d.Get("source_port_range").(string)
	destinationPortRange := d.Get("destination_port_range").(string)
	sourceAddressPrefix := d.Get("source_address_prefix").(string)
	destinationAddressPrefix := d.Get("destination_address_prefix").(string)
	priority := int32(d.Get("priority").(int))
	access := d.Get("access").(string)
	direction := d.Get("direction").(string)
	protocol := d.Get("protocol").(string)

	// TODO should we put this into stack?
	/* if !meta.(*clients.Client).Features.Network.RelaxedLocking {
		locks.ByName(id.NetworkSecurityGroupName, networkSecurityGroupResourceName)
		defer locks.UnlockByName(id.NetworkSecurityGroupName, networkSecurityGroupResourceName)
	}*/

	rule := network.SecurityRule{
		Name: &id.Name,
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			SourcePortRange:          &sourcePortRange,
			DestinationPortRange:     &destinationPortRange,
			SourceAddressPrefix:      &sourceAddressPrefix,
			DestinationAddressPrefix: &destinationAddressPrefix,
			Priority:                 &priority,
			Access:                   network.SecurityRuleAccess(access),
			Direction:                network.SecurityRuleDirection(direction),
			Protocol:                 network.SecurityRuleProtocol(protocol),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		rule.SecurityRulePropertiesFormat.Description = &description
	}

	if r, ok := d.GetOk("source_port_ranges"); ok {
		var sourcePortRanges []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			sourcePortRanges = append(sourcePortRanges, s)
		}
		rule.SecurityRulePropertiesFormat.SourcePortRanges = &sourcePortRanges
	}

	if r, ok := d.GetOk("destination_port_ranges"); ok {
		var destinationPortRanges []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			destinationPortRanges = append(destinationPortRanges, s)
		}
		rule.SecurityRulePropertiesFormat.DestinationPortRanges = &destinationPortRanges
	}

	if r, ok := d.GetOk("source_address_prefixes"); ok {
		var sourceAddressPrefixes []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			sourceAddressPrefixes = append(sourceAddressPrefixes, s)
		}
		rule.SecurityRulePropertiesFormat.SourceAddressPrefixes = &sourceAddressPrefixes
	}

	if r, ok := d.GetOk("destination_address_prefixes"); ok {
		var destinationAddressPrefixes []string
		r := r.(*pluginsdk.Set).List()
		for _, v := range r {
			s := v.(string)
			destinationAddressPrefixes = append(destinationAddressPrefixes, s)
		}
		rule.SecurityRulePropertiesFormat.DestinationAddressPrefixes = &destinationAddressPrefixes
	}

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name, rule)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for completion of %s: %+v", id, err)
	}

	d.SetId(id.ID()) // TODO before release confirm no state migration is required for this

	return networkSecurityRuleRead(d, meta)
}

func networkSecurityRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityRuleClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("making Read request on %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("network_security_group_name", id.NetworkSecurityGroupName)

	if props := resp.SecurityRulePropertiesFormat; props != nil {
		d.Set("description", props.Description)
		d.Set("protocol", string(props.Protocol))
		d.Set("destination_address_prefix", props.DestinationAddressPrefix)
		d.Set("destination_address_prefixes", props.DestinationAddressPrefixes)
		d.Set("destination_port_range", props.DestinationPortRange)
		d.Set("destination_port_ranges", props.DestinationPortRanges)
		d.Set("source_address_prefix", props.SourceAddressPrefix)
		d.Set("source_address_prefixes", props.SourceAddressPrefixes)
		d.Set("source_port_range", props.SourcePortRange)
		d.Set("source_port_ranges", props.SourcePortRanges)
		d.Set("access", string(props.Access))
		d.Set("priority", int(*props.Priority))
		d.Set("direction", string(props.Direction))
	}

	return nil
}

func networkSecurityRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.SecurityRuleClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SecurityRuleID(d.Id())
	if err != nil {
		return err
	}

	// TODO should we put this into stack?
	/* if !meta.(*clients.Client).Features.Network.RelaxedLocking {
		locks.ByName(id.NetworkSecurityGroupName, networkSecurityGroupResourceName)
		defer locks.UnlockByName(id.NetworkSecurityGroupName, networkSecurityGroupResourceName)
	}*/

	future, err := client.Delete(ctx, id.ResourceGroup, id.NetworkSecurityGroupName, id.Name)
	if err != nil {
		return fmt.Errorf("Deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return nil
}
