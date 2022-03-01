---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_virtual_network_peering"
description: |-
  Manages a virtual network peering which allows resources to access other
  resources in the linked virtual network.
---

# azurestack_virtual_network_peering

Manages a virtual network peering which allows resources to access other
resources in the linked virtual network.

## Example Usage

```hcl
resource "azurestack_resource_group" "example" {
  name     = "peeredvnets-rg"
  location = "ashregion"
}

resource "azurestack_virtual_network" "example-1" {
  name                = "peternetwork1"
  resource_group_name = azurestack_resource_group.example.name
  address_space       = ["10.0.1.0/24"]
  location            = "ashregion"
}

resource "azurestack_virtual_network" "example-2" {
  name                = "peternetwork2"
  resource_group_name = azurestack_resource_group.example.name
  address_space       = ["10.0.2.0/24"]
  location            = "ashregion"
}

resource "azurestack_virtual_network_peering" "example-1" {
  name                      = "peer1to2"
  resource_group_name       = azurestack_resource_group.example.name
  virtual_network_name      = azurestack_virtual_network.example-1.name
  remote_virtual_network_id = azurestack_virtual_network.example-2.id
}

resource "azurestack_virtual_network_peering" "example-2" {
  name                      = "peer2to1"
  resource_group_name       = azurestack_resource_group.example.name
  virtual_network_name      = azurestack_virtual_network.example-2.name
  remote_virtual_network_id = azurestack_virtual_network.example-1.id
}
```



## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual network peering. Changing this forces a new resource to be created.
* `virtual_network_name` - (Required) The name of the virtual network. Changing this forces a new resource to be created.
* `remote_virtual_network_id` - (Required) The full Azure resource ID of the remote virtual network.  Changing this forces a new resource to be created.
* `resource_group_name` - (Required) The name of the resource group in which to create the virtual network peering. Changing this forces a new resource to be created.
* `allow_virtual_network_access` - (Optional) Controls if the VMs in the remote virtual network can access VMs in the local virtual network. Defaults to true.
* `allow_forwarded_traffic` - (Optional) Controls if forwarded traffic from VMs in the remote virtual network is allowed. Defaults to false.
* `allow_gateway_transit` - (Optional) Controls gatewayLinks can be used in the remote virtual network’s link to the local virtual network.
* `use_remote_gateways` - (Optional) Controls if remote gateways can be used on the local virtual network. If the flag is set to `true`, and `allow_gateway_transit` on the remote peering is also `true`, virtual network will use gateways of remote virtual network for transit. Only one peering can have this flag set to `true`. This flag cannot be set if virtual network already has a gateway. Defaults to `false`.


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Virtual Network Peering.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Virtual Network Peering.
* `update` - (Defaults to 30 minutes) Used when updating the Virtual Network Peering.
* `read` - (Defaults to 5 minutes) Used when retrieving the Virtual Network Peering.
* `delete` - (Defaults to 30 minutes) Used when deleting the Virtual Network Peering.

## Note

Virtual Network peerings cannot be created, updated or deleted concurrently.

## Import

Virtual Network Peerings can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_virtual_network_peering.examplePeering /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/virtualNetworks/myvnet1/virtualNetworkPeerings/myvnet1peering
```