---
subcategory: "Compute"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_managed_disk"
description: |-
  Get information about an existing Managed Disk.
---

# Data Source: azurestack_managed_disk

Use this data source to access information about an existing Managed Disk.

## Example Usage

```hcl
data "azurestack_managed_disk" "existing" {
  name                = "example-datadisk"
  resource_group_name = "example-resources"
}

output "id" {
  value = data.azurestack_managed_disk.existing.id
}
```

## Argument Reference

* `name` - Specifies the name of the Managed Disk.

* `resource_group_name` - Specifies the name of the Resource Group where this Managed Disk exists.

## Attributes Reference

* `disk_size_gb` - The size of the Managed Disk in gigabytes.

* `image_reference_id` - The ID of the source image used for creating this Managed Disk.

* `os_type` - The operating system used for this Managed Disk.

* `storage_account_type` - The storage account type for the Managed Disk.

* `source_uri` - The Source URI for this Managed Disk.

* `source_resource_id` - The ID of an existing Managed Disk which this Disk was created from.

* `storage_account_id` - The ID of the Storage Account where the `source_uri` is located.

* `tags` - A mapping of tags assigned to the resource.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Managed Disk.
