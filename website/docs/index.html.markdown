---
layout: "azurestack"
page_title: "Provider: Azure Stack"
sidebar_current: "docs-azurestack-index"
description: |-
  The Azure Stack Provider is used to manage resources in Azure Stack via the Azure Resource Manager API's.

---

# Azure Stack Provider

The Azure Stack Provider is used to manage resources in Azure Stack via the Azure Resource Manager API's.

Use the navigation to the left to read about the available resources.

# Creating Credentials

Terraform supports authenticating to Azure Stack through a Service Principal - [this page explains how to Create a Service Principal](authenticating_via_service_principal.html).

## Example Usage

```hcl
# Configure the Azure Provider
provider "azurestack" { }

# Create a resource group
resource "azurestack_resource_group" "network" {
  name     = "production"
  location = "West US"
}

# Create a virtual network within the resource group
resource "azurestack_virtual_network" "network" {
  name                = "production-network"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.network.location}"
  resource_group_name = "${azurestack_resource_group.network.name}"

  subnet {
    name           = "subnet1"
    address_prefix = "10.0.1.0/24"
  }

  subnet {
    name           = "subnet2"
    address_prefix = "10.0.2.0/24"
  }

  subnet {
    name           = "subnet3"
    address_prefix = "10.0.3.0/24"
  }
}
```

## Argument Reference

The following arguments are supported:

* `arm_endpoint` - (Optional) The Azure Resource Manager API Endpoint for
  your Azure Stack instance, such as `https://management.westus.mydomain.com`.
  It can also be sourced from the `ARM_ENDPOINT` environment variable.

* `subscription_id` - (Optional) The subscription ID to use. It can also
  be sourced from the `ARM_SUBSCRIPTION_ID` environment variable.

* `client_id` - (Optional) The client ID to use. It can also be sourced from
  the `ARM_CLIENT_ID` environment variable.

* `client_secret` - (Optional) The client secret to use. It can also be sourced from
  the `ARM_CLIENT_SECRET` environment variable.

* `tenant_id` - (Optional) The tenant ID to use. It can also be sourced from the
  `ARM_TENANT_ID` environment variable.

* `skip_credentials_validation` - (Optional) Prevents the provider from validating
  the given credentials. When set to `true`, `skip_provider_registration` is assumed.
  It can also be sourced from the `ARM_SKIP_CREDENTIALS_VALIDATION` environment
  variable; defaults to `false`.

* `skip_provider_registration` - (Optional) Prevents the provider from registering
  the ARM provider namespaces, this can be used if you don't wish to give the Active
  Directory Application permission to register resource providers. It can also be
  sourced from the `ARM_SKIP_PROVIDER_REGISTRATION` environment variable; defaults
  to `false`.

## Testing

The following Environment Variables must be set to run the acceptance tests:

~> **NOTE:** The Acceptance Tests require the use of a Service Principal.

* `ARM_ENDPOINT` - The Azure Resource Manager API Endpoint for Azure Stack.
* `ARM_SUBSCRIPTION_ID` - The ID of the Azure Subscription in which to run the Acceptance Tests.
* `ARM_CLIENT_ID` - The Client ID of the Service Principal.
* `ARM_CLIENT_SECRET` - The Client Secret associated with the Service Principal.
* `ARM_TENANT_ID` - The Tenant ID to use.
* `ARM_TEST_LOCATION` - The Azure Stack Location to provision resources in for the Acceptance Tests.
