---
subcategory: "DNS"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_dns_a_record"
sidebar_current: "docs-azurestack-resource-dns-a-record"
description: |-
  Manages a DNS A Record.
---

# azurestack_dns_a_record

Enables you to manage DNS A Records within Azure DNS.

## Example Usage

```hcl
resource "azurestack_resource_group" "test" {
  name     = "acceptanceTestResourceGroup1"
  location = "West US"
}

resource "azurestack_dns_zone" "test" {
  name                = "mydomain.com"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_dns_a_record" "test" {
  name                = "test"
  zone_name           = "${azurestack_dns_zone.test.name}"
  resource_group_name = "${azurestack_resource_group.test.name}"
  ttl                 = 300
  records             = ["10.0.180.17"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the DNS A Record.

* `resource_group_name` - (Required) Specifies the resource group where the resource exists. Changing this forces a new resource to be created.

* `zone_name` - (Required) Specifies the DNS Zone where the resource exists. Changing this forces a new resource to be created.

* `TTL` - (Required) The Time To Live (TTL) of the DNS record.

* `records` - (Required) List of IPv4 Addresses.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The DNS A Record ID.

## Import

A records can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_dns_a_record.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Network/dnsZones/zone1/A/myrecord1
```
