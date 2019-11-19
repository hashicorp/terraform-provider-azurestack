---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_route_table"
sidebar_current: "docs-azurestack-resource-network-route-table"
description: |-
  Manages a Route Table

---

# azurestack_route_table

Manages a Route Table

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acceptanceTestResourceGroup1"
  location = "West US"
}

resource "azurestack_route_table" "test" {
  name                = "acceptanceTestSecurityGroup1"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  disable_bgp_route_propagation = false

  route {
    name           = "route1"
    address_prefix = "10.1.0.0/16"
    next_hop_type  = "vnetlocal"
  }

  tags = {
    environment = "Production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the route table. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create the route table. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `route` - (Optional) Can be specified multiple times to define multiple routes. Each `route` block supports fields documented below.

* `tags` - (Optional) A mapping of tags to assign to the resource.

The `route` block supports:

* `name` - (Required) The name of the route.

* `address_prefix` - (Required) The destination CIDR to which the route applies, such as 10.1.0.0/16

* `next_hop_type` - (Required) The type of Azure hop the packet should be sent to. Possible values are `VirtualNetworkGateway`, `VnetLocal`, `Internet`, `VirtualAppliance` and `None`.

* `next_hop_in_ip_address` - (Optional) Contains the IP address packets should be forwarded to. Next hop values are only allowed in routes where the next hop type is `VirtualAppliance`.


## Attributes Reference

The following attributes are exported:

* `id` - The Route Table ID.
* `subnets` - The collection of Subnets associated with this route table.

## Import

Route Tables can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_route_table.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/routeTables/mytable1
```
