package loganalytics

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/operationalinsights/mgmt/2020-08-01/operationalinsights"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/loganalytics/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/loganalytics/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceLogAnalyticsSavedSearch() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceLogAnalyticsSavedSearchCreate,
		Read:   resourceLogAnalyticsSavedSearchRead,
		Delete: resourceLogAnalyticsSavedSearchDelete,

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"log_analytics_workspace_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LogAnalyticsWorkspaceID,
				// https://github.com/Azure/azure-rest-api-specs/issues/9330
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				// https://github.com/Azure/azure-rest-api-specs/issues/9330
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"category": {
				Type:         pluginsdk.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"display_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"query": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"function_alias": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"function_parameters": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
					ValidateFunc: validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9!-_]*:[a-zA-Z0-9!_-]+=[a-zA-Z0-9!_-]+|^[a-zA-Z0-9!-_]*:[a-zA-Z0-9!_-]+`),
						"Log Analytics Saved Search Function Parameters must be in the following format: param-name1:type1=default_value1 OR param-name1:type1 OR param-name1:string='string goes here'",
					),
				},
			},

			"tags": tags.ForceNewSchema(),
		},
	}
}

func resourceLogAnalyticsSavedSearchCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.SavedSearchesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM Log Analytics Saved Search creation.")

	name := d.Get("name").(string)
	workspaceID := d.Get("log_analytics_workspace_id").(string)
	id, err := parse.LogAnalyticsWorkspaceID(workspaceID)
	if err != nil {
		return err
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Log Analytics Saved Search %q (WorkSpace %q / Resource Group %q): %s", name, id.WorkspaceName, id.ResourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_log_analytics_saved_search", *existing.ID)
		}
	}

	parameters := operationalinsights.SavedSearch{
		SavedSearchProperties: &operationalinsights.SavedSearchProperties{
			Category:      utils.String(d.Get("category").(string)),
			DisplayName:   utils.String(d.Get("display_name").(string)),
			Query:         utils.String(d.Get("query").(string)),
			FunctionAlias: utils.String(d.Get("function_alias").(string)),
			Tags:          expandSavedSearchTag(d.Get("tags").(map[string]interface{})), // expand tags because it's defined as object set in service
		},
	}

	if v, ok := d.GetOk("function_parameters"); ok {
		attrs := v.(*pluginsdk.Set).List()
		result := make([]string, 0)
		for _, item := range attrs {
			if item != nil {
				result = append(result, item.(string))
			}
		}
		parameters.SavedSearchProperties.FunctionParameters = utils.String(strings.Join(result, ", "))
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.WorkspaceName, name, parameters); err != nil {
		return fmt.Errorf("creating Saved Search %q (Log Analytics Workspace %q / Resource Group %q): %+v", name, id.WorkspaceName, id.ResourceGroup, err)
	}

	read, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, name)
	if err != nil {
		return fmt.Errorf("retrieving Saved Search %q (Log Analytics Workspace %q / Resource Group %q): %+v", name, id.WorkspaceName, id.ResourceGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("cannot read Log Analytics Saved Search %q (WorkSpace %q / Resource Group %q): %s", name, id.WorkspaceName, id.ResourceGroup, err)
	}

	d.SetId(*read.ID)

	return resourceLogAnalyticsSavedSearchRead(d, meta)
}

func resourceLogAnalyticsSavedSearchRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.SavedSearchesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()
	// FIXME: @favoretti: API returns ID without a leading slash
	id, err := parse.LogAnalyticsSavedSearchID(fmt.Sprintf("/%s", strings.TrimPrefix(d.Id(), "/")))
	if err != nil {
		return err
	}
	workspaceId := parse.NewLogAnalyticsWorkspaceID(subscriptionId, id.ResourceGroup, id.WorkspaceName).ID()

	resp, err := client.Get(ctx, id.ResourceGroup, id.WorkspaceName, id.SavedSearcheName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Saved Search %q (Log Analytics Workspace %q / Resource Group %q): %s", id.WorkspaceName, id.WorkspaceName, id.ResourceGroup, err)
	}

	d.Set("name", id.SavedSearcheName)
	d.Set("log_analytics_workspace_id", workspaceId)

	if props := resp.SavedSearchProperties; props != nil {
		d.Set("display_name", props.DisplayName)
		d.Set("category", props.Category)
		d.Set("query", props.Query)
		d.Set("function_alias", props.FunctionAlias)
		functionParams := make([]string, 0)
		if props.FunctionParameters != nil {
			functionParams = strings.Split(*props.FunctionParameters, ", ")
		}
		d.Set("function_parameters", functionParams)

		// flatten tags because it's defined as object set in service
		if err := d.Set("tags", flattenSavedSearchTag(props.Tags)); err != nil {
			return fmt.Errorf("setting `tag`: %+v", err)
		}
	}

	return nil
}

func resourceLogAnalyticsSavedSearchDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).LogAnalytics.SavedSearchesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()
	// FIXME: @favoretti: API returns ID without a leading slash
	id, err := parse.LogAnalyticsSavedSearchID(fmt.Sprintf("/%s", strings.TrimPrefix(d.Id(), "/")))
	if err != nil {
		return err
	}

	if _, err = client.Delete(ctx, id.ResourceGroup, id.WorkspaceName, id.SavedSearcheName); err != nil {
		return fmt.Errorf("deleting Saved Search %q (Log Analytics Workspace %q / Resource Group %q): %s", id.WorkspaceName, id.WorkspaceName, id.ResourceGroup, err)
	}

	return nil
}

func expandSavedSearchTag(input map[string]interface{}) *[]operationalinsights.Tag {
	results := make([]operationalinsights.Tag, 0)
	for key, value := range input {
		result := operationalinsights.Tag{
			Name:  utils.String(key),
			Value: utils.String(value.(string)),
		}
		results = append(results, result)
	}
	return &results
}

func flattenSavedSearchTag(input *[]operationalinsights.Tag) map[string]interface{} {
	results := make(map[string]interface{})
	if input == nil {
		return results
	}

	for _, item := range *input {
		var key string
		if item.Name != nil {
			key = *item.Name
		}
		var value string
		if item.Value != nil {
			value = *item.Value
		}
		results[key] = value
	}
	return results
}
