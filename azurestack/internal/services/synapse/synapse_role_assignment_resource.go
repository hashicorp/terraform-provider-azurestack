package synapse

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/2020-02-01-preview/accesscontrol"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/synapse/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/synapse/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceSynapseRoleAssignment() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceSynapseRoleAssignmentCreate,
		Read:   resourceSynapseRoleAssignmentRead,
		Delete: resourceSynapseRoleAssignmentDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.RoleAssignmentID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"synapse_workspace_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.WorkspaceID,
			},

			"principal_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"role_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Workspace Admin",
					"Apache Spark Admin",
					"Sql Admin",
				}, false),
			},
		},
	}
}

func resourceSynapseRoleAssignmentCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	synapseClient := meta.(*clients.Client).Synapse
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	environment := meta.(*clients.Client).Account.Environment

	workspaceId, err := parse.WorkspaceID(d.Get("synapse_workspace_id").(string))
	if err != nil {
		return err
	}
	principalID := d.Get("principal_id").(string)
	roleName := d.Get("role_name").(string)

	client, err := synapseClient.AccessControlClient(workspaceId.Name, environment.SynapseEndpointSuffix)
	if err != nil {
		return err
	}
	roleId, err := getRoleIdByName(ctx, client, roleName)
	if err != nil {
		return err
	}

	// check exist
	listResp, err := client.GetRoleAssignments(ctx, roleId, principalID, "")
	if err != nil {
		if !utils.ResponseWasNotFound(listResp.Response) {
			return fmt.Errorf("checking for presence of existing Synapse Role Assignment (workspace %q): %+v", workspaceId.Name, err)
		}
	}
	if listResp.Value != nil && len(*listResp.Value) != 0 {
		existing := (*listResp.Value)[0]
		if existing.ID != nil && *existing.ID != "" {
			resourceId := parse.NewRoleAssignmentId(*workspaceId, *existing.ID).ID()
			return tf.ImportAsExistsError("azurerm_synapse_role_assignment", resourceId)
		}
	}

	// create
	roleAssignment := accesscontrol.RoleAssignmentOptions{
		RoleID:      utils.String(roleId),
		PrincipalID: utils.String(principalID),
	}
	resp, err := client.CreateRoleAssignment(ctx, roleAssignment)
	if err != nil {
		return fmt.Errorf("creating Synapse RoleAssignment %q: %+v", roleName, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Synapse RoleAssignment %q", roleName)
	}

	resourceId := parse.NewRoleAssignmentId(*workspaceId, *resp.ID).ID()
	d.SetId(resourceId)
	return resourceSynapseRoleAssignmentRead(d, meta)
}

func resourceSynapseRoleAssignmentRead(d *pluginsdk.ResourceData, meta interface{}) error {
	synapseClient := meta.(*clients.Client).Synapse
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()
	environment := meta.(*clients.Client).Account.Environment

	id, err := parse.RoleAssignmentID(d.Id())
	if err != nil {
		return err
	}

	client, err := synapseClient.AccessControlClient(id.Workspace.Name, environment.SynapseEndpointSuffix)
	if err != nil {
		return err
	}
	resp, err := client.GetRoleAssignmentByID(ctx, id.DataPlaneAssignmentId)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] synapse role assignment %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Synapse RoleAssignment (Resource Group %q): %+v", id.Workspace.Name, err)
	}

	if resp.RoleID != nil {
		role, err := client.GetRoleDefinitionByID(ctx, *resp.RoleID)
		if err != nil {
			return fmt.Errorf("retrieving role definition by ID %q: %+v", *resp.RoleID, err)
		}
		d.Set("role_name", role.Name)
	}

	d.Set("synapse_workspace_id", id.Workspace.ID())
	d.Set("principal_id", resp.PrincipalID)
	return nil
}

func resourceSynapseRoleAssignmentDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	synapseClient := meta.(*clients.Client).Synapse
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()
	environment := meta.(*clients.Client).Account.Environment

	id, err := parse.RoleAssignmentID(d.Id())
	if err != nil {
		return err
	}

	client, err := synapseClient.AccessControlClient(id.Workspace.Name, environment.SynapseEndpointSuffix)
	if err != nil {
		return err
	}
	if _, err := client.DeleteRoleAssignmentByID(ctx, id.DataPlaneAssignmentId); err != nil {
		return fmt.Errorf("deleting Synapse RoleAssignment %q (workspace %q): %+v", id, id.Workspace.Name, err)
	}

	return nil
}

func getRoleIdByName(ctx context.Context, client *accesscontrol.BaseClient, roleName string) (string, error) {
	resp, err := client.GetRoleDefinitions(ctx)
	if err != nil {
		return "", fmt.Errorf("listing synapse role definitions %+v", err)
	}

	var availableRoleName []string
	for resp.NotDone() {
		for _, role := range resp.Values() {
			if role.Name != nil {
				if *role.Name == roleName && role.ID != nil {
					return *role.ID, nil
				}
				availableRoleName = append(availableRoleName, *role.Name)
			}
		}
		if err := resp.NextWithContext(ctx); err != nil {
			return "", fmt.Errorf("retrieving next page of synapse role definitions: %+v", err)
		}
	}
	return "", fmt.Errorf("role name %q invalid. Available role names are %q", roleName, strings.Join(availableRoleName, ","))
}
