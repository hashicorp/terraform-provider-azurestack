---
layout: "azurestack"
page_title: "Azure Stack Provider: Authenticating via a Service Principal using a Client Secret"
sidebar_current: "docs-azurestack-auth-service-principal-client-secret"
description: |-
  This guide explains how to use a Service Principal and a Client Secret to authenticate with the Azure Stack Provider.

---

# Azure Stack Provider: Authenticating using a Service Principal using a Client Secret

Terraform supports authenticating to Azure Stack using [the Azure CLI](azure_cli.html) or a Service Principal, either using a Client Secret (which is detailed in this guide) or [using a Client Certificate](service_principal_client_certificate.html).

## Creating a Service Principal

A Service Principal is an application within Azure Active Directory whose authentication tokens can be used as the `client_id`, `client_secret`, and `tenant_id` fields needed by Terraform (`subscription_id` can be independently recovered from your Azure account details).

### Creating a Service Principal in the Azure Portal

~> **NOTE:** This needs to be completed in the main Azure (Public) Portal - not the Azure Stack Portal.

There are two tasks needed to create a Service Principal via [the Azure Portal](https://portal.azure.com):

 1. Create an Application in Azure Active Directory (which acts as a Service Principal)
 2. Grant the Application access to manage resources in your Azure Subscription

### 1. Creating an Application in Azure Active Directory

Firstly navigate to [the **Azure Active Directory** overview](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview) within the Azure Portal - [then select the **App Registration** blade](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps/RegisteredApps/Overview) and click **Endpoints** at the top of the **App Registration** blade. A list of URIs will be displayed and you need to locate the URI for **OAUTH 2.0 AUTHORIZATION ENDPOINT** which contains a GUID. This is your Tenant ID / the `tenant_id` field mentioned above.

Next, navigate back to [the **App Registration** blade](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps/RegisteredApps/Overview) - from here we'll create the Application in Azure Active Directory. To do this click **Add** at the top to add a new Application within Azure Active Directory. On this page, set the following values then press **Create**:

- **Name** - this is a friendly identifier and can be anything (e.g. "Terraform")
- **Application Type** - this should be set to "Web app / API"
- **Sign-on URL** - this can be anything, providing it's a valid URI (e.g. https://terra.form)

Finally need to create a Password for the Azure Active Directory Application - to do this select **Settings** and then **Keys**. This screen displays the Passwords (Client Secrets) and Public Keys (Client Certificates) which are associated with this Azure Active Directory Application.

Enter a description for the Key and select when this password should expire - and then press **Save**. At this point the Password should be displayed - you'll need to copy it now, since it's only displayed once - which is the `client_secret`.

### 2. Granting the Application access to manage resources in your Azure and Azure Stack Subscriptions

Once the Application exists in Azure Active Directory - we can grant it permissions to modify resources in the Subscription. To do this, [navigate to the **Subscriptions** blade within the Azure Portal](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade), then select the Subscription you wish to use, then click **Access Control (IAM)**, and finally **Add**.

~> **NOTE:**  This will only give SPN access to your Azure Subscription - This is **NOT** required to interact with Azure Stack. To allow SPN access to Azure Stack you need to do it under Azure Stack Subscription [navigate to the **Subscriptions** blade within the Azure Stack Portal](https://portal.{region}.{domain}/#blade/Microsoft_Azure_Billing/SubscriptionsBlade), then select the Subscription you wish to use, then click **Access Control (IAM)**, and finally **Add**.

Firstly, specify a Role which grants the appropriate permissions needed for the Service Principal (for example, `Contributor` will grant Read/Write on all resources in the Subscription). There's more information about [the built in roles available here](https://azure.microsoft.com/en-gb/documentation/articles/role-based-access-built-in-roles/).

Secondly, search for and select the name of the Application created in Azure Active Directory to assign it this role - then press **Save**.

### 3. Configuring the Service Principal in Terraform

As we've obtained the credentials for this Service Principal - it's possible to configure it in a few different ways.

When storing the credentials as Environment Variables, for example:

```bash
$ export ARM_ENDPOINT="00000000-0000-0000-0000-000000000000"
$ export ARM_CLIENT_ID="00000000-0000-0000-0000-000000000000"
$ export ARM_CLIENT_SECRET="00000000-0000-0000-0000-000000000000"
$ export ARM_SUBSCRIPTION_ID="00000000-0000-0000-0000-000000000000"
$ export ARM_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

The following Provider block can be specified - where `0.5.0` is the version of the Azure Stack Provider that you'd like to use:

```
provider "azurestack" {
  # Whilst version is optional, we /strongly recommend/ using it to pin the version of the Provider being used
  version = "=0.5.0"
}
```

More information on [the fields supported in the Provider block can be found here](../index.html#argument-reference).

---

It's also possible to configure these variables either in-line or from using variables in Terraform (as the `client_secret` is in this example), like so:

~> **NOTE:** We'd recommend not defining these variables in-line since they could easily be checked into Source Control.

```
variable "client_secret" {}

provider "azurestack" {
  # Whilst version is optional, we /strongly recommend/ using it to pin the version of the Provider being used
  version = "=0.5.0"

  arm_endpoint    = "https://management.region.myazurestack.com"
  subscription_id = "00000000-0000-0000-0000-000000000000"
  client_id       = "00000000-0000-0000-0000-000000000000"
  client_secret   = "${var.client_secret}"
  tenant_id       = "00000000-0000-0000-0000-000000000000"
}
```

More information on [the fields supported in the Provider block can be found here](../index.html#argument-reference).
