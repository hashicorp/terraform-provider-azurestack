---
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_virtual_machine"
sidebar_current: "docs-azurestack-resource-compute-virtual-machine"
description: |-
  Manages a Virtual Machine.
---

# azurestack_virtual_machine

Manages a virtual machine.

## Example Usage with Managed Disks

```hcl
resource "azurestack_resource_group" "test" {
  name = "acctestrg"

  # This is Azure Stack Region so it will be different per Azure Stack and should not be in the format of "West US" etc... those are not the same values
  location = "region1"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn"
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

resource "azurestack_network_interface" "test" {
  name                = "acctni"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_F2"

  # Uncomment this line to delete the OS disk automatically when deleting the VM
  # delete_os_disk_on_termination = true


  # Uncomment this line to delete the data disks automatically when deleting the VM
  # delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags {
    environment = "staging"
  }
}
```

## Example Usage with Managed Disks and Public IP

```hcl
resource "azurestack_resource_group" "test" {
  name = "acctestrg"

  # This is Azure Stack Region so it will be different per Azure Stack and should not be in the format of "West US" etc... those are not the same values
  location = "region1"
}

resource "azurestack_public_ip" "test" {
  name                         = "acceptanceTestPublicIp1"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"

  tags {
    environment = "Production"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn"
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

resource "azurestack_network_interface" "test" {
  name                = "acctni"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
  }
}


resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D2_v2"

  # Uncomment this line to delete the OS disk automatically when deleting the VM
  # delete_os_disk_on_termination = true


  # Uncomment this line to delete the data disks automatically when deleting the VM
  # delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags {
    environment = "staging"
  }
}
```

## Example Usage with Unmanaged Disks

```hcl
resource "azurestack_resource_group" "test" {
  name = "acctestrg"

  # This is Azure Stack Region so it will be different per Azure Stack and should not be in the format of "West US" etc... those are not the same values
  location = "region1"
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn"
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

resource "azurestack_network_interface" "test" {
  name                = "acctni"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
  }
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_F2"

  # Uncomment this line to delete the OS disk automatically when deleting the VM
  # delete_os_disk_on_termination = true


  # Uncomment this line to delete the data disks automatically when deleting the VM
  # delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
  storage_os_disk {
    name          = "myosdisk1"
    vhd_uri       = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
    caching       = "ReadWrite"
    create_option = "FromImage"
  }
  # Optional data disks
  storage_data_disk {
    name          = "datadisk0"
    vhd_uri       = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/datadisk0.vhd"
    disk_size_gb  = "1023"
    create_option = "Empty"
    lun           = 0
  }
  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
  os_profile_linux_config {
    disable_password_authentication = false
  }
  tags = {
    environment = "staging"
  }
}
```

## Example Usage with Unmanaged Disks and Public IP

