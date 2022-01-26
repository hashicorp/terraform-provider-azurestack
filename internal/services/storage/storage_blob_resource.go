package storage

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/migration"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
	"github.com/tombuildsstuff/giovanni/storage/2018-11-09/blob/blobs"
)

func storageBlob() *schema.Resource {
	return &schema.Resource{
		Create: storageBlobCreate,
		Read:   storageBlobRead,
		Delete: storageBlobDelete,

		// TODO check schema and confirm old stack provider can upgrade to this
		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.BlobV0ToV1{},
		}),

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO: add validation
			},

			// "resource_group_name": commonschema.ResourceGroupName(),
			// todo property deprecate

			"storage_account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageAccountName,
			},

			"storage_container_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageContainerName,
			},

			"type": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Append",
					"Block",
					"Page",
				}, false),
			},

			"size": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntDivisibleBy(512),
			},

			"source": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_uri", "source_content"},
			},

			"source_content": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source", "source_uri"},
			},

			"source_uri": {
				Type:          pluginsdk.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source", "source_content"},
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"parallelism": {
				// TODO: @tombuildsstuff - a note this only works for Page blobs
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Default:      8,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			// removed in new storage sdk
			/*"attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},*/
		},
	}
}

func storageBlobCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	accountName := d.Get("storage_account_name").(string)
	containerName := d.Get("storage_container_name").(string)
	name := d.Get("name").(string)

	account, err := storageClient.FindAccount(ctx, accountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Blob %q (Container %q): %s", accountName, name, containerName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", accountName)
	}

	blobsClient, err := storageClient.BlobsClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Blobs Client: %s", err)
	}

	id := blobsClient.GetResourceID(accountName, containerName, name)
	if d.IsNewResource() {
		input := blobs.GetPropertiesInput{}
		props, err := blobsClient.GetProperties(ctx, accountName, containerName, name, input)
		if err != nil {
			if !utils.ResponseWasNotFound(props.Response) {
				return fmt.Errorf("checking if Blob %q exists (Container %q / Account %q / Resource Group %q): %s", name, containerName, accountName, account.ResourceGroup, err)
			}
		}
		if !utils.ResponseWasNotFound(props.Response) {
			return tf.ImportAsExistsError("azurestack_storage_blob", id)
		}
	}

	log.Printf("[DEBUG] Creating Blob %q in Container %q within Storage Account %q..", name, containerName, accountName)
	blobInput := BlobUpload{
		AccountName:   accountName,
		ContainerName: containerName,
		BlobName:      name,
		Client:        blobsClient,

		BlobType:      d.Get("type").(string),
		Parallelism:   d.Get("parallelism").(int),
		Size:          d.Get("size").(int),
		Source:        d.Get("source").(string),
		SourceContent: d.Get("source_content").(string),
		SourceUri:     d.Get("source_uri").(string),
	}
	if err := blobInput.Create(ctx); err != nil {
		return fmt.Errorf("creating Blob %q (Container %q / Account %q): %s", name, containerName, accountName, err)
	}
	log.Printf("[DEBUG] Created Blob %q in Container %q within Storage Account %q.", name, containerName, accountName)

	d.SetId(id)

	return storageBlobUpdate(d, meta)
}

func storageBlobUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := blobs.ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing %q: %s", d.Id(), err)
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Blob %q (Container %q): %s", id.AccountName, id.BlobName, id.ContainerName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}

	/*	blobsClient, err := storageClient.BlobsClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("building Blobs Client: %s", err)
		}
	*/
	return storageBlobRead(d, meta)
}

func storageBlobRead(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := blobs.ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing %q: %s", d.Id(), err)
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Blob %q (Container %q): %s", id.AccountName, id.BlobName, id.ContainerName, err)
	}
	if account == nil {
		log.Printf("[DEBUG] Unable to locate Account %q for Blob %q (Container %q) - assuming removed & removing from state!", id.AccountName, id.BlobName, id.ContainerName)
		d.SetId("")
		return nil
	}

	blobsClient, err := storageClient.BlobsClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Blobs Client: %s", err)
	}

	log.Printf("[INFO] Retrieving Storage Blob %q (Container %q / Account %q).", id.BlobName, id.ContainerName, id.AccountName)
	input := blobs.GetPropertiesInput{}
	props, err := blobsClient.GetProperties(ctx, id.AccountName, id.ContainerName, id.BlobName, input)
	if err != nil {
		if utils.ResponseWasNotFound(props.Response) {
			log.Printf("[INFO] Blob %q was not found in Container %q / Account %q - assuming removed & removing from state...", id.BlobName, id.ContainerName, id.AccountName)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving properties for Blob %q (Container %q / Account %q): %s", id.BlobName, id.ContainerName, id.AccountName, err)
	}

	d.Set("name", id.BlobName)
	d.Set("storage_container_name", id.ContainerName)
	d.Set("storage_account_name", id.AccountName)

	d.Set("type", strings.TrimSuffix(string(props.BlobType), "Blob"))
	d.Set("url", d.Id())

	// The CopySource is only returned if the blob hasn't been modified (e.g. metadata configured etc)
	// as such, we need to conditionally set this to ensure it's trackable if possible
	if props.CopySource != "" {
		d.Set("source_uri", props.CopySource)
	}

	return nil
}

func storageBlobDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := blobs.ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("parsing %q: %s", d.Id(), err)
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("retrieving Account %q for Blob %q (Container %q): %s", id.AccountName, id.BlobName, id.ContainerName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}

	blobsClient, err := storageClient.BlobsClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building Blobs Client: %s", err)
	}

	log.Printf("[INFO] Deleting Blob %q from Container %q / Storage Account %q", id.BlobName, id.ContainerName, id.AccountName)
	input := blobs.DeleteInput{
		DeleteSnapshots: true,
	}
	if _, err := blobsClient.Delete(ctx, id.AccountName, id.ContainerName, id.BlobName, input); err != nil {
		return fmt.Errorf("deleting Blob %q (Container %q / Account %q): %s", id.BlobName, id.ContainerName, id.AccountName, err)
	}

	return nil
}
