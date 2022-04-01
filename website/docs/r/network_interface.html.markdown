---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_network_interface"
description: |-
  Manages a Network Interface located in a Virtual Network, usually attached to a Virtual Machine.

---

# azurestack_network_interface

Manages a Network Interface located in a Virtual Network, usually attached to a Virtual Machine.

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
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_network_interface" "test" {
  name                = "acceptanceTestNetworkInterface1"
  location            = azurestack_resource_group.test.location
  resource_group_name = azurestack_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurestack_subnet.test.id
    private_ip_address_allocation = "dynamic"
  }

  tags = {
    environment = "staging"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the network interface. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create the network interface. Changing this forces a new resource to be created.

* `location` - (Required) The location/region where the network interface is created. Changing this forces a new resource to be created.

* `enable_ip_forwarding` - (Optional) Enables IP Forwarding on the NIC. Defaults to `false`.

* `dns_servers` - (Optional) List of DNS servers IP addresses to use for this NIC, overrides the VNet-level server list

* `ip_configuration` - (Required) One or more `ip_configuration` associated with this NIC as documented below.

* `tags` - (Optional) A mapping of tags to assign to the resource.

The `ip_configuration` block supports:

* `name` - (Required) User-defined name of the IP.

* `subnet_id` - (Required) Reference to a subnet in which this NIC has been created.

* `private_ip_address` - (Optional) Static IP Address.

* `private_ip_address_allocation` - (Required) Defines how a private IP address is assigned. Options are Static or Dynamic.

* `public_ip_address_id` - (Optional) Reference to a Public IP Address to associate with this NIC

* `private_ip_address_version` - (Optional) The IP Version to use. Possible values are `IPv4`.

* `primary` - (Optional) Is this the Primary Network Interface? If set to `true` this should be the first `ip_configuration` in the array.

## Attributes Reference

The following attributes are exported:

* `id` - The Virtual Network Interface ID.
* `mac_address` - The media access control (MAC) address of the network interface.
* `private_ip_address` - The private ip address of the network interface.
* `virtual_machine_id` - Reference to a VM with which this NIC has been associated.
* `applied_dns_servers` - If the VM that uses this NIC is part of an Availability Set, then this list will have the union of all DNS servers from all NICs that are part of the Availability Set

## Import

Network Interfaces can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_network_interface.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/microsoft.network/networkInterfaces/nic1
```