```hcl
resource "azurestack_resource_group" "test" {
  name = "acctestrg"

  # This is Azure Stack Region so it will be different per Azure Stack and should not be in the format of "West US" etc... those are not the same values
  location = "region1"
}

resource "azurestack_public_ip" "test" {
  name                         = "acceptanceTestPublicIp1"
  location                     = "${azurestack_resource_group.test.location}"
  resource_group_name          = "${azurestack_resource_group.test.name}"
  public_ip_address_allocation = "static"

  tags = {
    environment = "Production"
  }
}

resource "azurestack_virtual_network" "test" {
  name                = "acctvn"
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

resource "azurestack_network_interface" "test" {
  name                = "acctni"
  location            = "${azurestack_resource_group.test.location}"
  resource_group_name = "${azurestack_resource_group.test.name}"

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = "${azurestack_subnet.test.id}"
    private_ip_address_allocation = "dynamic"
    public_ip_address_id          = "${azurestack_public_ip.test.id}"
  }
}

resource "azurestack_storage_account" "test" {
  name                     = "accsa"
  resource_group_name      = "${azurestack_resource_group.test.name}"
  location                 = "${azurestack_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}

resource "azurestack_storage_container" "test" {
  name                  = "vhds"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  storage_account_name  = "${azurestack_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurestack_virtual_machine" "test" {
  name                  = "acctvm"
  location              = "${azurestack_resource_group.test.location}"
  resource_group_name   = "${azurestack_resource_group.test.name}"
  network_interface_ids = ["${azurestack_network_interface.test.id}"]
  vm_size               = "Standard_D2_v2"

  # Uncomment this line to delete the OS disk automatically when deleting the VM
  # delete_os_disk_on_termination = true


  # Uncomment this line to delete the data disks automatically when deleting the VM
  # delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
  storage_os_disk {
    name          = "myosdisk1"
    vhd_uri       = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/myosdisk1.vhd"
    caching       = "ReadWrite"
    create_option = "FromImage"
  }
  # Optional data disks
  storage_data_disk {
    name          = "datadisk0"
    vhd_uri       = "${azurestack_storage_account.test.primary_blob_endpoint}${azurestack_storage_container.test.name}/datadisk0.vhd"
    disk_size_gb  = "1023"
    create_option = "Empty"
    lun           = 0
  }
  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
  os_profile_linux_config {
    disable_password_authentication = false
  }
  tags = {
    environment = "staging"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the virtual machine resource. Changing this forces a
    new resource to be created.
* `resource_group_name` - (Required) The name of the resource group in which to
    create the virtual machine.
* `location` - (Required) Specifies the supported Azure Stack Region where the resource exists. Changing this forces a new resource to be created.
* `plan` - (Optional) A plan block as documented below.
* `availability_set_id` - (Optional) The Id of the Availability Set in which to create the virtual machine
* `boot_diagnostics` - (Optional) A boot diagnostics profile block as referenced below.
* `vm_size` - (Required) Specifies the [size of the virtual machine](https://azure.microsoft.com/en-us/documentation/articles/virtual-machines-size-specs/).
* `storage_image_reference` - (Optional) A Storage Image Reference block as documented below.
* `storage_os_disk` - (Required) A `storage_os_disk` block.
* `storage_data_disk` - (Optional) A list of Storage Data disk blocks as referenced below.
* `delete_os_disk_on_termination` - (Optional) Should the OS Disk be deleted when the Virtual Machine is destroyed? Defaults to `false`.
* `delete_data_disks_on_termination` - (Optional) Flag to enable deletion of storage data disk VHD blobs when the VM is deleted, defaults to `false`.
* `os_profile` - (Optional) An OS Profile block as documented below. Required when `create_option` in the `storage_os_disk` block is set to `FromImage`.
* `identity` - (Optional) An identity block as documented below.

* `license_type` - (Optional, when a Windows machine) Specifies the Windows OS license type. If supplied, the only allowed values are `Windows_Client` and `Windows_Server`.
* `os_profile_windows_config` - (Required, when a Windows machine) A Windows config block as documented below.
* `os_profile_linux_config` - (Required, when a Linux machine) A Linux config block as documented below.
* `os_profile_secrets` - (Optional) A collection of Secret blocks as documented below.
* `network_interface_ids` - (Required) Specifies the list of resource IDs for the network interfaces associated with the virtual machine.
* `primary_network_interface_id` - (Optional) Specifies the resource ID for the primary network interface associated with the virtual machine.
* `tags` - (Optional) A mapping of tags to assign to the resource.

For more information on the different example configurations, please check out the [azure documentation](https://msdn.microsoft.com/en-us/library/mt163591.aspx#Anchor_2)

`Plan` supports the following:

* `name` - (Required) Specifies the name of the image from the marketplace.
* `publisher` - (Required) Specifies the publisher of the image.
* `product` - (Required) Specifies the product of the image from the marketplace.

`boot_diagnostics` supports the following:

* `enabled`: (Required) Whether to enable boot diagnostics for the virtual machine.
* `storage_uri`: (Required) Blob endpoint for the storage account to hold the virtual machine's diagnostic files. This must be the root of a storage account, and not a storage container.

`storage_image_reference` supports the following:

* `id` - (Optional) Specifies the ID of the (custom) image to use to create the virtual
machine, for example:

```hcl

