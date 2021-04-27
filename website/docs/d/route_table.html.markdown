---
subcategory: ""
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_route_table"
sidebar_current: "docs-azurestack-data-source-route-table"
description: |-
  Gets information about a Route Table

---

# Data Source: azurestack_route_table

Gets information about a Route Table

## Example Usage

```hcl
data "azurestack_route_table" "test" {
  name                = "myroutetable"
  resource_group_name = "some-resource-group"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Route Table.

* `resource_group_name` - (Required) The name of the Resource Group in which the Route Table exists.

## Attributes Reference

The following attributes are exported:

* `id` - The Route Table ID.

* `location` - The Azure Region in which the Route Table exists.

* `route` - One or more `route` blocks as documented below.

* `subnets` - The collection of Subnets associated with this route table.

* `tags` - A mapping of tags assigned to the Route Table.

The `route` block exports the following:

* `name` - The name of the Route.

* `address_prefix` - The destination CIDR to which the route applies.

* `next_hop_type` - The type of Azure hop the packet should be sent to.

* `next_hop_in_ip_address` - Contains the IP address packets should be forwarded to.
