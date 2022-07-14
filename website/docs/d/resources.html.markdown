---
subcategory: "Base"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_resources"
description: |-
  Gets information about an existing Resources.
---

# Data Source: azurestack_resources

Use this data source to access information about existing resources.

## Example Usage

```hcl
# Get Resources from a Resource Group
data "azurestack_resources" "example" {
  resource_group_name = "example-resources"
}

# Get Resources with specific Tags
data "azurestack_resources" "example" {
  resource_group_name = "example-resources"

  required_tags = {
    environment = "production"
    role        = "webserver"
  }
}

# Get resources by type, create spoke vNet peerings
data "azurestack_resources" "spokes" {
  type = "Microsoft.Network/virtualNetworks"

  required_tags = {
    environment = "production"
    role        = "spokeNetwork"
  }
}

resource "azurestack_virtual_network_peering" "spoke_peers" {
  count = length(data.azurestack_resources.spokes.resources)

  name                      = "hub2${data.azurestack_resources.spokes.resources[count.index].name}"
  resource_group_name       = azurestack_resource_group.hub.name
  virtual_network_name      = azurestack_virtual_network.hub.name
  remote_virtual_network_id = data.azurestack_resources.spokes.resources[count.index].id
}
```

## Argument Reference

~> **Note:** At least one of `name`, `resource_group_name` or `type` must be specified.

* `name` - (Optional) The name of the Resource.

* `resource_group_name` - (Optional) The name of the Resource group where the Resources are located.

* `type` - (Optional) The Resource Type of the Resources you want to list (e.g. `Microsoft.Network/virtualNetworks`). A full list of available Resource Types can be found [here](https://docs.microsoft.com/en-us/azure/azure-resource-manager/azure-services-resource-providers).

* `required_tags` - (Optional) A mapping of tags which the resource has to have in order to be included in the result.

## Attributes Reference

* `resources` - One or more `resource` blocks as defined below.

---

The `resource` block exports the following:

* `name` - The name of this Resource.

* `id` - The ID of this Resource.

* `type` - The type of this Resource. (e.g. `Microsoft.Network/virtualNetworks`).

* `location` - The Azure Region in which this Resource exists.

* `tags` - A map of tags assigned to this Resource.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Resources.
