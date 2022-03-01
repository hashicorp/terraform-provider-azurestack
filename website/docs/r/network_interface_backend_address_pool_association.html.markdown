---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_network_interface_backend_address_pool_association"
description: |-
  Manages the association between a Network Interface and a Load Balancer's Backend Address Pool.

---

# azurestack_network_interface_backend_address_pool_association

Manages the association between a Network Interface and a Load Balancer's Backend Address Pool.

## Example Usage

```hcl
resource "azurestack_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurestack_virtual_network" "example" {
  name                = "example-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name
}

resource "azurestack_subnet" "example" {
  name                 = "internal"
  resource_group_name  = azurestack_resource_group.example.name
  virtual_network_name = azurestack_virtual_network.example.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurestack_public_ip" "example" {
  name                = "example-pip"
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name
  allocation_method   = "Static"
}

resource "azurestack_lb" "example" {
  name                = "example-lb"
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name

  frontend_ip_configuration {
    name                 = "primary"
    public_ip_address_id = azurestack_public_ip.example.id
  }
}

resource "azurestack_lb_backend_address_pool" "example" {
  resource_group_name = azurestack_resource_group.example.name
  loadbalancer_id     = azurestack_lb.example.id
  name                = "acctestpool"
}

resource "azurestack_network_interface" "example" {
  name                = "example-nic"
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurestack_subnet.example.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurestack_network_interface_backend_address_pool_association" "example" {
  network_interface_id    = azurestack_network_interface.example.id
  ip_configuration_name   = "testconfiguration1"
  backend_address_pool_id = azurestack_lb_backend_address_pool.example.id
}
```

## Argument Reference

The following arguments are supported:

* `network_interface_id` - (Required) The ID of the Network Interface. Changing this forces a new resource to be created.

* `ip_configuration_name` - (Required) The Name of the IP Configuration within the Network Interface which should be connected to the Backend Address Pool. Changing this forces a new resource to be created.

* `backend_address_pool_id` - (Required) The ID of the Load Balancer Backend Address Pool which this Network Interface should be connected to. Changing this forces a new resource to be created.

## Attributes Reference

The following attributes are exported:

* `id` - The (Terraform specific) ID of the Association between the Network Interface and the Load Balancers Backend Address Pool.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the association between the Network Interface and the Load Balancers Backend Address Pool.
* `update` - (Defaults to 30 minutes) Used when updating the association between the Network Interface and the Load Balancers Backend Address Pool.
* `read` - (Defaults to 5 minutes) Used when retrieving the association between the Network Interface and the Load Balancers Backend Address Pool.
* `delete` - (Defaults to 30 minutes) Used when deleting the association between the Network Interface and the Load Balancers Backend Address Pool.

## Import

Associations between Network Interfaces and Load Balancer Backend Address Pools can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_network_interface_backend_address_pool_association.association1 "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/microsoft.network/networkInterfaces/nic1/ipConfigurations/example|/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/loadBalancers/lb1/backendAddressPools/pool1"
```

-> **NOTE:** This ID is specific to Terraform - and is of the format `{networkInterfaceId}/ipConfigurations/{ipConfigurationName}|{backendAddressPoolId}`.
