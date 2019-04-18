---
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_public_ip"
sidebar_current: "docs-azurestack-datasource-public-ip-x"
description: |-
  Retrieves information about the specified public IP address.

---

# Data Source: azurestack_public_ip

Use this data source to access the properties of an existing Azure Public IP Address.

## Example Usage (reference an existing)

```hcl
data "azurestack_public_ip" "test" {
  name                = "name_of_public_ip"
  resource_group_name = "name_of_resource_group"
}

output "domain_name_label" {
  value = "${data.azurestack_public_ip.test.domain_name_label}"
}

output "public_ip_address" {
  value = "${data.azurestack_public_ip.test.ip_address}"
}
```

## Example Usage (Retrieve the Dynamic Public IP of a new VM)

```hcl
resource "azurestack_resource_group" "test" {
  name     = "test-resources"
  location = "West US 2"
}

resource "azurestack_virtual_network" "test" {
  name                = "test-network"
  address_space       = ["10.0.0.0/16"]
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"
}

resource "azurestack_subnet" "test" {
  name                 = "acctsub"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  virtual_network_name = "${azurestack_virtual_network.test.name}"
  address_prefix       = "10.0.2.0/24"
}

resource "azurestack_public_ip" "test" {
  name                         = "test-pip"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "Dynamic"
  idle_timeout_in_minutes      = 30

  tags = {
    environment = "test"
  }
}

resource "azurestack_network_interface" "test" {
  name                = "test-nic"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "static"
    private_ip_address            = "10.0.2.5"
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "test-vm"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]

  # ...
}

data "azurestack_public_ip" "test" {
  name                = "${azurestack_public_ip.test.name}"
  resource_group_name = "${azurestack_virtual_machine.test.resource_group_name}"
}

output "public_ip_address" {
  value = "${data.azurestack_public_ip.test.ip_address}"
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the public IP address.
* `resource_group_name` - (Required) Specifies the name of the resource group.


## Attributes Reference

* `domain_name_label` - The label for the Domain Name.
* `idle_timeout_in_minutes` - Specifies the timeout for the TCP idle connection.
* `fqdn` - Fully qualified domain name of the A DNS record associated with the public IP. This is the concatenation of the domainNameLabel and the regionalized DNS zone.
* `ip_address` - The IP address value that was allocated.
* `tags` - A mapping of tags to assigned to the resource.
