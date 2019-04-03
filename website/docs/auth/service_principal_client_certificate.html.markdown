---
layout: "azurestack"
page_title: "Azure Stack Provider: Authenticating via a Service Principal using a Client Certificate"
sidebar_current: "docs-azurestack-auth-service-principal-client-certificate"
description: |-
  This guide explains how to use a Service Principal and a Client Certificate to authenticate with the Azure Stack Provider.

---

# Azure Stack Provider: Authenticating using a Service Principal using a Client Certificate

Terraform supports authenticating to Azure Stack using [the Azure CLI](azure_cli.html) or a Service Principal, either [using a Client Secret](service_principal_client_secret.html) or using a Client Certificate (which is detailed in this guide).

## Creating a Service Principal

A Service Principal is an application within Azure Active Directory which can have authentication tokens associated with it. In this example we'll generate a new certificate, then create and assign it to a Service Principal; so that it can be used for authentication.

### Creating a Service Principal in the Azure Portal

~> **NOTE:** This needs to be completed in the main Azure (Public) Portal - not the Azure Stack Portal.

There are three tasks needed to create a Service Principal via [the Azure Portal](https://portal.azure.com):

 1. Generating a Certificate which can be used for Authentication
 2. Create an Application in Azure Active Directory (which acts as a Service Principal) and then associating the Certificate with it
 2. Grant the Application access to manage resources in your Azure Subscription

### 1. Generating a Certificate

Firstly we need to create a certificate which can be used for authentication. To do that we're going to generate a Certificate Signing Request (also known as a CSR) using `openssl` (this can also be achieved using PowerShell, however that's outside the scope of this document):

```bash
$ openssl req \
   -newkey rsa:4096 -nodes -keyout "service-principal.key" \
   -out "service-principal.csr"
```

We can now sign that Certificate Signing Request, in this example we're going to self-sign this certificate using the Key we just generated; however it's also possible to do this using a Certificate Authority. In order to do that we're again going to use `openssl`:

```bash
$ openssl x509 \
  -signkey "service-principal.key" \
  -in "service-principal.csr" \
  -req -days 365 -out "service-principal.crt"
```

Finally we can generate a PFX file which can be used to authenticate with Azure:

```
$ openssl pkcs12 -export -out "service-principal.pfx" -inkey "service-principal.key" -in "service-principal.crt"
```

Now that we've generated a certificate, we can create the Azure Active Directory application.

### 2. Creating an Application in Azure Active Directory

Firstly navigate to [the **Azure Active Directory** overview](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview) within the Azure Portal - [then select the **App Registration** blade](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps/RegisteredApps/Overview) and click **Endpoints** at the top of the **App Registration** blade. A list of URIs will be displayed and you need to locate the URI for **OAUTH 2.0 AUTHORIZATION ENDPOINT** which contains a GUID. This is your Tenant ID / the `tenant_id` field mentioned above.

Next, navigate back to [the **App Registration** blade](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps/RegisteredApps/Overview) - from here we'll create the Application in Azure Active Directory. To do this click **Add** at the top to add a new Application within Azure Active Directory. On this page, set the following values then press **Create**:

- **Name** - this is a friendly identifier and can be anything (e.g. "Terraform")
- **Application Type** - this should be set to "Web app / API"
- **Sign-on URL** - this can be anything, providing it's a valid URI (e.g. https://terra.form)

Once that's done - select the Application you just created in [the **App Registration** blade](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps/RegisteredApps/Overview). At the top of this page, the "Application ID" GUID is the `client_id` you'll need.

Finally need to associate the Client Certificate with the Azure Active Directory Application - to do this select **Settings** and then **Keys**. This screen displays the Passwords (Client Secrets) and Public Keys (Client Certificates) which are associated with this Azure Active Directory Application.

The Public Key associated with the generated Certificate can be uploaded by selecting **Upload Public Key**, selecting the file which should be uploaded (in the example above, this'd be `service-principal.crt`) - and then hitting **Save**.

### 3. Granting the Application access to manage resources in your Azure and Azure Stack Subscriptions

Once the Application exists in Azure Active Directory - we can grant it permissions to modify resources in the Subscription. To do this, [navigate to the **Subscriptions** blade within the Azure Portal](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade), then select the Subscription you wish to use, then click **Access Control (IAM)**, and finally **Add**.

~> **NOTE:**  This will only give SPN access to your Azure Subscription - This is **NOT** required to interact with Azure Stack. To allow SPN access to Azure Stack you need to do it under Azure Stack Subscription [navigate to the **Subscriptions** blade within the Azure Stack Portal](https://portal.{region}.{domain}/#blade/Microsoft_Azure_Billing/SubscriptionsBlade), then select the Subscription you wish to use, then click **Access Control (IAM)**, and finally **Add**.

Firstly, specify a Role which grants the appropriate permissions needed for the Service Principal (for example, `Contributor` will grant Read/Write on all resources in the Subscription). There's more information about [the built in roles available here](https://azure.microsoft.com/en-gb/documentation/articles/role-based-access-built-in-roles/).

Secondly, search for and select the name of the Application created in Azure Active Directory to assign it this role - then press **Save**.

At this point the newly created Azure Active Directory Application should be associated with the Certificate that we generated earlier (which can be used as a Client Certificate) - and should have permissions to the Azure Subscription.

It should then be possible to configure these credentials in Terraform, either by using setting the relevant Environment Variables:

```bash
export ARM_ENDPOINT="https://management.region.mycloud.com"
export ARM_CLIENT_CERTIFICATE_PASSWORD="hello-world"
export ARM_CLIENT_CERTIFICATE_PATH="/Users/myuser/keys/service-principal.pfx"
export ARM_CLIENT_ID="00000000-0000-0000-0000-000000000000"
export ARM_SUBSCRIPTION_ID="00000000-0000-0000-0000-000000000000"
export ARM_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

and then using the following Provider Block:

```hcl
provider "azurestack" {
  # whilst the `version` attribute is optional, we'd recommend pinning to a particular version
  version = "=0.5.0"
}
```

Alternatively you can define these fields within the Provider Block:

```hcl
provider "azurestack" {
  # whilst the `version` attribute is optional, we'd recommend pinning to a particular version
  version = "=0.5.0"

  arm_endpoint                = "https://management.region.mycloud.com"
  client_id                   = "00000000-0000-0000-0000-000000000000"
  client_certificate_password = "my-password"
  client_certificate_path     = "./service-principal.pfx"
  subscription_id             = "00000000-0000-0000-0000-000000000000"
  tenant_id                   = "00000000-0000-0000-0000-000000000000"
}
```

More information on [the fields supported in the Provider block can be found here](../index.html#argument-reference).