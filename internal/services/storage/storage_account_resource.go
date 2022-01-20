package storage

// TODO - bring in line with the azurerm version of this resource

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/pluginsdk"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/storage/mgmt/storage"
	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurestack/internal/az/tags"
	"github.com/hashicorp/terraform-provider-azurestack/internal/clients"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/parse"
	"github.com/hashicorp/terraform-provider-azurestack/internal/services/storage/validate"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurestack/internal/tf/timeouts"
	"github.com/hashicorp/terraform-provider-azurestack/internal/utils"
)

const blobStorageAccountDefaultAccessTier = "Hot"

func storageAccount() *schema.Resource {
	return &schema.Resource{
		Create: storageAccountCreate,
		Read:   storageAccountRead,
		Update: storageAccountUpdate,
		Delete: storageAccountDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.StorageAccountID(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageAccountName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"location": commonschema.Location(),

			"account_kind": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(storage.Storage),
					string(storage.BlobStorage),
				}, true),
				Default: string(storage.Storage),
			},

			// Constants not in the 2017-03-09 profile
			"account_tier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Standard",
					"Premium",
				}, true), // TODO should we try removing all case ignores for 1.0?
				DiffSuppressFunc: suppress.CaseDifference,
			},

			// Constants not in 2017-03-09 profile
			"account_replication_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"LRS",
					"ZRS",
					"GRS",
					"RAGRS",
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			// Constants not in 2017-03-09 profile
			"account_encryption_source": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string("Microsoft.Storage"),
				ValidateFunc: validation.StringInSlice([]string{
					string(storage.MicrosoftKeyvault),
					string(storage.MicrosoftStorage),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"custom_domain": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"use_subdomain": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"enable_blob_encryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"primary_location": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_location": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_blob_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_blob_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_queue_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_queue_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_table_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_table_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// NOTE: The API does not appear to expose a secondary file endpoint
			"primary_file_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_connection_string": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_connection_string": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_blob_connection_string": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_blob_connection_string": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func storageAccountCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	storageAccountName := d.Get("name").(string)
	accountKind := d.Get("account_kind").(string)

	location := d.Get("location").(string)

	accountTier := d.Get("account_tier").(string)
	replicationType := d.Get("account_replication_type").(string)
	storageType := fmt.Sprintf("%s_%s", accountTier, replicationType)

	// Not supported by the profile in the same struct as the original, both of the
	// following commented lines will be read and set later on the correct
	// structs
	// storageAccountEncryptionSource := d.Get("account_encryption_source").(string)
	// enableBlobEncryption := d.Get("enable_blob_encryption").(bool)

	parameters := storage.AccountCreateParameters{
		Location: &location,
		Sku: &storage.Sku{
			Name: storage.SkuName(storageType),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
		Kind: storage.Kind(accountKind),

		// If any paramers are specified withouth the right values this will fail
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
	}

	if _, ok := d.GetOk("custom_domain"); ok {
		parameters.CustomDomain = expandStorageAccountCustomDomain(d)
	}

	// BlobStorage does not support ZRS
	if accountKind == string(storage.BlobStorage) {
		if string(parameters.Sku.Name) == string(storage.StandardZRS) {
			return fmt.Errorf("A `account_replication_type` of `ZRS` isn't supported for Blob Storage accounts.")
		}
		accessTier, ok := d.GetOk("access_tier")
		if !ok {
			// default to "Hot"
			accessTier = blobStorageAccountDefaultAccessTier
		}

		parameters.AccountPropertiesCreateParameters.AccessTier = storage.AccessTier(accessTier.(string))

		enableBlobEncryption := d.Get("enable_blob_encryption").(bool)

		if enableBlobEncryption {
			// if the encryption is enabled, then set the arguments
			storageAccountEncryptionSource := d.Get("account_encryption_source").(string)
			parameters.AccountPropertiesCreateParameters.Encryption = &storage.Encryption{
				Services: &storage.EncryptionServices{
					Blob: &storage.EncryptionService{
						Enabled: pointer.FromBool(enableBlobEncryption),
					},
				},
				KeySource: storage.KeySource(storageAccountEncryptionSource),
			}
		}
	}

	future, err := client.Create(ctx, resourceGroupName, storageAccountName, parameters)
	if err != nil {
		return fmt.Errorf(
			"Error creating Azure Storage Account %q: %+v",
			storageAccountName, err)
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return fmt.Errorf(
			"Error while waiting for Azure Storage Account %q: %+v",
			storageAccountName, err)
	}

	account, err := future.Result(*client)
	if err != nil {
		return fmt.Errorf(
			"Error while fetching Azure Storage Account %q: %+v",
			storageAccountName, err)
	}

	log.Printf("[INFO] storage account %q ID: %q", storageAccountName, *account.ID)
	d.SetId(*account.ID)

	return storageAccountRead(d, meta)
}

// resourceArmStorageAccountUpdate is unusual in the ARM API where most resources have a combined
// and idempotent operation for CreateOrUpdate. In particular updating all of the parameters
// available requires a call to Update per parameter...
func storageAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageAccountID(d.Id())
	if err != nil {
		return err
	}

	accountTier := d.Get("account_tier").(string)
	replicationType := d.Get("account_replication_type").(string)
	storageType := fmt.Sprintf("%s_%s", accountTier, replicationType)
	accountKind := d.Get("account_kind").(string)

	if accountKind == string(storage.BlobStorage) {
		if storageType == string(storage.StandardZRS) {
			return fmt.Errorf("A `account_replication_type` of `ZRS` isn't supported for Blob Storage accounts.")
		}
	}

	d.Partial(true)

	if d.HasChange("account_replication_type") {
		sku := storage.Sku{
			Name: storage.SkuName(storageType),
		}

		opts := storage.AccountUpdateParameters{
			Sku: &sku,
		}

		_, err := client.Update(ctx, id.ResourceGroup, id.Name, opts)
		if err != nil {
			return fmt.Errorf("updating Azure Storage Account type %q: %+v", id.Name, err)
		}
	}

	// if d.HasChange("access_tier") {
	// 	accessTier := d.Get("access_tier").(string)

	// 	opts := storage.AccountUpdateParameters{
	// 		AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
	// 			AccessTier: storage.AccessTier(accessTier),
	// 		},
	// 	}

	// 	_, err := client.Update(ctx, resourceGroupName, storageAccountName, opts)
	// 	if err != nil {
	// 		return fmt.Errorf("updating Azure Storage Account access_tier %q: %+v", storageAccountName, err)
	// 	}

	// 	d.SetPartial("access_tier")
	// }

	if d.HasChange("tags") {
		opts := storage.AccountUpdateParameters{
			Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
		}

		_, err := client.Update(ctx, id.ResourceGroup, id.Name, opts)
		if err != nil {
			return fmt.Errorf("updating Azure Storage Account tags %q: %+v", id.Name, err)
		}
	}

	if d.HasChange("enable_blob_encryption") {
		encryptionSource := d.Get("account_encryption_source").(string)

		opts := storage.AccountUpdateParameters{
			AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
				Encryption: &storage.Encryption{
					Services:  &storage.EncryptionServices{},
					KeySource: storage.KeySource(encryptionSource),
				},
			},
		}

		if d.HasChange("enable_blob_encryption") {
			enableEncryption := d.Get("enable_blob_encryption").(bool)
			opts.Encryption.Services.Blob = &storage.EncryptionService{
				Enabled: pointer.FromBool(enableEncryption),
			}
		}

		_, err := client.Update(ctx, id.ResourceGroup, id.Name, opts)
		if err != nil {
			return fmt.Errorf("updating Azure Storage Account Encryption %q: %+v", id.Name, err)
		}
	}

	if d.HasChange("custom_domain") {
		customDomain := expandStorageAccountCustomDomain(d)
		opts := storage.AccountUpdateParameters{
			AccountPropertiesUpdateParameters: &storage.AccountPropertiesUpdateParameters{
				CustomDomain: customDomain,
			},
		}

		_, err := client.Update(ctx, id.ResourceGroup, id.Name, opts)
		if err != nil {
			return fmt.Errorf("updating Azure Storage Account Custom Domain %q: %+v", id.Name, err)
		}
	}

	d.Partial(false)
	return nil
}

