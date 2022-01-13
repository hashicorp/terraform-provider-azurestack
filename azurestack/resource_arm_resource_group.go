package azurestack

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/azurestack/helpers/pointer"
	"github.com/hashicorp/terraform-provider-azurestack/azurestack/helpers/response"
)

func resourceArmResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmResourceGroupCreateUpdate,
		Read:   resourceArmResourceGroupRead,
		Update: resourceArmResourceGroupCreateUpdate,
		Delete: resourceArmResourceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": resourceGroupNameSchema(),

			"location": locationSchema(),

			"tags": tagsSchema(),
		},
	}
}

func resourceArmResourceGroupCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).resourceGroupsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	tags := d.Get("tags").(map[string]interface{})
	parameters := resources.Group{
		Location: pointer.FromString(location),
		Tags:     *expandTags(tags),
	}
	_, err := client.CreateOrUpdate(ctx, name, parameters)
	if err != nil {
		return fmt.Errorf("creating resource group: %+v", err)
	}

	resp, err := client.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("retrieving resource group: %+v", err)
	}

	d.SetId(*resp.ID)

	return resourceArmResourceGroupRead(d, meta)
}

func resourceArmResourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).resourceGroupsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing Azure Resource ID %q: %+v", d.Id(), err)
	}

	name := id.ResourceGroup

	resp, err := client.Get(ctx, name)
	if err != nil {
		if response.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading resource group %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading resource group: %+v", err)
	}

	d.Set("name", resp.Name)
	d.Set("location", azureStackNormalizeLocation(*resp.Location))
	flattenAndSetTags(d, &resp.Tags)

	return nil
}

func resourceArmResourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).resourceGroupsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing Azure Resource ID %q: %+v", d.Id(), err)
	}

	name := id.ResourceGroup

	deleteFuture, err := client.Delete(ctx, name)
	if err != nil {
		if response.WasNotFound(deleteFuture.Response()) {
			return nil
		}

		return fmt.Errorf("deleting Resource Group %q: %+v", name, err)
	}

	err = deleteFuture.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		if response.WasNotFound(deleteFuture.Response()) {
			return nil
		}

		return fmt.Errorf("deleting Resource Group %q: %+v", name, err)
	}

	return nil
}
