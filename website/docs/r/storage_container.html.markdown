---
subcategory: "Storage"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_storage_container"
description: |-
  Manages a Azure Storage Container.
---

# azurestack_storage_container

Manages an Azure Storage Container.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acctestrg"
  location = "westus"
}

resource "azurestack_storage_account" "test" {
  name                     = "accteststorageaccount"
  resource_group_name      = azurestack_resource_group.test.name
  location                 = "westus"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = azurestack_resource_group.test.name
  storage_account_name  = azurestack_storage_account.test.name
  container_access_type = "private"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the storage container. Must be unique within the storage service the container is located.

* `storage_account_name` - (Required) Specifies the storage account in which to create the storage container.
 Changing this forces a new resource to be created.

* `container_access_type` - (Optional) The 'interface' for access the container provides. Can be either `blob`, `container` or `private`. Defaults to `private`. Changing this forces a new resource to be created.

* `metadata` - (Optional) A mapping of MetaData for this Container. All metadata keys should be lowercase.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The storage container Resource ID.

* `has_immutability_policy` - Is there an Immutability Policy configured on this Storage Container?

* `has_legal_hold` - Is there a Legal Hold configured on this Storage Container?
