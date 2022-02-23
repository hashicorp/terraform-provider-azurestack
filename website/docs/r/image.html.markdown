---
subcategory: "Compute"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_image"
description: |-
  Manages a custom virtual machine image that can be used to create virtual machines.
---

# azurestack_image

Manages a custom virtual machine image that can be used to create virtual machines.

## Example Usage Creating from VHD

```hcl
resource "azurestack_resource_group" "example" {
  name     = "example-resources"
  location = "local"
}

resource "azurestack_image" "example" {
  name                = "acctest"
  location            = "local"
  resource_group_name = azurestack_resource_group.example.name
  zone_resilient      = false

  os_disk {
    os_type  = "Linux"
    os_state = "Generalized"
    blob_uri = "{blob_uri}"
    size_gb  = 30
    caching  = "None"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the image. Changing this forces a
    new resource to be created.
* `resource_group_name` - (Required) The name of the resource group in which to create
    the image. Changing this forces a new resource to be created.
* `location` - (Required) Specified the supported Azure location where the resource exists.
    Changing this forces a new resource to be created.
* `source_virtual_machine_id` - (Optional) The Virtual Machine ID from which to create the image.
* `os_disk` - (Optional) One or more `os_disk` elements as defined below.
* `data_disk` - (Optional) One or more `data_disk` elements as defined below.
* `tags` - (Optional) A mapping of tags to assign to the resource.

`os_disk` supports the following:

* `os_type` - (Required) Specifies the type of operating system contained in the virtual machine image. Possible values are: Windows or Linux.
* `os_state` - (Required) Specifies the state of the operating system contained in the blob. Currently, the only value is Generalized.
* `managed_disk_id` - (Optional) Specifies the ID of the managed disk resource that you want to use to create the image.
* `blob_uri` - (Optional) Specifies the URI in Azure storage of the blob that you want to use to create the image.
* `caching` - (Optional) Specifies the caching mode as `ReadWrite`, `ReadOnly`, or `None`. The default is `None`.
* `size_gb` - (Optional) Specifies the size of the image to be created. The target size can't be smaller than the source size.

`data_disk` supports the following:

* `lun` - (Required) Specifies the logical unit number of the data disk.
* `managed_disk_id` - (Optional) Specifies the ID of the managed disk resource that you want to use to create the image.
* `blob_uri` - (Optional) Specifies the URI in Azure storage of the blob that you want to use to create the image.
* `caching` - (Optional) Specifies the caching mode as `ReadWrite`, `ReadOnly`, or `None`. The default is `None`.
* `size_gb` - (Optional) Specifies the size of the image to be created. The target size can't be smaller than the source size.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Image.

## Timeouts



The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 90 minutes) Used when creating the Image.
* `update` - (Defaults to 90 minutes) Used when updating the Image.
* `read` - (Defaults to 5 minutes) Used when retrieving the Image.
* `delete` - (Defaults to 90 minutes) Used when deleting the Image.

## Import

Images can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_image.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/microsoft.compute/images/image1
```
