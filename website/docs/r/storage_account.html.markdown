---
subcategory: "Storage"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_storage_account"
description: |-
  Manages a Azure Storage Account.
---

# azurestack_storage_account

Manages an Azure Storage Account.

## Example Usage

```hcl
resource "azurestack_resource_group" "testrg" {
  name     = "resourceGroupName"
  location = "westus"
}

resource "azurestack_storage_account" "testsa" {
  name                     = "storageaccountname"
  resource_group_name      = azurestack_resource_group.testrg.name
  location                 = "westus"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the storage account. Changing this forces a
    new resource to be created. This must be unique across the entire Azure service,
    not just within the resource group.

* `resource_group_name` - (Required) The name of the resource group in which to
    create the storage account. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the supported Azure location where the
    resource exists. Changing this forces a new resource to be created.

* `account_kind` - (Optional) Defines the Kind of account. Valid option is `Storage`.
   . Changing this forces a new resource to be created.
    Defaults to `Storage` currently as per [Azure Stack Storage Differences](https://docs.microsoft.com/en-us/azure/azure-stack/user/azure-stack-acs-differences)

* `account_tier` - (Required) Defines the Tier to use for this storage account. Valid options are `Standard` and `Premium`. Changing this forces a new resource to be created - **`Can be provisioned, but no performance limit or guarantee.`**

* `account_replication_type` - (Required) Defines the type of replication to use for this storage account. Valid option is `LRS` currently as per [Azure Stack Storage Differences](https://docs.microsoft.com/en-us/azure/azure-stack/user/azure-stack-acs-differences)

* `account_encryption_source` - (Optional) The Encryption Source for this Storage Account. Possible values are `Microsoft.Keyvault` and `Microsoft.Storage`. Defaults to `Microsoft.Storage`.

* `custom_domain` - (Optional) A `custom_domain` block as documented below.

* `tags` - (Optional) A mapping of tags to assign to the resource.\

* `enable_https_traffic_only` - (Optional) Boolean flag which forces HTTPS if enabled, see [here](https://docs.microsoft.com/en-us/azure/storage/storage-require-secure-transfer/)
  for more information. Defaults to `true`.

---

* `custom_domain` supports the following:

* `name` - (Optional) The Custom Domain Name to use for the Storage Account, which will be validated by Azure.
* `use_subdomain` - (Optional) Should the Custom Domain Name be validated by using indirect CNAME validation?

~> **Note:** [More information on Validation is available here](https://docs.microsoft.com/en-gb/azure/storage/blobs/storage-custom-domain-name)

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The storage account Resource ID.
* `primary_location` - The primary location of the storage account.
* `secondary_location` - The secondary location of the storage account.
* `primary_blob_endpoint` - The endpoint URL for blob storage in the primary location.
* `secondary_blob_endpoint` - The endpoint URL for blob storage in the secondary location.
* `primary_queue_endpoint` - The endpoint URL for queue storage in the primary location.
* `secondary_queue_endpoint` - The endpoint URL for queue storage in the secondary location.
* `primary_table_endpoint` - The endpoint URL for table storage in the primary location.
* `secondary_table_endpoint` - The endpoint URL for table storage in the secondary location.
* `primary_file_endpoint` - The endpoint URL for file storage in the primary location.
* `primary_access_key` - The primary access key for the storage account
* `secondary_access_key` - The secondary access key for the storage account
* `primary_connection_string` - The connection string associated with the primary location
* `secondary_connection_string` - The connection string associated with the secondary location
* `primary_blob_connection_string` - The connection string associated with the primary blob location
* `secondary_blob_connection_string` - The connection string associated with the secondary blob location

## Import

Storage Accounts can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_storage_account.storageAcc1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroup/providers/Microsoft.Storage/storageAccounts/myaccount
```