resource "azurestack_image" "test" {
	name = "test"
  ...
}

resource "azurestack_virtual_machine" "test" {
	name = "test"
  ...

	storage_image_reference {
		id = "${azurestack_image.test.id}"
	}

...
```

* `publisher` - (Required, when not using image resource) Specifies the publisher of the image used to create the virtual machine. Changing this forces a new resource to be created.
* `offer` - (Required, when not using image resource) Specifies the offer of the image used to create the virtual machine. Changing this forces a new resource to be created.
* `sku` - (Required, when not using image resource) Specifies the SKU of the image used to create the virtual machine. Changing this forces a new resource to be created.
* `version` - (Optional) Specifies the version of the image used to create the virtual machine. Changing this forces a new resource to be created.

`storage_os_disk` block supports the following:

* `name` - (Required) Specifies the disk name.
* `create_option` - (Required) Specifies how the OS Disk should be created. Possible values are `Attach` (managed disks only) and `FromImage`.
* `caching` - (Optional) Specifies the caching requirements for the OS Disk. Possible values include `None`, `ReadOnly` and `ReadWrite`.
* `image_uri` - (Optional) Specifies the image_uri in the form publisherName:offer:skus:version. `image_uri` can also specify the [VHD uri](https://azure.microsoft.com/en-us/documentation/articles/virtual-machines-linux-cli-deploy-templates/#create-a-custom-vm-image) of a custom VM image to clone. When cloning a custom disk image the `os_type` documented below becomes required.
* `os_type` - (Optional) Specifies the Operating System on the OS Disk. Possible values are `Linux` and `Windows`.
* `disk_size_gb` - (Optional) Specifies the size of the os disk in gigabytes.

The following properties apply when using Managed Disks:

* `managed_disk_id` - (Optional) Specifies the ID of an existing Managed Disk which should be attached as the OS Disk of this Virtual Machine. If this is set then the `create_option` must be set to `Attach`.

* `managed_disk_type` - (Optional) Specifies the type of Managed Disk which should be created. Possible values are `Standard_LRS` or `Premium_LRS`.

The following properties apply when using Unmanaged Disks:

* `vhd_uri` - (Optional) Specifies the URI of the VHD file backing this Unmanaged OS Disk. Changing this forces a new resource to be created.

`storage_data_disk` supports the following:

* `name` - (Required) Specifies the name of the data disk.
* `create_option` - (Required) Specifies how the data disk should be created. Possible values are `Attach`, `FromImage` and `Empty`.
* `disk_size_gb` - (Required) Specifies the size of the data disk in gigabytes.
* `caching` - (Optional) Specifies the caching requirements.
* `lun` - (Required) Specifies the logical unit number of the data disk.

The following properties apply when using Managed Disks:

* `managed_disk_type` - (Optional) Specifies the type of managed disk to create. Possible values are either `Standard_LRS` or `Premium_LRS`.

* `managed_disk_id` - (Optional) Specifies the ID of an Existing Managed Disk which should be attached to this Virtual Machine. When this field is set `create_option` must be set to `Attach`.

The following properties apply when using Unmanaged Disks:

* `vhd_uri` - (Optional) Specifies the URI of the VHD file backing this Unmanaged Data Disk. Changing this forces a new resource to be created.

`os_profile` supports the following:

* `computer_name` - (Required) Specifies the name of the virtual machine.
* `admin_username` - (Required) Specifies the name of the administrator account.
* `admin_password` - (Required for Windows, Optional for Linux) Specifies the password of the administrator account.
* `custom_data` - (Optional) Specifies custom data to supply to the machine. On linux-based systems, this can be used as a cloud-init script. On other systems, this will be copied as a file on disk. Internally, Terraform will base64 encode this value before sending it to the API. The maximum length of the binary array is 65535 bytes.

~> **NOTE:** `admin_password` must be between 6-72 characters long and must satisfy at least 3 of password complexity requirements from the following:
1. Contains an uppercase character
2. Contains a lowercase character
3. Contains a numeric digit
4. Contains a special character

`identity` supports the following:

* `type` - (Required) Specifies the identity type of the virtual machine. The only allowable value is `SystemAssigned`. To enable Managed Service Identity the virtual machine extension "ManagedIdentityExtensionForWindows" or "ManagedIdentityExtensionForLinux" must also be added to the virtual machine. The Principal ID can be retrieved after the virtual machine has been created, e.g.

```hcl
resource "azurestack_virtual_machine" "test" {
  name = "test"

  identity = {
    type = "SystemAssigned"
  }
}

resource "azurestack_virtual_machine_extension" "test" {
  name                 = "test"
  resource_group_name  = "${azurestack_resource_group.test.name}"
  location             = "${azurestack_resource_group.test.location}"
  virtual_machine_name = "${azurestack_virtual_machine.test.name}"
  publisher            = "Microsoft.ManagedIdentity"
  type                 = "ManagedIdentityExtensionForWindows"
  type_handler_version = "1.0"

  settings = <<SETTINGS
    {
        "port": 50342
    }
SETTINGS
}

output "principal_id" {
  value = "${lookup(azurestack_virtual_machine.test.identity[0], "principal_id")}"
}
```

`os_profile_windows_config` supports the following:

* `provision_vm_agent` - (Optional) This value defaults to false.
* `enable_automatic_upgrades` - (Optional) This value defaults to false.
* `winrm` - (Optional) A collection of WinRM configuration blocks as documented below.
* `additional_unattend_config` - (Optional) An Additional Unattended Config block as documented below.

`winrm` supports the following:

* `protocol` - (Required) Specifies the protocol of listener
* `certificate_url` - (Optional) Specifies URL of the certificate with which new Virtual Machines is provisioned.

`additional_unattend_config` supports the following:

* `pass` - (Required) Specifies the name of the pass that the content applies to. The only allowable value is `oobeSystem`.
* `component` - (Required) Specifies the name of the component to configure with the added content. The only allowable value is `Microsoft-Windows-Shell-Setup`.
* `setting_name` - (Required) Specifies the name of the setting to which the content applies. Possible values are: `FirstLogonCommands` and `AutoLogon`.
* `content` - (Optional) Specifies the base-64 encoded XML formatted content that is added to the unattend.xml file for the specified path and component.

`os_profile_linux_config` supports the following:

* `disable_password_authentication` - (Required) Specifies whether password authentication should be disabled. If set to `false`, an `admin_password` must be specified.
* `ssh_keys` - (Optional) Specifies a collection of `path` and `key_data` to be placed on the virtual machine.

~> **Note:** Please note that the only allowed `path` is `/home/<username>/.ssh/authorized_keys` due to a limitation of Azure.

`os_profile_secrets` supports the following:

* `source_vault_id` - (Required) Specifies the key vault to use.
* `vault_certificates` - (Required) A collection of Vault Certificates as documented below

`vault_certificates` support the following:

* `certificate_url` - (Required) Specifies the URI of the key vault secrets in the format of `https://<vaultEndpoint>/secrets/<secretName>/<secretVersion>`. Stored secret is the Base64 encoding of a JSON Object that which is encoded in UTF-8 of which the contents need to be

```json
{
  "data":"<Base64-encoded-certificate>",
  "dataType":"pfx",
  "password":"<pfx-file-password>"
}
```

* `certificate_store` - (Required, on windows machines) Specifies the certificate store on the Virtual Machine where the certificate should be added to.

## Attributes Reference

The following attributes are exported:

* `id` - The virtual machine ID.

## Import

Virtual Machines can be imported using the `resource id`, e.g.

```hcl
terraform import azurestack_virtual_machine.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/microsoft.compute/virtualMachines/machine1
```
