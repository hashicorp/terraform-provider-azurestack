---
layout: "azurestack"
page_title: "Azure Resource Manager: The Features Block"
description: |-
Azure Resource Manager: The Features Block

---

# The Features Block

The Azure Stack Provider allows the behaviour of certain resources to be configured using the `features` block.

This allows different users to select the behaviour they require, for example some users may wish for the OS Disks for a Virtual Machine to be removed automatically when the Virtual Machine is destroyed - whereas other users may wish for these OS Disks to be detached but not deleted.

## Example Usage

If you wish to use the default behaviours of the Azure Stack Provider, then you only need to define an empty `features` block as below:

```hcl
provider "azurestack" {
  features {}
}
```

Each of the blocks defined below can be optionally specified to configure the behaviour as needed - this example shows all the possible behaviours which can be configured:

```hcl
provider "azurestack" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = true
    }

    virtual_machine {
      delete_os_disk_on_deletion     = true
      graceful_shutdown              = false
      skip_shutdown_and_force_delete = false
    }

    virtual_machine_scale_set {
      force_delete                  = false
      roll_instances_when_required  = true
      scale_to_zero_before_deletion = true
    }
  }
}
```

## Arguments Reference

The `features` block supports the following:

* `resource_group` - (Optional) A `resource_group` block as defined below.

* `virtual_machine` - (Optional) A `virtual_machine` block as defined below.

* `virtual_machine_scale_set` - (Optional) A `virtual_machine_scale_set` block as defined below.

---

The `resource_group` block supports the following:

* `prevent_deletion_if_contains_resources` - (Optional) Should the `azurestack_resource_group` resource check that there are no Resources within the Resource Group during deletion? This means that all Resources within the Resource Group must be deleted prior to deleting the Resource Group. Defaults to `false`.

-> **Note:** This will be defaulted to `true` in the next major version of the Azure Stack Provider (3.0).

---

The `virtual_machine` block supports the following:

* `delete_os_disk_on_deletion` - (Optional) Should the `azurestack_linux_virtual_machine` and `azurestack_windows_virtual_machine` resources delete the OS Disk attached to the Virtual Machine when the Virtual Machine is destroyed? Defaults to `true`.

~> **Note:** This does not affect the older `azurestack_virtual_machine` resource, which has its own flags for managing this within the resource.

* `graceful_shutdown` - (Optional) Should the `azurestack_linux_virtual_machine` and `azurestack_windows_virtual_machine` request a graceful shutdown when the Virtual Machine is destroyed? Defaults to `false`.

~> **Note:** When using a graceful shutdown, Azure gives the Virtual Machine a 5 minutes window in which to complete the shutdown process, at which point the machine will be force powered off - [more information can be found in this blog post](https://azure.microsoft.com/en-us/blog/linux-and-graceful-shutdowns-2/).

* `skip_shutdown_and_force_delete` - Should the `azurestack_linux_virtual_machine` and `azurestack_windows_virtual_machine` skip the shutdown command and `Force Delete`, this provides the ability to forcefully and immediately delete the VM and detach all sub-resources associated with the virtual machine. This allows those freed resources to be reattached to another VM instance or deleted. Defaults to `false`.

~> **Note:** Support for Force Delete is in an opt-in Preview.

---

The `virtual_machine_scale_set` block supports the following:

* `force_delete` - Should the `azurestack_linux_virtual_machine_scale_set` and `azurestack_windows_virtual_machine_scale_set` resources `Force Delete`, this provides the ability to forcefully and immediately delete the VM and detach all sub-resources associated with the virtual machine. This allows those freed resources to be reattached to another VM instance or deleted. Defaults to `false`.

~> **Note:** Support for Force Delete is in an opt-in Preview.

* `roll_instances_when_required` - (Optional) Should the `azurestack_linux_virtual_machine_scale_set` and `azurestack_windows_virtual_machine_scale_set` resources automatically roll the instances in the Scale Set when Required (for example when updating the Sku/Image). Defaults to `true`.

* `scale_to_zero_before_deletion` - (Optional) Should the `azurestack_linux_virtual_machine_scale_set` and `azurestack_windows_virtual_machine_scale_set` resources scale to 0 instances before deleting the resource. Defaults to `true`.
