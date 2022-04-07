---
subcategory: "Compute"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_managed_disk"
description: |-
  Manages a Managed Disk.
---

# azurestack_managed_disk

Manage a managed disk.

## Example Usage with Create Empty

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acctestRG"
  location = "West US 2"
}

resource "azurestack_managed_disk" "test" {
  name                 = "acctestmd"
  location             = "West US 2"
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "staging"
  }
}
```

## Example Usage with Create Copy

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acctestRG"
  location = "West US 2"
}

resource "azurestack_managed_disk" "source" {
  name                 = "acctestmd1"
  location             = "West US 2"
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_managed_disk" "copy" {
  name                 = "acctestmd2"
  location             = "West US 2"
  resource_group_name  = azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Copy"
  source_resource_id   = azurestack_managed_disk.source.id
  disk_size_gb         = "1"

  tags = {
    environment = "staging"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the managed disk. Changing this forces a
    new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create
    the managed disk.

* `location` - (Required) Specified the supported Azure location where the resource exists.
    Changing this forces a new resource to be created.

* `storage_account_type` - (Required) The type of storage to use for the managed disk.
    Allowable values are `Standard_LRS` or `Premium_LRS`.

* `create_option` - (Required) The method to use when creating the managed disk. Possible values include:
 * `Import` - Import a VHD file in to the managed disk (VHD specified with `source_uri`).
 * `Empty` - Create an empty managed disk.
 * `Copy` - Copy an existing managed disk or snapshot (specified with `source_resource_id`).
 * `FromImage` - Copy a Platform Image (specified with `image_reference_id`)

* `source_uri` - (Optional) URI to a valid VHD file to be used when `create_option` is `Import`.

* `source_resource_id` - (Optional) ID of an existing managed disk to copy `create_option` is `Copy`.

* `image_reference_id` - (Optional) ID of an existing platform/marketplace disk image to copy when `create_option` is `FromImage`.

* `os_type` - (Optional) Specify a value when the source of an `Import` or `Copy`
    operation targets a source that contains an operating system. Valid values are `Linux` or `Windows`

* `disk_size_gb` - (Optional, Required for a new managed disk) Specifies the size of the managed disk to create in gigabytes.
    If `create_option` is `Copy` or `FromImage`, then the value must be equal to or greater than the source's size.

* `encryption` - (Optional) A `encryption` block as defined below.

* `hyper_v_generation` - (Optional) The HyperV Generation of the Disk when the source of an `Import` or `Copy` operation targets a source that contains an operating system. Possible values are `V1` and `V2`. Changing this forces a new resource to be created.

* `tags` - (Optional) A mapping of tags to assign to the resource.

---

The `encryption` block supports:

* `enabled` - (Required) Is Encryption enabled on this Managed Disk? Changing this forces a new resource to be created.

* `disk_encryption_key` - (Optional) A `disk_encryption_key` block as defined above.

* `key_encryption_key` - (Optional) A `key_encryption_key` block as defined below.

---

For more information on managed disks, such as sizing options and pricing, please check out the
[azure documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview).

## Attributes Reference

The following attributes are exported:

* `id` - The managed disk ID.

## Import

Managed Disks can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_managed_disk.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/microsoft.compute/disks/manageddisk1
```
