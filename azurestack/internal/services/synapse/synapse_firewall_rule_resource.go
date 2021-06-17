package synapse

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/mgmt/2019-06-01-preview/synapse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/synapse/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/synapse/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceSynapseFirewallRule() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSynapseFirewallRuleCreateUpdate,
		Read:   resourceSynapseFirewallRuleRead,
		Update: resourceSynapseFirewallRuleCreateUpdate,
		Delete: resourceSynapseFirewallRuleDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.FirewallRuleID(id)
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
				ValidateFunc: validate.FirewallRuleName,
			},

			"synapse_workspace_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.WorkspaceID,
			},

			"start_ip_address": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},

			"end_ip_address": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
		},
	}
}

func resourceSynapseFirewallRuleCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Synapse.FirewallRulesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	workspaceId, err := parse.WorkspaceID(d.Get("synapse_workspace_id").(string))
	if err != nil {
		return err
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, workspaceId.ResourceGroup, workspaceId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Synapse Firewall Rule %q (Resource Group %q / Workspace %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_synapse_firewall_rule", *existing.ID)
		}
	}

	parameters := synapse.IPFirewallRuleInfo{
		IPFirewallRuleProperties: &synapse.IPFirewallRuleProperties{
			StartIPAddress: utils.String(d.Get("start_ip_address").(string)),
			EndIPAddress:   utils.String(d.Get("end_ip_address").(string)),
		},
	}

	future, err := client.CreateOrUpdate(ctx, workspaceId.ResourceGroup, workspaceId.Name, name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating Synapse Firewall Rule %q (Resource Group %q / Workspace %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on creation/updation for Synapse Firewall Rule %q (Resource Group %q / Workspace %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
	}

	resp, err := client.Get(ctx, workspaceId.ResourceGroup, workspaceId.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Synapse Firewall Rule %q (Resource Group %q / Workspace %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
	}

	d.SetId(*resp.ID)

	return resourceSynapseFirewallRuleRead(d, meta)
}

func resourceSynapseFirewallRuleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Synapse.FirewallRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading Synapse Firewall Rule %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading Synapse Firewall Rule %q (Workspace %q / Resource Group %q): %+v", id.Name, id.WorkspaceName, id.ResourceGroup, err)
	}

	workspaceId := parse.NewWorkspaceID(id.SubscriptionId, id.ResourceGroup, id.WorkspaceName).ID()
	d.Set("name", id.Name)
	d.Set("synapse_workspace_id", workspaceId)

	if props := resp.IPFirewallRuleProperties; props != nil {
		d.Set("start_ip_address", props.StartIPAddress)
		d.Set("end_ip_address", props.EndIPAddress)
	}

	return nil
}

func resourceSynapseFirewallRuleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Synapse.FirewallRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallRuleID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.WorkspaceName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Synapse Firewall Rule %q (Workspace %q / Resource Group %q): %+v", id.Name, id.WorkspaceName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of Synapse Firewall Rule %q (Workspace %q / Resource Group %q): %+v", id.Name, id.WorkspaceName, id.ResourceGroup, err)
	}

	return nil
}
