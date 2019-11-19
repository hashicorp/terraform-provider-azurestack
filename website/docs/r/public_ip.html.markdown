---
subcategory: "Network"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_public_ip"
sidebar_current: "docs-azurestack-resource-network-public-ip"
description: |-
  Manages a Public IP Address.
---

# azurestack_public_ip

Manages a Public IP Address.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "resourceGroup1"
  location = "West US"
}

resource "azurestack_public_ip" "test" {
  name                         = "acceptanceTestPublicIp1"
  location                     = "West US"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"

  tags = {
    environment = "Production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Public IP resource . Changing this forces a
    new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to
    create the public ip.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `public_ip_address_allocation` - (Required) Defines whether the IP address is static or dynamic. Options are Static or Dynamic.

~> **Note** `Dynamic` Public IP Addresses aren't allocated until they're assigned to a resource (such as a Virtual Machine or a Load Balancer) by design within Azure - [more information is available below](#ip_address).

* `idle_timeout_in_minutes` - (Optional) Specifies the timeout for the TCP idle connection. The value can be set between 4 and 30 minutes.

* `domain_name_label` - (Optional) Label for the Domain Name. Will be used to make up the FQDN.  If a domain name label is specified, an A DNS record is created for the public IP in the Microsoft Azure DNS system.

* `reverse_fqdn` - (Optional) A fully qualified domain name that resolves to this public IP address. If the reverseFqdn is specified, then a PTR DNS record is created pointing from the IP address in the in-addr.arpa domain to the reverse FQDN.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The Public IP ID.
* `ip_address` - The IP address value that was allocated.

~> **Note** `Dynamic` Public IP Addresses aren't allocated until they're attached to a device (e.g. a Virtual Machine/Load Balancer). Instead you can obtain the IP Address once the the Public IP has been assigned via the `azurestack_public_ip` Data Source (not currently available)

* `fqdn` - Fully qualified domain name of the A DNS record associated with the public IP. This is the concatenation of the domainNameLabel and the regionalized DNS zone


## Import

Public IPs can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_public_ip.myPublicIp /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/publicIPAddresses/myPublicIpAddress1
```
