---
subcategory: "Load Balancer"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_lb_nat_pool"
sidebar_current: "docs-azurestack-resource-loadbalancer-nat-pool"
description: |-
  Manages a Load Balancer NAT Pool.
---

# azurestack_lb_nat_pool

Manages a Load Balancer NAT pool.

~> **NOTE** When using this resource, the Load Balancer needs to have a FrontEnd IP Configuration Attached

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

resource "azurestack_lb_nat_pool" "test" {
  resource_group_name            = "${azurestack_resource_group.test.name}"
  loadbalancer_id                = "${azurestack_lb.test.id}"
  name                           = "SampleApplicationPool"
  protocol                       = "Tcp"
  frontend_port_start            = 80
  frontend_port_end              = 81
  backend_port                   = 8080
  frontend_ip_configuration_name = "PublicIPAddress"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the NAT pool.
* `resource_group_name` - (Required) The name of the resource group in which to create the resource.
* `loadbalancer_id` - (Required) The ID of the Load Balancer in which to create the NAT pool.
* `frontend_ip_configuration_name` - (Required) The name of the frontend IP configuration exposing this rule.
* `protocol` - (Required) The transport protocol for the external endpoint. Possible values are `Udp` or `Tcp`.
* `frontend_port_start` - (Required) The first port number in the range of external ports that will be used to provide Inbound Nat to NICs associated with this Load Balancer. Possible values range between 1 and 65534, inclusive.
* `frontend_port_end` - (Required) The last port number in the range of external ports that will be used to provide Inbound Nat to NICs associated with this Load Balancer. Possible values range between 1 and 65534, inclusive.
* `backend_port` - (Required) The port used for the internal endpoint. Possible values range between 1 and 65535, inclusive.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Load Balancer to which the resource is attached.

## Import

Load Balancer NAT Pools can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_lb_nat_pool.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/loadBalancers/lb1/inboundNatPools/pool1
```
