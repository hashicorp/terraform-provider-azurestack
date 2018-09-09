---
layout: "azurestack"
page_title: "Azure Stack: azurestack_virtual_network_gateway_connection"
sidebar_current: "docs-azurestack-resource-network-virtual-network-gateway-connection"
description: |-
  Manages a connection in an existing Virtual Network Gateway.
---

# azurestack_virtual_network_gateway_connection

Manages a connection in an existing Virtual Network Gateway.

## Example Usage

### Site-to-Site connection

The following example shows a connection between an Azure virtual network
and an on-premises VPN device and network.

```hcl
resource "azurestack_resource_group" "test" {
  name     = "test"
  location = "West US"
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

resource "azurestack_local_network_gateway" "onpremise" {
  name                = "onpremise"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  gateway_address     = "168.62.225.23"
  address_space       = ["10.1.1.0/24"]
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

  active_active = false
  enable_bgp    = false
	sku           = "Basic"

  ip_configuration {
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = "${azurestack_subnet.test.id}"
  }
}

resource "azurestack_virtual_network_gateway_connection" "onpremise" {
  name                = "onpremise"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  type                       = "IPsec"
  virtual_network_gateway_id = "${azurestack_virtual_network_gateway.test.id}"
  local_network_gateway_id   = "${azurestack_local_network_gateway.onpremise.id}"

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
```

### VNet-to-VNet connection

The following example shows a connection between two Azure virtual network
in different locations/regions.

```hcl
resource "azurestack_resource_group" "us" {
    name = "us"
    location = "East US"
}

resource "azurestack_virtual_network" "us" {
  name = "us"
  location = "${azurestack_resource_group.us.location}"
  resource_group_name = "${azurestack_resource_group.us.name}"
  address_space = ["10.0.0.0/16"]
}

resource "azurestack_subnet" "us_gateway" {
  name = "GatewaySubnet"
  resource_group_name = "${azurestack_resource_group.us.name}"
  virtual_network_name = "${azurestack_virtual_network.us.name}"
  address_prefix = "10.0.1.0/24"
}

resource "azurestack_public_ip" "us" {
  name = "us"
  location = "${azurestack_resource_group.us.location}"
  resource_group_name = "${azurestack_resource_group.us.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "us" {
  name = "us-gateway"
  location = "${azurestack_resource_group.us.location}"
  resource_group_name = "${azurestack_resource_group.us.name}"

  type = "Vpn"
  vpn_type = "RouteBased"
	sku = "Basic"

  ip_configuration {
    public_ip_address_id = "${azurestack_public_ip.us.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id = "${azurestack_subnet.us_gateway.id}"
  }
}

resource "azurestack_resource_group" "europe" {
  name = "europe"
  location = "West Europe"
}

resource "azurestack_virtual_network" "europe" {
  name = "europe"
  location = "${azurestack_resource_group.europe.location}"
  resource_group_name = "${azurestack_resource_group.europe.name}"
  address_space = ["10.1.0.0/16"]
}

resource "azurestack_subnet" "europe_gateway" {
  name = "GatewaySubnet"
  resource_group_name = "${azurestack_resource_group.europe.name}"
  virtual_network_name = "${azurestack_virtual_network.europe.name}"
  address_prefix = "10.1.1.0/24"
}

resource "azurestack_public_ip" "europe" {
  name = "europe"
  location = "${azurestack_resource_group.europe.location}"
  resource_group_name = "${azurestack_resource_group.europe.name}"
  public_ip_address_allocation = "Dynamic"
}

resource "azurestack_virtual_network_gateway" "europe" {
  name = "europe-gateway"
  location = "${azurestack_resource_group.europe.location}"
  resource_group_name = "${azurestack_resource_group.europe.name}"

  type = "Vpn"
  vpn_type = "RouteBased"
  sku = "Basic"

  ip_configuration {
    public_ip_address_id = "${azurestack_public_ip.europe.id}"
    private_ip_address_allocation = "Dynamic"
    subnet_id = "${azurestack_subnet.europe_gateway.id}"
  }
}

resource "azurestack_virtual_network_gateway_connection" "us_to_europe" {
  name = "us-to-europe"
  location = "${azurestack_resource_group.us.location}"
  resource_group_name = "${azurestack_resource_group.us.name}"

  type = "Vnet2Vnet"
  virtual_network_gateway_id = "${azurestack_virtual_network_gateway.us.id}"
  peer_virtual_network_gateway_id = "${azurestack_virtual_network_gateway.europe.id}"

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}

resource "azurestack_virtual_network_gateway_connection" "europe_to_us" {
  name = "europe-to-us"
  location = "${azurestack_resource_group.europe.location}"
  resource_group_name = "${azurestack_resource_group.europe.name}"

  type = "Vnet2Vnet"
  virtual_network_gateway_id = "${azurestack_virtual_network_gateway.europe.id}"
  peer_virtual_network_gateway_id = "${azurestack_virtual_network_gateway.us.id}"

  shared_key = "4-v3ry-53cr37-1p53c-5h4r3d-k3y"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection. Changing the name forces a
    new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to
    create the connection Changing the name forces a new resource to be created.

* `location` - (Required) The location/region where the connection is
    located. Changing this forces a new resource to be created.

* `type` - (Required) The type of connection. Valid options are `IPsec`
    (Site-to-Site), `ExpressRoute` (ExpressRoute), and `Vnet2Vnet` (VNet-to-VNet).
    Each connection type requires different mandatory arguments (refer to the
    examples above). Changing the connection type will force a new connection
    to be created.

* `virtual_network_gateway_id` - (Required) The ID of the Virtual Network Gateway
    in which the connection will be created. Changing the gateway forces a new
    resource to be created.

* `authorization_key` - (Optional) The authorization key associated with the
    Express Route Circuit. This field is required only if the type is an
    ExpressRoute connection.

* `express_route_circuit_id` - (Optional) The ID of the Express Route Circuit
    when creating an ExpressRoute connection (i.e. when `type` is `ExpressRoute`).
    The Express Route Circuit can be in the same or in a different subscription.

* `peer_virtual_network_gateway_id` - (Optional) The ID of the peer virtual
    network gateway when creating a VNet-to-VNet connection (i.e. when `type`
    is `Vnet2Vnet`). The peer Virtual Network Gateway can be in the same or
    in a different subscription.

* `local_network_gateway_id` - (Optional) The ID of the local network gateway
    when creating Site-to-Site connection (i.e. when `type` is `IPsec`).

* `routing_weight` - (Optional) The routing weight. Defaults to `10`.

* `shared_key` - (Optional) The shared IPSec key. A key must be provided if a
    Site-to-Site or VNet-to-VNet connection is created whereas ExpressRoute
    connections do not need a shared key.

* `enable_bgp` - (Optional) If `true`, BGP (Border Gateway Protocol) is enabled
    for this connection. Defaults to `false`.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The connection ID.

## Import

Virtual Network Gateway Connections can be imported using their `resource id`, e.g.

```
terraform import azurestack_virtual_network_gateway_connection.testConnection /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myGroup1/providers/Microsoft.Network/connections/myConnection1
```
