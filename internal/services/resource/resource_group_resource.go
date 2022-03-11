package resource

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2020-09-01/resources/mgmt/resources"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/resource/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

func resourceGroup() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceGroupCreateUpdate,
		Read:   resourceGroupRead,
		Update: resourceGroupCreateUpdate,
		Delete: resourceGroupDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ResourceGroupID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(90 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(90 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": commonschema.ResourceGroupName(),

			"location": commonschema.Location(),

			"tags": commonschema.Tags(),
		},
	}
}

func resourceGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	if d.IsNewResource() {
		existing, err := client.Get(ctx, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing resource group: %+v", err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurestack_resource_group", *existing.ID)
		}
	}

	parameters := resources.Group{
		Location: pointer.FromString(location),
		Tags:     tags.Expand(t),
	}

	if _, err := client.CreateOrUpdate(ctx, name, parameters); err != nil {
		return fmt.Errorf("creating Resource Group %q: %+v", name, err)
	}

	resp, err := client.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("retrieving Resource Group %q: %+v", name, err)
	}

	// @tombuildsstuff: intentionally leaving this for now, since this'll need
	// details in the upgrade notes given how the Resource Group ID is cased incorrectly
	// but needs to be fixed (resourcegroups -> resourceGroups)
	d.SetId(*resp.ID)

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading resource group %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading resource group: %+v", err)
	}

	d.Set("name", resp.Name)
	d.Set("location", location.NormalizeNilable(resp.Location))
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceGroupDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Resource.GroupsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGroupID(d.Id())
	if err != nil {
		return err
	}

	// conditionally check for nested resources and error if they exist
	// TODO upgrade and enable
	/*if meta.(*clients.Client).Features.ResourceGroup.PreventDeletionIfContainsResources {
		resourceClient := meta.(*clients.Client).Resource.ResourcesClient
		results, err := resourceClient.ListByResourceGroupComplete(ctx, id.ResourceGroup, "", "", utils.Int32(500))
		if err != nil {
			return fmt.Errorf("listing resources in %s: %v", *id, err)
		}
		nestedResourceIds := make([]string, 0)
		for results.NotDone() {
			val := results.Value()
			if val.ID != nil {
				nestedResourceIds = append(nestedResourceIds, *val.ID)
			}

			if err := results.NextWithContext(ctx); err != nil {
				return fmt.Errorf("retrieving next page of nested items for %s: %+v", id, err)
			}
		}

		if len(nestedResourceIds) > 0 {
			return resourceGroupContainsItemsError(id.ResourceGroup, nestedResourceIds)
		}
	}*/

	deleteFuture, err := client.Delete(ctx, id.ResourceGroup)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	err = deleteFuture.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return fmt.Errorf("waiting for the deletion of %s: %+v", *id, err)
	}

	return nil
}
