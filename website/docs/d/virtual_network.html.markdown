---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_virtual_network"
description: |-
  Get information about the specified Virtual Network.
---

# Data Source: azurestack_virtual_network

Use this data source to access the properties of an Azure Virtual Network.

## Example Usage

```hcl
data "azurestack_virtual_network" "test" {
  name                = "production"
  resource_group_name = "networking"
}

output "virtual_network_id" {
  value = data.azurestack_virtual_network.test.id
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the Virtual Network.
* `resource_group_name` - (Required) Specifies the name of the resource group the Virtual Network is located in.

## Attributes Reference

* `id` - The ID of the virtual network.
* `location` - Location of the virtual network.
* `address_space` - The list of address spaces used by the virtual network.
* `dns_servers` - The list of DNS servers used by the virtual network.
* `guid` - The GUID of the virtual network.
* `subnets` - The list of name of the subnets that are attached to this virtual network.
* `vnet_peerings` - A mapping of name - virtual network id of the virtual network peerings.

