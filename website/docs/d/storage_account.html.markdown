---
subcategory: "Storage"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_storage_account"
description: |-
  Get information about the specified Storage Account.
---

# Data Source: azurestack_storage_account

Gets information about the specified Storage Account.

## Example Usage

```hcl
data "azurestack_storage_account" "test" {
  name                = "packerimages"
  resource_group_name = "packer-storage"
}

output "storage_account_tier" {
  value = data.azurestack_storage_account.test.account_tier
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the Storage Account
* `resource_group_name` - (Required) Specifies the name of the resource group the Storage Account is located in.

## Attributes Reference

* `id` - The ID of the Storage Account.

* `location` - The Azure location where the Storage Account exists

* `account_kind` - (Optional) Defines the Kind of account. Valid option is `Storage`.
   . Changing this forces a new resource to be created.
    Defaults to `Storage` currently as per [Azure Stack Storage Differences](https://docs.microsoft.com/en-us/azure/azure-stack/user/azure-stack-acs-differences)

* `account_tier` - Defines the Tier of this storage account.

* `account_replication_type` - Defines the type of replication used for this storage account.

* `account_encryption_source` - The Encryption Source for this Storage Account.

* `custom_domain` - A `custom_domain` block as documented below.

* `tags` - A mapping of tags to assigned to the resource.

* `enable_https_traffic_only` - Is traffic only allowed via HTTPS? See [here](https://docs.microsoft.com/en-us/azure/storage/storage-require-secure-transfer/)
  for more information.

* `primary_location` - The primary location of the Storage Account.

* `secondary_location` - The secondary location of the Storage Account.

* `primary_blob_endpoint` - The endpoint URL for blob storage in the primary location.

* `secondary_blob_endpoint` - The endpoint URL for blob storage in the secondary location.

* `primary_queue_endpoint` - The endpoint URL for queue storage in the primary location.

* `secondary_queue_endpoint` - The endpoint URL for queue storage in the secondary location.

* `primary_table_endpoint` - The endpoint URL for table storage in the primary location.

* `secondary_table_endpoint` - The endpoint URL for table storage in the secondary location.

* `primary_file_endpoint` - The endpoint URL for file storage in the primary location.

* `primary_access_key` - The primary access key for the Storage Account.

* `secondary_access_key` - The secondary access key for the Storage Account.

* `primary_connection_string` - The connection string associated with the primary location

* `secondary_connection_string` - The connection string associated with the secondary location

* `primary_blob_connection_string` - The connection string associated with the primary blob location

* `secondary_blob_connection_string` - The connection string associated with the secondary blob location

---

* `custom_domain` supports the following:

* `name` - The Custom Domain Name used for the Storage Account.
