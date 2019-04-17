---
layout: "azurestack"
page_title: "Azure Stack Provider: Authenticating using the Azure CLI"
sidebar_current: "docs-azurestack-auth-service-principal-azure-cli"
description: |-
  This guide explains how to authenticate using the Azure CLI with the Azure Stack Provider.

---

# Azure Stack Provider: Authenticating using the Azure CLI

Terraform supports authenticating to Azure Stack using the Azure CLI, a Service Principal (either using a Client Secret or a Client Certificate).

Terraform supports authenticating to Azure Stack using the Azure CLI (which is detailed in this guide) or a Service Principal, either [using a Client Secret](service_principal_client_secret.html) or [using a Client Certificate](service_principal_client_certificate.html).

~> **NOTE:** Authenticating via the Azure CLI is only supported when using a User Account. If you're using a Service Principal (for example via `az login --service-principal`) you should instead [authenticate via the Service Principal directly](service_principal_client_secret.html).

This guide assumes that the Certificate being used for Azure Stack is valid, or has been trusted on your machine ([instructions for trusting the Azure Stack Certificate can be found here](https://docs.microsoft.com/en-us/azure/azure-stack/user/azure-stack-version-profiles-azurecli2#trust-the-azure-stack-ca-root-certificate)).

---

We need to add the details for your Azure Stack Configuration to the Azure CLI:

```shell
$ az cloud register -n AzureStack --endpoint-resource-manager "https://management.region.mycloud.com" --suffix-storage-endpoint "region.mycloud.com" --suffix-keyvault-dns ".vault.region.mycloud.com"
```

-> Note: the values used will differ from the ones specified in the example above - as such you may need to contact your Azure Stack Administrator to determine the values required to connect to your Azure Stack instance.

Once that's done we can now switch to using the newly registered `AzureStack` cloud:

```shell
$ az cloud set --name AzureStack
```

Next you'll need to configure the Profile used for Azure Stack:

```shell
$ az cloud update --profile 2018-03-01-hybrid
```

-> **NOTE:** If you're using a version of Azure Stack prior to build 1808 - you'll need to use the Profile version `2017-03-09-profile`.

---

At this point we should be able to log into the Azure Stack instance using the Azure CLI:

```shell
$ az login
```

~> **NOTE:** Authenticating via the Azure CLI is only supported when using a User Account. If you're using a Service Principal (for example via `az login --service-principal`) you should instead [authenticate via the Service Principal directly](service_principal_client_secret.html).

This will prompt you to open a web browser, as shown below:

```shell
$ az login
Note, we have launched a browser for you to login. For old experience with device code, use "az login --use-device-code
```

Once logged in, it's possible to list the Subscriptions associated with the account via:

```shell
$ az account list
```

The output (similar to below) will display one or more Subscriptions:

```json
[
  {
    "cloudName": "AzureStack",
    "id": "00000000-0000-0000-0000-000000000000",
    "isDefault": true,
    "name": "Example Subscription",
    "state": "Enabled",
    "tenantId": "00000000-0000-0000-0000-000000000000",
    "user": {
      "name": "user@example.com",
      "type": "user"
    }
  }
]
```

In the snippet above, `id` refers to the Subscription ID and `isDefault` refers to whether this Subscription is configured as the default.

~> **Note:** When authenticating via the Azure CLI, Terraform will automatically connect to the Default Subscription. Therefore, if you have multiple subscriptions on the account, you may need to set the Default Subscription, via:

```shell
$ az account set --subscription="SUBSCRIPTION_ID"
```

## Configuring the Provider block when using the Azure CLI

When authenticating using the Azure CLI most of the details required can be inferred, as such to use the Default Subscription configured in the Azure CLI you should be able to use the following Provider Block:

```
provider "azurestack" {
  # whilst the version attribute is optional, we recommend pinning the Provider version being used
  version = "=0.5.0"
}
```

If you're using multiple subscriptions (or have access to multiple tenants) - it's possible to specify the Subscription ID to target a specific Subscription:

```
provider "azurestack" {
  # whilst the version attribute is optional, we recommend pinning the Provider version being used
  version         = "=0.5.0"
  subscription_id = "00000000-0000-0000-0000-000000000000"
}
```

More information on [the fields supported in the Provider block can be found here](../index.html#argument-reference).

~> **NOTE:** If you're previously authenticated using a Service Principal (configured via Environment Variables) - you must remove the `ARM_*` prefixed Environment Variables in order to be able to authenticate using the Azure CLI.
