package apimanagement

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/apimanagement/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/apimanagement/schemaz"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourceApiManagementApiVersionSet() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceApiManagementApiVersionSetRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": schemaz.SchemaApiManagementChildDataSourceName(),

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"api_management_name": schemaz.SchemaApiManagementDataSourceName(),

			"description": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"versioning_scheme": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"version_header_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"version_query_name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceApiManagementApiVersionSetRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ApiVersionSetClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("api_management_name").(string)

	resp, err := client.Get(ctx, resourceGroup, serviceName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf(": API Version Set %q (API Management Service %q / Resource Group %q) does not exist!", name, serviceName, resourceGroup)
		}

		return fmt.Errorf("reading API Version Set %q (API Management Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("retrieving API Version Set %q (API Management Service %q /Resource Group %q): ID was nil or empty", name, serviceName, resourceGroup)
	}

	id, err := parse.ApiVersionSetID(*resp.ID)
	if err != nil {
		return fmt.Errorf("Error parsing API Version Set ID: %q", *resp.ID)
	}

	d.SetId(id.ID())

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("api_management_name", serviceName)
	if props := resp.APIVersionSetContractProperties; props != nil {
		d.Set("description", props.Description)
		d.Set("display_name", props.DisplayName)
		d.Set("versioning_scheme", string(props.VersioningScheme))
		d.Set("version_header_name", props.VersionHeaderName)
		d.Set("version_query_name", props.VersionQueryName)
	}

	return nil
}