func storageAccountRead(d *schema.ResourceData, meta interface{}) error {
	endpointSuffix := meta.(*clients.Client).Account.Environment.StorageEndpointSuffix
	client := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageAccountID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetProperties(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading the state of AzurStack Storage Account %q: %+v", id.Name, err)
	}
	// (resGroup, name)
	keys, err := client.ListKeys(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return err
	}

	accessKeys := *keys.Keys
	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("account_kind", resp.Kind)

	if sku := resp.Sku; sku != nil {
		d.Set("account_type", sku.Name)
		d.Set("account_tier", sku.Tier)
		d.Set("account_replication_type", strings.Split(fmt.Sprintf("%v", sku.Name), "_")[1])
	}

	if props := resp.AccountProperties; props != nil {
		// Currently not supported on Azure Stack
		// d.Set("access_tier", props.AccessTier)

		if customDomain := props.CustomDomain; customDomain != nil {
			if err := d.Set("custom_domain", flattenStorageAccountCustomDomain(customDomain)); err != nil {
				return fmt.Errorf("flattening `custom_domain`: %+v", err)
			}
		}

		if encryption := props.Encryption; encryption != nil {
			if services := encryption.Services; services != nil {
				if blob := services.Blob; blob != nil {
					d.Set("enable_blob_encryption", blob.Enabled)
				}
			}
			d.Set("account_encryption_source", string(encryption.KeySource))
		}

		// Computed
		d.Set("primary_location", props.PrimaryLocation)
		d.Set("secondary_location", props.SecondaryLocation)

		if len(accessKeys) > 0 {
			pcs := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=%s", *resp.Name, *accessKeys[0].Value, endpointSuffix)
			d.Set("primary_connection_string", pcs)
		}

		if len(accessKeys) > 1 {
			scs := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=%s", *resp.Name, *accessKeys[1].Value, endpointSuffix)
			d.Set("secondary_connection_string", scs)
		}

		if endpoints := props.PrimaryEndpoints; endpoints != nil {
			d.Set("primary_blob_endpoint", endpoints.Blob)
			d.Set("primary_queue_endpoint", endpoints.Queue)
			d.Set("primary_table_endpoint", endpoints.Table)
			d.Set("primary_file_endpoint", endpoints.File)

			pscs := fmt.Sprintf("DefaultEndpointsProtocol=https;BlobEndpoint=%s;AccountName=%s;AccountKey=%s",
				*endpoints.Blob, *resp.Name, *accessKeys[0].Value)
			d.Set("primary_blob_connection_string", pscs)
		}

		if endpoints := props.SecondaryEndpoints; endpoints != nil {
			if blob := endpoints.Blob; blob != nil {
				d.Set("secondary_blob_endpoint", blob)
				sscs := fmt.Sprintf("DefaultEndpointsProtocol=https;BlobEndpoint=%s;AccountName=%s;AccountKey=%s",
					*blob, *resp.Name, *accessKeys[1].Value)
				d.Set("secondary_blob_connection_string", sscs)
			} else {
				d.Set("secondary_blob_endpoint", "")
				d.Set("secondary_blob_connection_string", "")
			}

			if endpoints.Queue != nil {
				d.Set("secondary_queue_endpoint", endpoints.Queue)
			} else {
				d.Set("secondary_queue_endpoint", "")
			}

			if endpoints.Table != nil {
				d.Set("secondary_table_endpoint", endpoints.Table)
			} else {
				d.Set("secondary_table_endpoint", "")
			}
		}
	}

	d.Set("primary_access_key", accessKeys[0].Value)
	d.Set("secondary_access_key", accessKeys[1].Value)

	return tags.FlattenAndSet(d, resp.Tags)
}

func storageAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.StorageAccountID(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("issuing AzureStack delete request for storage account %q: %+v", id.Name, err)
	}

	return nil
}

func expandStorageAccountCustomDomain(d *schema.ResourceData) *storage.CustomDomain {
	domains := d.Get("custom_domain").([]interface{})
	if len(domains) == 0 {
		return &storage.CustomDomain{
			Name: pointer.FromString(""),
		}
	}

	domain := domains[0].(map[string]interface{})
	name := domain["name"].(string)
	useSubDomain := domain["use_subdomain"].(bool)
	return &storage.CustomDomain{
		Name:             pointer.FromString(name),
		UseSubDomainName: pointer.FromBool(useSubDomain),
	}
}

func flattenStorageAccountCustomDomain(input *storage.CustomDomain) []interface{} {
	domain := make(map[string]interface{})

	domain["name"] = *input.Name
	// use_subdomain isn't returned

	return []interface{}{domain}
}
