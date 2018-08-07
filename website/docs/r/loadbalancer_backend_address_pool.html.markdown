---
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_lb_backend_address_pool"
sidebar_current: "docs-azurestack-resource-loadbalancer-backend-address-pool"
description: |-
  Manages a LoadBalancer Backend Address Pool.
---

# azurestack_lb_backend_address_pool

Manages a LoadBalancer Backend Address Pool.

~> **NOTE:** When using this resource, the LoadBalancer needs to have a FrontEnd IP Configuration Attached

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

resource "azurestack_lb_backend_address_pool" "test" {
  resource_group_name = "${azurestack_resource_group.test.name}"
  loadbalancer_id     = "${azurestack_lb.test.id}"
  name                = "BackEndAddressPool"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Backend Address Pool.
* `resource_group_name` - (Required) The name of the resource group in which to create the resource.
* `loadbalancer_id` - (Required) The ID of the LoadBalancer in which to create the Backend Address Pool.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the LoadBalancer to which the resource is attached.

## Import

Load Balancer Backend Address Pools can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_lb_backend_address_pool.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/loadBalancers/lb1/backendAddressPools/pool1
```

