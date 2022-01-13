---
subcategory: "Base"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_resource_group"
description: |-
  Get information about the specified resource group.
---

# Data Source: azurestack_resource_group

Use this data source to access the properties of an Azure resource group.

## Example Usage

```hcl
data "azurestack_resource_group" "test" {
  name = "dsrg_test"
}

resource "azurestack_managed_disk" "test" {
  name                 = "managed_disk_name"
  location             = data.azurestack_resource_group.test.location
  resource_group_name  = data.azurestack_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the resource group.

~> **NOTE:** If the specified location doesn't match the actual resource group location, an error message with the actual location value will be shown.

## Attributes Reference

* `location` - The location of the resource group.
* `tags` - A mapping of tags assigned to the resource group.
