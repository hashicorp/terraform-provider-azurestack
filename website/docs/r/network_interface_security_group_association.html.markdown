---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_network_interface_security_group_association"
description: |-
  Manages the association between a Network Interface and a Network Security Group.

---

# azurestack_network_interface_security_group_association

Manages the association between a Network Interface and a Network Security Group.

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
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_security_group" "example" {
  name                = "example-nsg"
  location            = azurestack_resource_group.example.location
  resource_group_name = azurestack_resource_group.example.name
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

resource "azurestack_network_interface_security_group_association" "example" {
  network_interface_id      = azurestack_network_interface.example.id
  network_security_group_id = azurestack_network_security_group.example.id
}
```

## Argument Reference

The following arguments are supported:

* `network_interface_id` - (Required) The ID of the Network Interface. Changing this forces a new resource to be created.

* `network_security_group_id` - (Required) The ID of the Network Security Group which should be attached to the Network Interface. Changing this forces a new resource to be created.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `id` - The (Terraform specific) ID of the Association between the Network Interface and the Network Interface.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/language/resources/syntax#operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the association between the Network Interface and the Network Security Group.
* `update` - (Defaults to 30 minutes) Used when updating the association between the Network Interface and the Network Security Group.
* `read` - (Defaults to 5 minutes) Used when retrieving the association between the Network Interface and the Network Security Group.
* `delete` - (Defaults to 30 minutes) Used when deleting the association between the Network Interface and the Network Security Group.

## Import

Associations between Network Interfaces and Network Security Group can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_network_interface_security_group_association.association1 "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/networkInterfaces/example|/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/networkSecurityGroups/group1"
```

-> **NOTE:** This ID is specific to Terraform - and is of the format `{networkInterfaceId}|{networkSecurityGroupId}`.
