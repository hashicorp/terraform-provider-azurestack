---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_subnet_network_security_group_association"
description: |-
  Associates a [Network Security Group](network_security_group.html) with a [Subnet](subnet.html) within a [Virtual Network](virtual_network.html).

---

# azurestack_subnet_network_security_group_association

Associates a [Network Security Group](network_security_group.html) with a [Subnet](subnet.html) within a [Virtual Network](virtual_network.html).

## Example Usage

```hcl
resource "azurestack_resource_group" "example" {
  name     = "example-resources"
  location = "ash"
}

resource "azurestack_virtual_network" "example" {
  name                = "example-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name
}

resource "azurestack_subnet" "example" {
  name                 = "frontend"
  resource_group_name  = azurestack_resource_group.example.name
  virtual_network_name = azurestack_virtual_network.example.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_security_group" "example" {
  name                = "example-nsg"
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurestack_subnet_network_security_group_association" "example" {
  subnet_id                 = azurestack_subnet.example.id
  network_security_group_id = azurestack_network_security_group.example.id
}
```

## Argument Reference

The following arguments are supported:

* `network_security_group_id` - (Required) The ID of the Network Security Group which should be associated with the Subnet. Changing this forces a new resource to be created.

* `subnet_id` - (Required) The ID of the Subnet. Changing this forces a new resource to be created.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The ID of the Subnet.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/language/resources/syntax#operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Subnet Network Security Group Association.
* `update` - (Defaults to 30 minutes) Used when updating the Subnet Network Security Group Association.
* `read` - (Defaults to 5 minutes) Used when retrieving the Subnet Network Security Group Association.
* `delete` - (Defaults to 30 minutes) Used when deleting the Subnet Network Security Group Association.

## Import

Subnet `<->` Network Security Group Associations can be imported using the `resource id` of the Subnet, e.g.

```shell
terraform import azurestack_subnet_network_security_group_association.association1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/virtualNetworks/myvnet1/subnets/mysubnet1
```
