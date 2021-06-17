package databricks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/databricks/mgmt/2018-04-01/databricks"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/databricks/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/databricks/validate"
	resourcesParse "github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/resource/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceDatabricksWorkspace() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceDatabricksWorkspaceCreateUpdate,
		Read:   resourceDatabricksWorkspaceRead,
		Update: resourceDatabricksWorkspaceCreateUpdate,
		Delete: resourceDatabricksWorkspaceDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.WorkspaceID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.WorkspaceName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"standard",
					"premium",
					"trial",
				}, false),
			},

			"tags": tags.Schema(),

			"managed_resource_group_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"custom_parameters": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"no_public_ip": {
							Type:     pluginsdk.TypeBool,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{"custom_parameters.0.no_public_ip", "custom_parameters.0.public_subnet_name",
								"custom_parameters.0.private_subnet_name", "custom_parameters.0.virtual_network_id",
							},
						},

						"public_subnet_name": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{"custom_parameters.0.no_public_ip", "custom_parameters.0.public_subnet_name",
								"custom_parameters.0.private_subnet_name", "custom_parameters.0.virtual_network_id",
							},
						},

						"private_subnet_name": {
							Type:     pluginsdk.TypeString,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{"custom_parameters.0.no_public_ip", "custom_parameters.0.public_subnet_name",
								"custom_parameters.0.private_subnet_name", "custom_parameters.0.virtual_network_id",
							},
						},

						"virtual_network_id": {
							Type:         pluginsdk.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: azure.ValidateResourceIDOrEmpty,
							AtLeastOneOf: []string{"custom_parameters.0.no_public_ip", "custom_parameters.0.public_subnet_name",
								"custom_parameters.0.private_subnet_name", "custom_parameters.0.virtual_network_id",
							},
						},
					},
				},
			},

			"managed_resource_group_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"workspace_url": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"workspace_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: pluginsdk.CustomizeDiffShim(func(ctx context.Context, d *pluginsdk.ResourceDiff, v interface{}) error {
			if d.HasChange("sku") {
				sku, changedSKU := d.GetChange("sku")

				if changedSKU == "trial" {
					log.Printf("[DEBUG] recreate databricks workspace, could't be migrated to %s", changedSKU)
					d.ForceNew("sku")
				} else {
					log.Printf("[DEBUG] databricks workspace can be upgraded from %s to %s", sku, changedSKU)
				}
			}

			return nil
		}),
	}
}

func resourceDatabricksWorkspaceCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataBricks.WorkspacesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewWorkspaceID(subscriptionId, d.Get("resource_group_name").(string), d.Get("name").(string))
	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_databricks_workspace", id.ID())
		}
	}

	skuName := d.Get("sku").(string)
	managedResourceGroupName := d.Get("managed_resource_group_name").(string)

	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})
	expandedTags := tags.Expand(t)

	if managedResourceGroupName == "" {
		// no managed resource group name was provided, we use the default pattern
		log.Printf("[DEBUG][azurerm_databricks_workspace] no managed resource group id was provided, we use the default pattern.")
		managedResourceGroupName = fmt.Sprintf("databricks-rg-%s", id.ResourceGroup)
	}
	managedResourceGroupID := resourcesParse.NewResourceGroupID(subscriptionId, managedResourceGroupName).ID()

	customParamsRaw := d.Get("custom_parameters").([]interface{})
	customParams := expandWorkspaceCustomParameters(customParamsRaw)

	// Including the Tags in the workspace parameters will update the tags on
	// the workspace only
	workspace := databricks.Workspace{
		Sku: &databricks.Sku{
			Name: utils.String(skuName),
		},
		Location: utils.String(location),
		WorkspaceProperties: &databricks.WorkspaceProperties{
			ManagedResourceGroupID: &managedResourceGroupID,
			Parameters:             customParams,
		},
		Tags: expandedTags,
	}

	future, err := client.CreateOrUpdate(ctx, workspace, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("creating/updating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for create/update of %s: %+v", id, err)
	}

	// Only call Update(e.g. PATCH) if it is not a new resource and the Tags have changed
	// this will cause the updated tags to be propagated to all of the connected
	// workspace resources.
	// TODO: can be removed once https://github.com/Azure/azure-sdk-for-go/issues/14571 is fixed
	if !d.IsNewResource() && d.HasChange("tags") {
		workspaceUpdate := databricks.WorkspaceUpdate{
			Tags: expandedTags,
		}

		future, err := client.Update(ctx, workspaceUpdate, id.ResourceGroup, id.Name)
		if err != nil {
			return fmt.Errorf("updating %s Tags: %+v", id, err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for %s Tags to be updated: %+v", id, err)
		}
	}

	d.SetId(id.ID())
	return resourceDatabricksWorkspaceRead(d, meta)
}

func resourceDatabricksWorkspaceRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataBricks.WorkspacesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.WorkspaceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %s was not found - removing from state", *id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku", sku.Name)
	}

	if props := resp.WorkspaceProperties; props != nil {
		managedResourceGroupID, err := resourcesParse.ResourceGroupID(*props.ManagedResourceGroupID)
		if err != nil {
			return err
		}
		d.Set("managed_resource_group_id", props.ManagedResourceGroupID)
		d.Set("managed_resource_group_name", managedResourceGroupID.ResourceGroup)

		if err := d.Set("custom_parameters", flattenWorkspaceCustomParameters(props.Parameters)); err != nil {
			return fmt.Errorf("setting `custom_parameters`: %+v", err)
		}

		d.Set("workspace_url", props.WorkspaceURL)
		d.Set("workspace_id", props.WorkspaceID)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceDatabricksWorkspaceDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataBricks.WorkspacesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.WorkspaceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
		}
	}

	return nil
}

func flattenWorkspaceCustomParameters(p *databricks.WorkspaceCustomParameters) []interface{} {
	if p == nil {
		return nil
	}

	parameters := make(map[string]interface{})

	if v := p.EnableNoPublicIP; v != nil {
		if v.Value != nil {
			parameters["no_public_ip"] = *v.Value
		}
	}

	if v := p.CustomPrivateSubnetName; v != nil {
		if v.Value != nil {
			parameters["private_subnet_name"] = *v.Value
		}
	}

	if v := p.CustomPublicSubnetName; v != nil {
		if v.Value != nil {
			parameters["public_subnet_name"] = *v.Value
		}
	}

	if v := p.CustomVirtualNetworkID; v != nil {
		if v.Value != nil {
			parameters["virtual_network_id"] = *v.Value
		}
	}

	return []interface{}{parameters}
}

func expandWorkspaceCustomParameters(input []interface{}) *databricks.WorkspaceCustomParameters {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	config := input[0].(map[string]interface{})
	parameters := databricks.WorkspaceCustomParameters{}

	if v, ok := config["no_public_ip"].(bool); ok {
		parameters.EnableNoPublicIP = &databricks.WorkspaceCustomBooleanParameter{
			Value: &v,
		}
	}

	if v := config["public_subnet_name"].(string); v != "" {
		parameters.CustomPublicSubnetName = &databricks.WorkspaceCustomStringParameter{
			Value: &v,
		}
	}

	if v := config["private_subnet_name"].(string); v != "" {
		parameters.CustomPrivateSubnetName = &databricks.WorkspaceCustomStringParameter{
			Value: &v,
		}
	}

	if v := config["virtual_network_id"].(string); v != "" {
		parameters.CustomVirtualNetworkID = &databricks.WorkspaceCustomStringParameter{
			Value: &v,
		}
	}

	return &parameters
}
