---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_subnet"
description: |-
  Manages a subnet. Subnets represent network segments within the IP space defined by the virtual network.

---

# azurestack_subnet

Manages a subnet. Subnets represent network segments within the IP space defined by the virtual network.

~> **NOTE on Virtual Networks and Subnet's:** Terraform currently
provides both a standalone [Subnet resource](subnet.html), and allows for Subnets to be defined in-line within the [Virtual Network resource](virtual_network.html).
At this time you cannot use a Virtual Network with in-line Subnets in conjunction with any Subnet resources. Doing so will cause a conflict of Subnet configurations and will overwrite Subnet's.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acceptanceTestResourceGroup1"
  location = "West US"
}

resource "azurestack_virtual_network" "test" {
  name                = "acceptanceTestVirtualNetwork1"
  address_space       = ["10.0.0.0/16"]
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name
}

resource "azurestack_subnet" "test" {
  name                 = "testsubnet"
  resource_group_name  = azurestack_resource_group.test.name
  virtual_network_name = azurestack_virtual_network.test.name
  address_prefix       = "10.0.1.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the subnet. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create the subnet. Changing this forces a new resource to be created.

* `virtual_network_name` - (Required) The name of the virtual network to which to attach the subnet. Changing this forces a new resource to be created.

* `address_prefix` - (Required) The address prefix to use for the subnet.

## Attributes Reference

The following attributes are exported:

* `id` - The subnet ID.
* `name` - The name of the subnet.
* `resource_group_name` - The name of the resource group in which the subnet is created in.
* `virtual_network_name` - The name of the virtual network in which the subnet is created in
* `address_prefix` - The address prefix for the subnet

## Import

Subnets can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_subnet.testSubnet /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/virtualNetworks/myvnet1/subnets/mysubnet1
```
