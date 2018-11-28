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

Terraform supports authenticating to Azure Stack using [the Azure CLI](auth/azure_cli.html) or a Service Principal (either using a [Client Secret](auth/service_principal_client_secret.html) or a [Client Certificate](auth/service_principal_client_certificate.html)).

## Example Usage

```hcl
# Configure the Azure Stack Provider
provider "azurestack" {
  # NOTE: we recommend pinning the version of the Provider which should be used in the Provider block
  # version = "=0.5.0"
}

# Create a resource group
resource "azurestack_resource_group" "test" {
  name     = "production"
  location = "West US"
}

# Create a virtual network within the resource group
resource "azurestack_virtual_network" "test" {
  name                = "production-network"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

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

* `arm_endpoint` - (Optional) The Azure Resource Manager Endpoint for your Azure Stack instance, for example `https://management.westus.mydomain.com`. This can also be sourceed from the `ARM_ENDPOINT` Environment Variable.

* `client_id` - (Optional) The Client ID which should be used. This can also be sourceed from the `ARM_CLIENT_ID` Environment Variable.

* `subscription_id` - (Optional) The Subscription ID which should be used. This can also be sourced from the `ARM_SUBSCRIPTION_ID` Environment Variable.

* `tenant_id` - (Optional) The Tenant ID which should be used. This can also be sourced from the `ARM_TENANT_ID` Environment Variable.

---

When authenticating as a Service Principal using a Client Certificate, the following fields can be set:

* `client_certificate_password` - (Optional) The password associated with the Client Certificate. This can also be sourced from the `ARM_CLIENT_CERTIFICATE_PASSWORD` Environment Variable.

* `client_certificate_path` - (Optional) The path to the Client Certificate associated with the Service Principal which should be used. This can also be sourced from the `ARM_CLIENT_CERTIFICATE_PATH` Environment Variable.

More information on [how to configure a Service Principal using a Client Certificate can be found in this guide](auth/service_principal_client_certificate.html).

---

When authenticating as a Service Principal using a Client Secret, the following fields can be set:

* `client_secret` - (Optional) The Client Secret which should be used. This can also be sourced from the `ARM_CLIENT_SECRET` Environment Variable.

More information on [how to configure a Service Principal using a Client Secret can be found in this guide](auth/service_principal_client_secret.html).

---

For some advanced scenarios, such as where more granular permissions are necessary - the following properties can be set:

* `skip_credentials_validation` - (Optional) Should the Azure Stack Provider skip verifying the credentials being used are valid? This can also be sourced from the `ARM_SKIP_CREDENTIALS_VALIDATION` Environment Variable. Defaults to `false`.

* `skip_provider_registration` - (Optional) Should the Azure Stack Provider skip registering any required Resource Providers? This can also be sourced from the `ARM_SKIP_PROVIDER_REGISTRATION` Environment Variable. Defaults to `false`.

## Testing

The following Environment Variables must be set to run the acceptance tests:

~> **NOTE:** The Acceptance Tests require the use of a Service Principal using a Client Secret.

* `ARM_ENDPOINT` - The Azure Resource Manager API Endpoint for Azure Stack.
* `ARM_SUBSCRIPTION_ID` - The ID of the Azure Subscription in which to run the Acceptance Tests.
* `ARM_CLIENT_ID` - The Client ID of the Service Principal.
* `ARM_CLIENT_SECRET` - The Client Secret associated with the Service Principal.
* `ARM_TENANT_ID` - The Tenant ID to use.
* `ARM_TEST_LOCATION` - The Azure Stack Location to provision resources in for the Acceptance Tests.
