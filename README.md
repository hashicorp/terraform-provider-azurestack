Azure Stack Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

Compatibility
------------

The Azure Stack provider is only compatible with specific profile versions as listed below:

| Azure Stack Profile Version | Supported Azure Stack Provider Versions |
| --------------------------- | --------------------------------------- |
| 2019-03-01                  | 0.8+                                    |
| 2017-10-01                  | 0.1-0.7                                 |

You can pin the version of the Azure Stack Provider you're using like so:

```hcl
provider "azurestack" {
  version = "=0.9.0"
}
```

General Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.9 (to build the provider plugin)

Windows Specific Requirements
-----------------------------
- [Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm)
- [Git Bash for Windows](https://git-scm.com/download/win)

For *GNU32 Make*, make sure its bin path is added to PATH environment variable.*

For *Git Bash for Windows*, at the step of "Adjusting your PATH environment", please choose "Use Git and optional Unix tools from Windows Command Prompt".*

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-azurestack`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-azurestack
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-azurestack
$ make build
```

Using the provider
----------------------

```
# These variables can also be set as Environment Variables
# see http://terraform.io/docs/providers/azurestack/index.html for more info
provider "azurestack" {
  # arm_endpoint    = "..."
  # subscription_id = "..."
  # client_id       = "..."
  # client_secret   = "..."
  # tenant_id       = "..."
}

# Create a resource group
resource "azurestack_resource_group" "production" {
  name     = "production"
  location = "West US"
}

# Create a virtual network in the web_servers resource group
resource "azurestack_virtual_network" "network" {
  name                = "productionNetwork"
  address_space       = ["10.0.0.0/16"]
  location            = "West US"
  resource_group_name = azurestack_resource_group.production.name

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

Further [usage documentation is available on the Terraform website](https://www.terraform.io/docs/providers/azurestack/index.html).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.9+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-azurestack
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

The following ENV variables must be set in your shell prior to running acceptance tests:

- ARM_ENDPOINT
- ARM_CLIENT_ID
- ARM_CLIENT_SECRET
- ARM_SUBSCRIPTION_ID
- ARM_TENANT_ID
- ARM_TEST_LOCATION

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
