---
subcategory: ""
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_subnet"
sidebar_current: "docs-azurestack-datasource-subnet"
description: |-
  Get information about the specified Subnet located within a Virtual Network.
---

# Data Source: azurestack_subnet

Use this data source to access the properties of an Azure Subnet located within a Virtual Network.

## Example Usage

```hcl
data "azurestack_subnet" "test" {
  name                 = "backend"
  virtual_network_name = "production"
  resource_group_name  = "networking"
}

output "subnet_id" {
  value = "${data.azurestack_subnet.test.id}"
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the Subnet.
* `virtual_network_name` - (Required) Specifies the name of the Virtual Network this Subnet is located within.
* `resource_group_name` - (Required) Specifies the name of the resource group the Virtual Network is located in.

## Attributes Reference

* `id` - The ID of the Subnet.
* `address_prefix` - The address prefix used for the subnet.
* `network_security_group_id` - The ID of the Network Security Group associated with the subnet.
* `route_table_id` - The ID of the Route Table associated with this subnet.
* `ip_configurations` - The collection of IP Configurations with IPs within this subnet.
