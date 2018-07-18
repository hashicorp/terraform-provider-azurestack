---
layout: "azurestack"
page_title: "Azure Stack Provider: Authenticating via a Service Principal"
sidebar_current: "docs-azurestack-index-authentication-service-principal"
description: |-
  This guide will cover how to use a Service Principal (Shared Account) as authentication for the Azure Stack Provider.

---

# Azure Stack Provider: Authenticating using a Service Principal

Terraform supports authenticating to Azure Stack through a Service Principal. At this time this is the only supported authentication method for Azure Stack.

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

Once that's done - select the Application you just created in [the **App Registration** blade](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps/RegisteredApps/Overview). At the top of this page, the "Application ID" GUID is the `client_id` you'll need.

Finally, we can create the `client_secret` by selecting **Keys** and then generating a new key by entering a description, selecting how long the `client_secret` should be valid for - and finally pressing **Save**. This value will only be visible whilst on the page, so be sure to copy it now (otherwise you'll need to regenerate a new key).

### 2. Granting the Application access to manage resources in your Azure and Azure Stack Subscriptions

Once the Application exists in Azure Active Directory - we can grant it permissions to modify resources in the Subscription. To do this, [navigate to the **Subscriptions** blade within the Azure Portal](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade), then select the Subscription you wish to use, then click **Access Control (IAM)**, and finally **Add**.

~> **NOTE:**  This will only give SPN access to your Azure Subscription - This is **NOT** required to interact with Azure Stack. To allow SPN access to Azure Stack you need to do it under Azure Stack Subscription [navigate to the **Subscriptions** blade within the Azure Stack Portal](https://portal.{region}.{domain}/#blade/Microsoft_Azure_Billing/SubscriptionsBlade), then select the Subscription you wish to use, then click **Access Control (IAM)**, and finally **Add**.

Firstly, specify a Role which grants the appropriate permissions needed for the Service Principal (for example, `Contributor` will grant Read/Write on all resources in the Subscription). There's more information about [the built in roles available here](https://azure.microsoft.com/en-gb/documentation/articles/role-based-access-built-in-roles/).

Secondly, search for and select the name of the Application created in Azure Active Directory to assign it this role - then press **Save**.

## Configuring your Service Principal

Service Principals can be configured in Terraform in one of two ways, either as Environment Variables or in the Provider block. Please see [this section](index.html#argument-reference) for an example of which fields are available and can be specified either through Environment Variables - or in the Provider Block.

### Example of Environment Variables

- `variables.tf`

  ```hcl
  variable "arm_endpoint" {}
  variable "subscription_id" {}
  variable "client_id" {}
  variable "client_secret" {}
  variable "tenant_id" {}
  ```

- `example.tf`

  ```hcl
  provider "azurestack" {
    arm_endpoint    = "${var.arm_endpoint}"
    subscription_id = "${var.subscription_id}"
    client_id       = "${var.client_id}"
    client_secret   = "${var.client_secret}"
    tenant_id       = "${var.tenant_id}"
  }
  ```

- `terraform.tfvars`

  ```hcl
  # Configure the Azure Stack Provider
  arm_endpoint    = "https://management.{region}.{domain}"
  subscription_id = "xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx"
  client_id       = "{applicationId}@{tenantDomain}"
  client_secret   = "{applicationPassword}"
  tenant_id       = "xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx"
  ```

### Example of Provider Block

- `example.tf`

  ```hcl
  # Configure the Azure Stack Provider
  arm_endpoint    = "https://management.{region}.{domain}"
  subscription_id = "xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx"
  client_id       = "{applicationId}@{tenantDomain}"
  client_secret   = "{applicationPassword}"
  tenant_id       = "xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxx"
  ```