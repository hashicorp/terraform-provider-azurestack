---
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_client_config"
sidebar_current: "docs-azurestack-datasource-client-config"
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
  value = "${data.azurestack_client_config.current.service_principal_application_id}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `client_id` is set to the Azure Client ID (Application Object ID).
* `tenant_id` is set to the Azure Tenant ID.
* `subscription_id` is set to the Azure Subscription ID.
* `service_principal_application_id` is the Service Principal Application ID.
* `service_principal_object_id` is the Service Principal Object ID.

~> **Note:** To better understand "application" and "service principal", please read
[Application and service principal objects in Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-application-objects).
