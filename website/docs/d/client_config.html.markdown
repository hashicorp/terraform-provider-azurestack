---
subcategory: "Base"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_client_config"
description: |-
  Get information about the configuration of the azurestack provider.
---

# Data Source: azurestack_client_config

Use this data source to access the configuration of the Azure Stack
provider.

## Example Usage

```hcl
data "azurestack_client_config" "current" {}

output "account_id" {
  value = data.azurestack_client_config.current.client_id
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `client_id` is set to the Azure Client ID (Application Object ID).
* `tenant_id` is set to the Azure Tenant ID.
* `subscription_id` is set to the Azure Subscription ID.
* `object_id` is set to the Azure Object ID.

-> **Note:** This is only applicable when not using ADFS.
