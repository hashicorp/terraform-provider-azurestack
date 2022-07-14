<a href="https://terraform.io">
    <img src=".github/tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

# Terraform Provider for Azure Stack

* [Terraform Website](https://www.terraform.io)
* [AzureStack Provider Documentation](https://registry.terraform.io/providers/hashicorp/azurestack/latest/docs)
* [Slack Workspace for Contributors](https://terraform-azure.slack.com) ([Request Invite](https://join.slack.com/t/terraform-azure/shared_invite/enQtNDMzNjQ5NzcxMDc3LWNiY2ZhNThhNDgzNmY0MTM0N2MwZjE4ZGU0MjcxYjUyMzRmN2E5NjZhZmQ0ZTA1OTExMGNjYzA4ZDkwZDYxNDE))


## Compatibility

The Azure Stack provider is only compatible with specific profile versions as listed below:

| Azure Stack Profile Version | Supported Azure Stack Provider Versions |
| --------------------------- | --------------------------------------- |
| 2019-03-01                  | 0.8+                                    |
| 2017-10-01                  | 0.1-0.7                                 |

## Usage

```hcl
# We strongly recommend using the required_providers block to set the
# Azure Provider source and version being used
terraform {
  required_providers {
    azurestack = {
      source = "hashicorp/azurestack"
      version = "=1.0.0"
    }
  }
}

# Configure the Microsoft Azure Stack Provider
provider "azurestack" {
  features {}

  # More information on the authentication methods supported by
  # the AzureStack Provider can be found here:
  # https://registry.terraform.io/providers/hashicorp/azurestack/latest/docs

  # metadata_hostname = "..."
  # subscription_id   = "..."
  # client_id         = "..."
  # client_secret     = "..."
  # tenant_id         = "..."
}

# Create a resource group
resource "azurestack_resource_group" "example" {
  name     = "production-resources"
  location = "StackEnv"
}

# Create a virtual network in the production-resources resource group
resource "azurestack_virtual_network" "example" {
  name                = "example-network"
  resource_group_name = azurestack_resource_group.example.name
  location            = azurestack_resource_group.example.location
  address_space       = ["10.0.0.0/16"]
}
```

Further [usage documentation is available on the Terraform website](https://registry.terraform.io/providers/hashicorp/azurestack/latest/docs).

## Developer Requirements

* [Terraform](https://www.terraform.io/downloads.html) version 0.12.x + (but 1.x is recommended)
* [Go](https://golang.org/doc/install) version 1.17.x (to build the provider plugin)

### On Windows

If you're on Windows you'll also need:
* [Git Bash for Windows](https://git-scm.com/download/win)
* [Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm)

For *GNU32 Make*, make sure its bin path is added to PATH environment variable.*

For *Git Bash for Windows*, at the step of "Adjusting your PATH environment", please choose "Use Git and optional Unix tools from Windows Command Prompt".*

Or install via [Chocolatey](https://chocolatey.org/install) (`Git Bash for Windows` must be installed per steps above)

```powershell
choco install make golang terraform -y
refreshenv
```

You must run `Developing the Provider` commands in `bash` because `sh` scrips are invoked as part of these.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.16+ is **required**). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

First clone the repository to: `$GOPATH/src/github.com/hashicorp/terraform-provider-azurestack`

```sh
$ mkdir -p $GOPATH/src/github.com/hashicorp; cd $GOPATH/src/github.com/hashicorp
$ git clone git@github.com:hashicorp/terraform-provider-azurestack
$ cd $GOPATH/src/github.com/hashicorp/terraform-provider-azurestack
```

Once inside the provider directory, you can run `make tools` to install the dependent tooling required to compile the provider.

At this point you can compile the provider by running `make build`, which will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-azurestack
...
```

You can also cross-compile if necessary:

```sh
GOOS=windows GOARCH=amd64 make build
```

In order to run the `Unit Tests` for the provider, you can run:

```sh
$ make test
```

The majority of tests in the provider are `Acceptance Tests` - which provisions real resources in Azure. It's possible to run the entire acceptance test suite by running `make testacc` - however it's likely you'll want to run a subset, which you can do using a prefix, by running:

```sh
make acctests SERVICE='<service>' TESTARGS='-run=<nameOfTheTest>' TESTTIMEOUT='60m'
```

* `<service>` is the name of the folder which contains the file with the test(s) you want to run. The available folders are found in `azurestack/internal/services/`. So examples are `mssql`, `compute` or `mariadb`
* `<nameOfTheTest>` should be self-explanatory as it is the name of the test you want to run. An example could be `TestAccMsSqlServerExtendedAuditingPolicy_basic`. Since `-run` can be used with regular expressions you can use it to specify multiple tests like in `TestAccMsSqlServerExtendedAuditingPolicy_` to run all tests that match that expression

The following Environment Variables must be set in your shell prior to running acceptance tests:

- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`
- `ARM_METADATA_HOST`
- `ARM_TEST_LOCATION`

---

## Developer: Using the locally compiled Azure Provider binary

When using Terraform 0.14 and later, after successfully compiling the Azure Provider, you must [instruct Terraform to use your locally compiled provider binary](https://www.terraform.io/docs/commands/cli-config.html#development-overrides-for-provider-developers) instead of the official binary from the Terraform Registry.

For example, add the following to `~/.terraformrc` for a provider binary located in `/home/developer/go/bin`:

```hcl
provider_installation {

  # Use /home/developer/go/bin as an overridden package directory
  # for the hashicorp/azurestack provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # azurestack provider plugin in the given directory.
  dev_overrides {
    "hashicorp/azurestack" = "/home/developer/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

---
