package storage

// TODO - bring in line with the azurerm version of this resource

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
)

func storageContainer() *schema.Resource {
	return &schema.Resource{
		Create: storageContainerCreate,
		Read:   storageContainerRead,
		Delete: storageContainerDelete,

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageContainerName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"storage_account_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageAccountName,
			},

			"container_access_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "private",
				// todo validate correctly
				/*ValidateFunc: validation.StringInSlice([]string{
					string(containers.Blob),
					string(containers.Container),
					"private",
				}, false),*/
			},

			// todo: it doesn't appear this was set in the old stack provider?
			"properties": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func storageContainerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)

	blobClient, accountExists, err := client.GetBlobStorageClientForStorageAccount(ctx, resourceGroupName, storageAccountName)
	if err != nil {
		return err
	}
	if !accountExists {
		return fmt.Errorf("Storage Account %q Not Found", storageAccountName)
	}

	name := d.Get("name").(string)

	var accessType storage.ContainerAccessType
	if d.Get("container_access_type").(string) == "private" {
		accessType = storage.ContainerAccessType("")
	} else {
		accessType = storage.ContainerAccessType(d.Get("container_access_type").(string))
	}

	log.Printf("[INFO] Creating container %q in storage account %q.", name, storageAccountName)
	reference := blobClient.GetContainerReference(name)

	err = resource.RetryContext(ctx, 120*time.Second, checkContainerIsCreated(reference))
	if err != nil {
		return fmt.Errorf("creating container %q in storage account %q: %s", name, storageAccountName, err)
	}

	permissions := storage.ContainerPermissions{
		AccessType: accessType,
	}
	permissionOptions := &storage.SetContainerPermissionOptions{}
	err = reference.SetPermissions(permissions, permissionOptions)
	if err != nil {
		return fmt.Errorf("setting permissions for container %s in storage account %s: %+v", name, storageAccountName, err)
	}

	d.SetId(name)
	return storageContainerRead(d, meta)
}

func checkContainerIsCreated(reference *storage.Container) func() *resource.RetryError {
	return func() *resource.RetryError {
		createOptions := &storage.CreateContainerOptions{}
		_, err := reference.CreateIfNotExists(createOptions)
		if err != nil {
			return resource.RetryableError(err)
		}

		return nil
	}
}

// resourceAzureStorageContainerRead does all the necessary API calls to
// read the status of the storage container off Azure.
func storageContainerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)

	blobClient, accountExists, err := client.GetBlobStorageClientForStorageAccount(ctx, resourceGroupName, storageAccountName)
	if err != nil {
		return err
	}
	if !accountExists {
		log.Printf("[DEBUG] Storage account %q not found, removing container %q from state", storageAccountName, d.Id())
		d.SetId("")
		return nil
	}

	name := d.Get("name").(string)
	containers, err := blobClient.ListContainers(storage.ListContainersParameters{
		Prefix:  name,
		Timeout: 90,
	})
	if err != nil {
		return fmt.Errorf("Failed to retrieve storage containers in account %q: %s", name, err)
	}

	var found bool
	for _, cont := range containers.Containers {
		if cont.Name == name {
			found = true

			props := make(map[string]interface{})
			props["last_modified"] = cont.Properties.LastModified
			props["lease_status"] = cont.Properties.LeaseStatus
			props["lease_state"] = cont.Properties.LeaseState
			props["lease_duration"] = cont.Properties.LeaseDuration

			d.Set("properties", props)
		}
	}

	if !found {
		log.Printf("[INFO] Storage container %q does not exist in account %q, removing from state...", name, storageAccountName)
		d.SetId("")
	}

	return nil
}

// resourceAzureStorageContainerDelete does all the necessary API calls to
// delete a storage container off Azure.
func storageContainerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	storageAccountName := d.Get("storage_account_name").(string)

	blobClient, accountExists, err := client.GetBlobStorageClientForStorageAccount(ctx, resourceGroupName, storageAccountName)
	if err != nil {
		return err
	}
	if !accountExists {
		log.Printf("[INFO]Storage Account %q doesn't exist so the container won't exist", storageAccountName)
		return nil
	}

	name := d.Get("name").(string)

	log.Printf("[INFO] Deleting storage container %q in account %q", name, storageAccountName)
	reference := blobClient.GetContainerReference(name)
	deleteOptions := &storage.DeleteContainerOptions{}
	if _, err := reference.DeleteIfExists(deleteOptions); err != nil {
		return fmt.Errorf("deleting storage container %q from storage account %q: %s", name, storageAccountName, err)
	}

	d.SetId("")
	return nil
}
