---
subcategory: "Storage"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_storage_blob"
description: |-
  Manages a Azure Storage Blob.
---

# azurestack_storage_blob

Manages an Azure Storage Blob.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acctestrg-d"
  location = "westus"
}

resource "azurestack_storage_account" "test" {
  name                     = "acctestaccs"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = "westus"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = azurestack_resource_group.test.name
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}

resource "azurestack_storage_blob" "testsb" {
  name = "sample.vhd"

  resource_group_name    = azurestack_resource_group.test.name
  storage_account_name   = azurestack_storage_account.test.name
  storage_container_name = azurestack_storage_container.test.name

  type = "page"
  size = 5120
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the storage blob. Must be unique within the storage container the blob is located.

* `storage_account_name` - (Required) Specifies the storage account in which to create the storage container.
 Changing this forces a new resource to be created.

* `storage_container_name` - (Required) The name of the storage container in which this blob should be created.

* `type` - (Optional) The type of the storage blob to be created. One of either `block` or `page`. When not copying from an existing blob,
    this becomes required.

* `size` - (Optional) Used only for `page` blobs to specify the size in bytes of the blob to be created. Must be a multiple of 512. Defaults to 0.

* `cache_control` - (Optional) Controls the [cache control header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control) content of the response when blob is requested .

* `content_type` - (Optional) The content type of the storage blob. Cannot be defined if `source_uri` is defined. Defaults to `application/octet-stream`.

* `source` - (Optional) An absolute path to a file on the local system. Cannot be defined if `source_uri` is defined.

* `source_content` - (Optional) The content for this blob which should be defined inline. This field can only be specified for Block blobs and cannot be specified if `source` or `source_uri` is specified.

* `source_uri` - (Optional) The URI of an existing blob, or a file in the Azure File service, to use as the source contents
    for the blob to be created. Changing this forces a new resource to be created. Cannot be defined if `source` is defined.

* `content_md5` - (Optional) The MD5 sum of the blob contents. Cannot be defined if `source_uri` is defined, or if blob type is Append or Page. Changing this forces a new resource to be created.

* `parallelism` - (Optional) The number of workers per CPU core to run for concurrent uploads. Defaults to `8`.

* `metadata` - (Optional) A map of custom blob metadata.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The storage blob Resource ID.
* `url` - The URL of the blob
