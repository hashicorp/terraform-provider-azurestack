---
layout: "azurestack"
page_title: "Azure Stack: azurestack_virtual_network_gateway"
sidebar_current: "docs-azurestack-resource-network-virtual-network-gateway"
description: |-
  Manages a Virtual Network Gateway to establish secure, cross-premises connectivity.
---

# azurestack_virtual_network_gateway

Manages a Virtual Network Gateway to establish secure, cross-premises connectivity.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "test"
  location = "Azure-stack-region"
}

resource "azurestack_virtual_network" "test" {
  name                = "test"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  address_space       = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "test" {
  name                 = "GatewaySubnet"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.1.0/24"
}

resource "azurestack_public_ip" "test" {
  name                         = "test"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "test" {
  name                = "test"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  type     = "Vpn"
  vpn_type = "RouteBased"
  sku      = "Basic"

  ip_configuration {
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test.id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection. Changing the name forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create the connection Changing the name forces a new resource to be created.

* `location` - (Required) The location/region where the connection is located. Changing this forces a new resource to be created.

* `type` - (Required) The type of the Virtual Network Gateway. Valid options is `Vpn`

* `vpn_type` - (Optional) The routing type of the Virtual Network Gateway. Only valid option is `RouteBased`.

* `enable_bgp` - (Optional) If `true`, BGP (Border Gateway Protocol) is enabled for this connection. Defaults to `false`.

* `sku` - (Required) Configuration of the size and capacity of the virtual network gateway. Valid options are `Basic`, `Standard` and `HighPerformance`.

* `ip_configuration` - (Required) One or two ip_configuration blocks documented below. An active-standby gateway requires exactly one ip_configuration block whereas an active-active gateway requires exactly two ip_configuration blocks.

* `tags` - (Optional) A mapping of tags to assign to the resource.

The `ip_configuration` block supports:

* `name` - (Optional) A user-defined name of the IP configuration. Defaults to vnetGatewayConfig.

* `private_ip_address_allocation` - (Optional) Defines how the private IP address of the gateways virtual interface is assigned. Valid options are Static or Dynamic. Defaults to Dynamic.

* `subnet_id` - (Required) The ID of the gateway subnet of a virtual network in which the virtual network gateway will be created. It is mandatory that the associated subnet is named `GatewaySubnet`. Therefore, each virtual network can contain at most a single Virtual Network Gateway.

* `public_ip_address_id` - (Optional) The ID of the public ip address to associate with the Virtual Network Gateway.

The `bgp_settings` block supports:

* `asn` - (Optional) The Autonomous System Number (ASN) to use as part of the BGP.

* `peering_address` - (Optional) The BGP peer IP address of the virtual network gateway. This address is needed to configure the created gateway as a BGP Peer on the on-premises VPN devices. The IP address must be part of the subnet of the Virtual Network Gateway. Changing this forces a new resource to be created


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Virtual Network Gateway.

## Import

Virtual Network Gateways can be imported using the `resource id`, e.g.

```
terraform import azurestack_virtual_network_gateway.testGateway /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myGroup1/providers/Microsoft.Network/virtualNetworkGateways/myGateway1
```
