---
subcategory: "Load Balancer"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_lb"
sidebar_current: "docs-azurestack-resource-loadbalancer-x"
description: |-
  Manages a Load Balancer Resource.
---

# azurestack_lb

Manages a Load Balancer Resource.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "LoadBalancerRG"
  location = "West US"
}

resource "azurestack_public_ip" "test" {
  name                         = "PublicIPForLB"
  location                     = "West US"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurestack_lb" "test" {
  name                = "TestLoadBalancer"
  location            = "West US"
  resource_group_name = "${azurestack_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = "${azurestack_public_ip.test.id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the LoadBalancer.
* `resource_group_name` - (Required) The name of the resource group in which to create the LoadBalancer.
* `location` - (Required) Specifies the supported Azure location where the resource exists.
* `frontend_ip_configuration` - (Optional) A frontend ip configuration block as documented below.

* `tags` - (Optional) A mapping of tags to assign to the resource.

`frontend_ip_configuration` supports the following:

* `name` - (Required) Specifies the name of the frontend ip configuration.
* `subnet_id` - (Optional) Reference to subnet associated with the IP Configuration.
* `private_ip_address` - (Optional) Private IP Address to assign to the Load Balancer. The last one and first four IPs in any range are reserved and cannot be manually assigned.
* `private_ip_address_allocation` - (Optional) Defines how a private IP address is assigned. Options are Static or Dynamic.
* `public_ip_address_id` - (Optional) Reference to Public IP address to be associated with the Load Balancer.


## Attributes Reference

The following attributes are exported:

* `id` - The LoadBalancer ID.
* `private_ip_address` - The first private IP address assigned to the load balancer in `frontend_ip_configuration` blocks, if any.
* `private_ip_addresses` - The list of private IP address assigned to the load balancer in `frontend_ip_configuration` blocks, if any.

## Import

Load Balancers can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_lb.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/loadBalancers/lb1
```
