---
subcategory: "Base"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_resource_group"
sidebar_current: "docs-azurestack-resource-resource-group"
description: |-
    Creates a new resource group on Azure.
---

# azurestack_resource_group

Creates a new resource group on Azure.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "testResourceGroup1"
  location = "West US"

  tags = {
    environment = "Production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the resource group. Must be unique on your
    Azure subscription.

* `location` - (Required) The location where the resource group should be created.
    For a list of all Azure locations, please consult [this link](http://azure.microsoft.com/en-us/regions/) or run `az account list-locations --output table`.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The resource group ID.


## Import

Resource Groups can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_resource_group.mygroup /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroup
```
