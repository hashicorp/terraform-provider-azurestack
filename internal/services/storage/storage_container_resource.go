package storage

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tombuildsstuff/giovanni/storage/2018-11-09/blob/containers"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/migration"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
)

func resourceStorageContainer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStorageContainerCreate,
		ReadContext:   resourceStorageContainerRead,
		DeleteContext: resourceStorageContainerDelete,
		UpdateContext: resourceStorageContainerUpdate,

		// TODO check schema and confirm old stack provider can upgrade to this
		SchemaVersion: 1,
		StateUpgraders: pluginsdk.StateUpgrades(map[int]pluginsdk.StateUpgrade{
			0: migration.ContainerV0ToV1{},
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		// TODO: replace this with an importer which validates the ID during import
		Importer: pluginsdk.DefaultImporter(),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageContainerName,
			},
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
				ValidateFunc: validation.StringInSlice([]string{
					string(containers.Blob),
					string(containers.Container),
					"private",
				}, false),
			},
			"metadata": {
				Type:         pluginsdk.TypeMap,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.MetaDataKeys,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
			"has_immutability_policy": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},
			"has_legal_hold": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},
		},
	}
}

func expandStorageContainerAccessLevel(input string) containers.AccessLevel {
	// for historical reasons, "private" above is an empty string in the API
	// so the enum doesn't 1:1 match. You could argue the SDK should handle this
	// but this is suitable for now
	if input == "private" {
		return containers.Private
	}

	return containers.AccessLevel(input)
}

func resourceStorageContainerCreate(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) diag.Diagnostics {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	containerName := d.Get("name").(string)
	accountName := d.Get("storage_account_name").(string)
	accessLevelRaw := d.Get("container_access_type").(string)
	accessLevel := expandStorageContainerAccessLevel(accessLevelRaw)

	metaDataRaw := d.Get("metadata").(map[string]interface{})
	metaData := ExpandMetaData(metaDataRaw)

	account, err := storageClient.FindAccount(ctx, accountName)
	if err != nil {
		return diag.Errorf("retrieving Account %q for Container %q: %s", accountName, containerName, err)
	}
	if account == nil {
		return diag.Errorf("Unable to locate Storage Account %q!", accountName)
	}

	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return diag.Errorf("building storage client: %+v", err)
	}

	id := parse.NewStorageContainerDataPlaneId(accountName, storageClient.Env.StorageEndpointSuffix, containerName).ID()
	exists, err := client.Exists(ctx, account.ResourceGroup, accountName, containerName)
	if err != nil {
		return diag.FromErr(err)
	}
	if exists != nil && *exists {
		return diag.FromErr(tf.ImportAsExistsError("azurestack_storage_container", id))
	}

	log.Printf("[INFO] Creating Container %q in Storage Account %q", containerName, accountName)
	input := containers.CreateInput{
		AccessLevel: accessLevel,
		MetaData:    metaData,
	}

	if err := client.Create(ctx, account.ResourceGroup, accountName, containerName, input); err != nil {
		return diag.Errorf("failed creating container: %+v", err)
	}

	d.SetId(id)
	return resourceStorageContainerRead(ctx, d, meta)
}

func resourceStorageContainerRead(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) diag.Diagnostics {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageContainerDataPlaneID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return diag.Errorf("retrieving Account %q for Container %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		log.Printf("[DEBUG] Unable to locate Account %q for Storage Container %q - assuming removed & removing from state", id.AccountName, id.Name)
		d.SetId("")
		return nil
	}
	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return diag.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	props, err := client.Get(ctx, account.ResourceGroup, id.AccountName, id.Name)
	if err != nil {
		return diag.Errorf("retrieving Container %q (Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
	}
	if props == nil {
		log.Printf("[DEBUG] Container %q was not found in Account %q / Resource Group %q - assuming removed & removing from state", id.Name, id.AccountName, account.ResourceGroup)
		d.SetId("")
		return nil
	}

	d.Set("name", id.Name)
	d.Set("storage_account_name", id.AccountName)
	d.Set("container_access_type", flattenStorageContainerAccessLevel(props.AccessLevel))

	if err := d.Set("metadata", FlattenMetaData(props.MetaData)); err != nil {
		return diag.Errorf("setting `metadata`: %+v", err)
	}

	d.Set("has_immutability_policy", props.HasImmutabilityPolicy)
	d.Set("has_legal_hold", props.HasLegalHold)

	return nil
}

func flattenStorageContainerAccessLevel(input containers.AccessLevel) string {
	// for historical reasons, "private" above is an empty string in the API
	if input == containers.Private {
		return "private"
	}

	return string(input)
}

func resourceStorageContainerDelete(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) diag.Diagnostics {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageContainerDataPlaneID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return diag.Errorf("retrieving Account %q for Container %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		return diag.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}
	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return diag.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	if err := client.Delete(ctx, account.ResourceGroup, id.AccountName, id.Name); err != nil {
		return diag.Errorf("deleting Container %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
	}

	return nil
}

func resourceStorageContainerUpdate(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) diag.Diagnostics {
	storageClient := meta.(*clients.Client).Storage
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageContainerDataPlaneID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return diag.Errorf("retrieving Account %q for Container %q: %s", id.AccountName, id.Name, err)
	}
	if account == nil {
		return diag.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}
	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return diag.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	if d.HasChange("container_access_type") {
		log.Printf("[DEBUG] Updating the Access Control for Container %q (Storage Account %q / Resource Group %q)..", id.Name, id.AccountName, account.ResourceGroup)
		accessLevelRaw := d.Get("container_access_type").(string)
		accessLevel := expandStorageContainerAccessLevel(accessLevelRaw)

		if err := client.UpdateAccessLevel(ctx, account.ResourceGroup, id.AccountName, id.Name, accessLevel); err != nil {
			return diag.Errorf("updating the Access Control for Container %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
		}

		log.Printf("[DEBUG] Updated the Access Control for Container %q (Storage Account %q / Resource Group %q)", id.Name, id.AccountName, account.ResourceGroup)
	}

	if d.HasChange("metadata") {
		log.Printf("[DEBUG] Updating the MetaData for Container %q (Storage Account %q / Resource Group %q)..", id.Name, id.AccountName, account.ResourceGroup)
		metaDataRaw := d.Get("metadata").(map[string]interface{})
		metaData := ExpandMetaData(metaDataRaw)

		if err := client.UpdateMetaData(ctx, account.ResourceGroup, id.AccountName, id.Name, metaData); err != nil {
			return diag.Errorf("updating the MetaData for Container %q (Storage Account %q / Resource Group %q): %s", id.Name, id.AccountName, account.ResourceGroup, err)
		}

		log.Printf("[DEBUG] Updated the MetaData for Container %q (Storage Account %q / Resource Group %q)", id.Name, id.AccountName, account.ResourceGroup)
	}

	return resourceStorageContainerRead(ctx, d, meta)
}

func ExpandMetaData(input map[string]interface{}) map[string]string {
	output := make(map[string]string)

	for k, v := range input {
		output[k] = v.(string)
	}

	return output
}

func FlattenMetaData(input map[string]string) map[string]interface{} {
	output := make(map[string]interface{})

	for k, v := range input {
		output[k] = v
	}

	return output
}
