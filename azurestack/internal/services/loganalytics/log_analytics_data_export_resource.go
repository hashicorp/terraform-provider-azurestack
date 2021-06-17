package loganalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/mgmt/2020-08-01/operationalinsights"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/loganalytics/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/loganalytics/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceLogAnalyticsDataExport() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceOperationalinsightsDataExportCreateUpdate,
		Read:   resourceOperationalinsightsDataExportRead,
		Update: resourceOperationalinsightsDataExportCreateUpdate,
		Delete: resourceOperationalinsightsDataExportDelete,
		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validate.LogAnalyticsDataExportName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"workspace_resource_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LogAnalyticsWorkspaceID,
			},

			"destination_resource_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"table_names": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Schema{
					Type:         pluginsdk.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},

			"export_rule_id": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOperationalinsightsDataExportCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.DataExportClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	workspace, err := parse.LogAnalyticsWorkspaceID(d.Get("workspace_resource_id").(string))
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, workspace.WorkspaceName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing Log Analytics Data Export Rule %q (Resource Group %q / workspaceName %q): %+v", name, resourceGroup, workspace.WorkspaceName, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_log_analytics_data_export_rule", *existing.ID)
		}
	}

	parameters := operationalinsights.DataExport{
		DataExportProperties: &operationalinsights.DataExportProperties{
			Destination: &operationalinsights.Destination{
				ResourceID: utils.String(d.Get("destination_resource_id").(string)),
			},
			TableNames: utils.ExpandStringSlice(d.Get("table_names").(*pluginsdk.Set).List()),
			Enable:     utils.Bool(d.Get("enabled").(bool)),
		},
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, workspace.WorkspaceName, name, parameters); err != nil {
		return fmt.Errorf("creating/updating Log Analytics Data Export Rule %q (Resource Group %q / workspaceName %q): %+v", name, resourceGroup, workspace.WorkspaceName, err)
	}

	resp, err := client.Get(ctx, resourceGroup, workspace.WorkspaceName, name)
	if err != nil {
		return fmt.Errorf("retrieving Log Analytics Data Export Rule %q (Resource Group %q / workspaceName %q): %+v", name, resourceGroup, workspace.WorkspaceName, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Log Analytics Data Export Rule %q (Resource Group %q / workspaceName %q) ID", name, resourceGroup, workspace.WorkspaceName)
	}

	d.SetId(*resp.ID)
	return resourceOperationalinsightsDataExportRead(d, meta)
}

func resourceOperationalinsightsDataExportRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.DataExportClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LogAnalyticsDataExportID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.DataexportName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Log Analytics %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Log Analytics Data Export Rule %q (Resource Group %q / workspaceName %q): %+v", id.DataexportName, id.ResourceGroup, id.WorkspaceName, err)
	}
	d.Set("name", id.DataexportName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("workspace_resource_id", parse.NewLogAnalyticsWorkspaceID(id.SubscriptionId, id.ResourceGroup, id.WorkspaceName).ID())
	if props := resp.DataExportProperties; props != nil {
		d.Set("export_rule_id", props.DataExportID)
		d.Set("destination_resource_id", flattenDataExportDestination(props.Destination))
		d.Set("enabled", props.Enable)
		d.Set("table_names", utils.FlattenStringSlice(props.TableNames))
	}
	return nil
}

func resourceOperationalinsightsDataExportDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.DataExportClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LogAnalyticsDataExportID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.WorkspaceName, id.DataexportName); err != nil {
		return fmt.Errorf("deleting Log Analytics Data Export Rule %q (Resource Group %q / workspaceName %q): %+v", id.DataexportName, id.ResourceGroup, id.WorkspaceName, err)
	}
	return nil
}

func flattenDataExportDestination(input *operationalinsights.Destination) string {
	if input == nil {
		return ""
	}

	var resourceID string
	if input.ResourceID != nil {
		resourceID = *input.ResourceID
	}

	return resourceID
}
