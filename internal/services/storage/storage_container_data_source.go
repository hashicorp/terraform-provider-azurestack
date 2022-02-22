package storage

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
)

func dataSourceStorageContainer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		ReadContext: dataSourceStorageContainerRead,
		Timeouts: &pluginsdk.ResourceTimeout{
			Read: pluginsdk.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
			"storage_account_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},
			"container_access_type": {
				Type:     pluginsdk.TypeString,
				Computed: true,
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

func dataSourceStorageContainerRead(ctx context.Context, d *pluginsdk.ResourceData, meta interface{}) diag.Diagnostics {
	storageClient := meta.(*clients.Client).Storage
	_ = ctx
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	containerName := d.Get("name").(string)
	accountName := d.Get("storage_account_name").(string)

	account, err := storageClient.FindAccount(ctx, accountName)
	if err != nil {
		return diag.Errorf("retrieving Account %q for Container %q: %s", accountName, containerName, err)
	}
	if account == nil {
		return diag.Errorf("Unable to locate Account %q for Storage Container %q", accountName, containerName)
	}

	client, err := storageClient.ContainersClient(ctx, *account)
	if err != nil {
		return diag.Errorf("building Containers Client for Storage Account %q (Resource Group %q): %s", accountName, account.ResourceGroup, err)
	}

	id := parse.NewStorageContainerDataPlaneId(accountName, "azurestack", containerName).ID()
	d.SetId(id)

	props, err := client.Get(ctx, account.ResourceGroup, accountName, containerName)
	if err != nil {
		return diag.Errorf("retrieving Container %q (Account %q / Resource Group %q): %s", containerName, accountName, account.ResourceGroup, err)
	}
	if props == nil {
		return diag.Errorf("Container %q was not found in Account %q / Resource Group %q", containerName, accountName, account.ResourceGroup)
	}

	d.Set("name", containerName)
	d.Set("storage_account_name", accountName)
	d.Set("container_access_type", flattenStorageContainerAccessLevel(props.AccessLevel))

	if err := d.Set("metadata", FlattenMetaData(props.MetaData)); err != nil {
		return diag.Errorf("setting `metadata`: %+v", err)
	}

	d.Set("has_immutability_policy", props.HasImmutabilityPolicy)
	d.Set("has_legal_hold", props.HasLegalHold)

	return nil
}
