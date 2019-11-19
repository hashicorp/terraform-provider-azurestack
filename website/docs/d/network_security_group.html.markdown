---
subcategory: ""
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_network_security_group"
sidebar_current: "docs-azurestack-datasource-network-security-group"
description: |-
  Get information about the specified Network Security Group.
---

# Data Source: azurestack_network_security_group

Use this data source to access the properties of a Network Security Group.

## Example Usage

```hcl
data "azurestack_network_security_group" "test" {
  name                = "${azurestack_network_security_group.test.name}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

output "location" {
  value = "${data.azurestack_network_security_group.test.location}"
}
```

## Argument Reference

* `name` - (Required) Specifies the Name of the Network Security Group.
* `resource_group_name` - (Required) Specifies the Name of the Resource Group within which the Network Security Group exists


## Attributes Reference

* `id` - The ID of the Network Security Group.

* `location` - The supported Azure location where the resource exists.

* `security_rule` - One or more `security_rule` blocks as defined below.

* `tags` - A mapping of tags assigned to the resource.


The `security_rule` block supports:

* `name` - The name of the security rule.

* `description` - The description for this rule.

* `protocol` - The network protocol this rule applies to.

* `source_port_range` - The Source Port or Range.

* `destination_port_range` - The Destination Port or Range.

* `source_address_prefix` - CIDR or source IP range or * to match any IP.

* `destination_address_prefix` - CIDR or destination IP range or * to match any IP.

* `access` - Is network traffic is allowed or denied?

* `priority` - The priority of the rule

* `direction` - The direction specifies if rule will be evaluated on incoming or outgoing traffic.
