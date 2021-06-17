package iothub

import (
	"fmt"
	"regexp"
	"time"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/iothub/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func dataSourceIotHubSharedAccessPolicy() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Read: dataSourceIotHubSharedAccessPolicyRead,

		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9!._-]{1,64}`), ""+
					"The shared access policy key name must not be empty, and must not exceed 64 characters in length.  The shared access policy key name can only contain alphanumeric characters, exclamation marks, periods, underscores and hyphens."),
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"iothub_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.IoTHubName,
			},

			"primary_key": {
				Type:      pluginsdk.TypeString,
				Sensitive: true,
				Computed:  true,
			},

			"primary_connection_string": {
				Type:      pluginsdk.TypeString,
				Sensitive: true,
				Computed:  true,
			},

			"secondary_key": {
				Type:      pluginsdk.TypeString,
				Sensitive: true,
				Computed:  true,
			},

			"secondary_connection_string": {
				Type:      pluginsdk.TypeString,
				Sensitive: true,
				Computed:  true,
			},
		},
	}
}

func dataSourceIotHubSharedAccessPolicyRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).IoTHub.ResourceClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	iothubName := d.Get("iothub_name").(string)

	accessPolicy, err := client.GetKeysForKeyName(ctx, resourceGroup, iothubName, name)
	if err != nil {
		if utils.ResponseWasNotFound(accessPolicy.Response) {
			return fmt.Errorf("Error: IotHub %q (Resource Group %q) was not found", name, resourceGroup)
		}

		return fmt.Errorf("Error loading IotHub Shared Access Policy %q (Resource Group %q): %+v", iothubName, resourceGroup, err)
	}

	iothub, err := client.Get(ctx, resourceGroup, iothubName)
	if err != nil {
		return fmt.Errorf("Error loading IotHub %q (Resource Group %q): %+v", iothubName, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("iothub_name", iothubName)
	d.Set("resource_group_name", resourceGroup)

	resourceID := fmt.Sprintf("%s/IotHubKeys/%s", *iothub.ID, name)
	d.SetId(resourceID)

	d.Set("primary_key", accessPolicy.PrimaryKey)
	if err := d.Set("primary_connection_string", getSharedAccessPolicyConnectionString(*iothub.Properties.HostName, name, *accessPolicy.PrimaryKey)); err != nil {
		return fmt.Errorf("error setting `primary_connection_string`: %v", err)
	}
	d.Set("secondary_key", accessPolicy.SecondaryKey)
	if err := d.Set("secondary_connection_string", getSharedAccessPolicyConnectionString(*iothub.Properties.HostName, name, *accessPolicy.SecondaryKey)); err != nil {
		return fmt.Errorf("error setting `secondary_connection_string`: %v", err)
	}

	return nil
}
