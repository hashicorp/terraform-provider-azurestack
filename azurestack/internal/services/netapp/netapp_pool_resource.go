package netapp

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/netapp/mgmt/2020-09-01/netapp"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/netapp/parse"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/netapp/validate"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/timeouts"
)

func resourceNetAppPool() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceNetAppPoolCreateUpdate,
		Read:   resourceNetAppPoolRead,
		Update: resourceNetAppPoolCreateUpdate,
		Delete: resourceNetAppPoolDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.CapacityPoolID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.PoolName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AccountName,
			},

			"service_level": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(netapp.Premium),
					string(netapp.Standard),
					string(netapp.Ultra),
				}, true),
			},

			"size_in_tb": {
				Type:         pluginsdk.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(4, 500),
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceNetAppPoolCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.PoolClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	accountName := d.Get("account_name").(string)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, accountName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for present of existing NetApp Pool %q (Resource Group %q): %+v", name, resourceGroup, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_netapp_pool", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	serviceLevel := d.Get("service_level").(string)
	sizeInTB := int64(d.Get("size_in_tb").(int))
	sizeInMB := sizeInTB * 1024 * 1024
	sizeInBytes := sizeInMB * 1024 * 1024

	capacityPoolParameters := netapp.CapacityPool{
		Location: utils.String(location),
		PoolProperties: &netapp.PoolProperties{
			ServiceLevel: netapp.ServiceLevel(serviceLevel),
			Size:         utils.Int64(sizeInBytes),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.CreateOrUpdate(ctx, capacityPoolParameters, resourceGroup, accountName, name)
	if err != nil {
		return fmt.Errorf("Error creating NetApp Pool %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation of NetApp Pool %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, accountName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving NetApp Pool %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if resp.ID == nil {
		return fmt.Errorf("Cannot read NetApp Pool %q (Resource Group %q) ID", name, resourceGroup)
	}
	d.SetId(*resp.ID)

	return resourceNetAppPoolRead(d, meta)
}

func resourceNetAppPoolRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.PoolClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CapacityPoolID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.NetAppAccountName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] NetApp Pools %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading NetApp Pools %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("account_name", id.NetAppAccountName)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	if poolProperties := resp.PoolProperties; poolProperties != nil {
		d.Set("service_level", poolProperties.ServiceLevel)

		sizeInTB := int64(0)
		if poolProperties.Size != nil {
			sizeInBytes := *poolProperties.Size
			sizeInMB := sizeInBytes / 1024 / 1024
			sizeInTB = sizeInMB / 1024 / 1024
		}
		d.Set("size_in_tb", int(sizeInTB))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceNetAppPoolDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).NetApp.PoolClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CapacityPoolID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.Delete(ctx, id.ResourceGroup, id.NetAppAccountName, id.Name); err != nil {
		return fmt.Errorf("Error deleting NetApp Pool %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	// The resource NetApp Pool depends on the resource NetApp Account.
	// Although the delete API returns 404 which means the NetApp Pool resource has been deleted.
	// Then it tries to immediately delete NetApp Account but it still throws error `Can not delete resource before nested resources are deleted.`
	// In this case we're going to re-check status code again.
	// For more details, see related Bug: https://github.com/Azure/azure-sdk-for-go/issues/6374
	log.Printf("[DEBUG] Waiting for NetApp Pool %q (Resource Group %q) to be deleted", id.Name, id.ResourceGroup)
	stateConf := &pluginsdk.StateChangeConf{
		ContinuousTargetOccurence: 5,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"200", "202"},
		Target:                    []string{"204", "404"},
		Refresh:                   netappPoolDeleteStateRefreshFunc(ctx, client, id.ResourceGroup, id.NetAppAccountName, id.Name),
		Timeout:                   d.Timeout(pluginsdk.TimeoutDelete),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for NetApp Pool %q (Resource Group %q) to be deleted: %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func netappPoolDeleteStateRefreshFunc(ctx context.Context, client *netapp.PoolsClient, resourceGroupName string, accountName string, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		res, err := client.Get(ctx, resourceGroupName, accountName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(res.Response) {
				return nil, "", fmt.Errorf("Error retrieving NetApp Pool %q (Resource Group %q): %s", name, resourceGroupName, err)
			}
		}

		return res, strconv.Itoa(res.StatusCode), nil
	}
}
