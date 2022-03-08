---
subcategory: "Authorization"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_role_definition"
description: |-
  Get information about an existing Role Definition.
---

# Data Source: azurestack_role_definition

Use this data source to access information about an existing Role Definition.

## Example Usage

```hcl
data "azurestack_subscription" "primary" {
}

resource "azurestack_role_definition" "custom" {
  role_definition_id = "00000000-0000-0000-0000-000000000000"
  name               = "CustomRoleDef"
  scope              = data.azurestack_subscription.primary.id
  #...
}

data "azurestack_role_definition" "custom" {
  role_definition_id = azurestack_role_definition.custom.role_definition_id
  scope              = data.azurestack_subscription.primary.id # /subscriptions/00000000-0000-0000-0000-000000000000
}

data "azurestack_role_definition" "custom-byname" {
  name  = azurestack_role_definition.custom.name
  scope = data.azurestack_subscription.primary.id
}

data "azurestack_role_definition" "builtin" {
  name = "Contributor"
}

output "custom_role_definition_id" {
  value = data.azurestack_role_definition.custom.id
}

output "contributor_role_definition_id" {
  value = data.azurestack_role_definition.builtin.id
}
```



## Argument Reference

* `name` - (Optional) Specifies the Name of either a built-in or custom Role Definition.

-> You can also use this for built-in roles such as `Contributor`, `Owner`, `Reader` and `Virtual Machine Contributor`

* `role_definition_id` - (Optional) Specifies the ID of the Role Definition as a UUID/GUID.
* `scope` - (Optional) Specifies the Scope at which the Custom Role Definition exists.

~> **Note:** One of `name` or `role_definition_id` must be specified.

## Attributes Reference

* `id` - the ID of the built-in Role Definition.
* `description` - the Description of the built-in Role.
* `type` - the Type of the Role.
* `permissions` - a `permissions` block as documented below.
* `assignable_scopes` - One or more assignable scopes for this Role Definition, such as `/subscriptions/0b1f6471-1bf0-4dda-aec3-111122223333`, `/subscriptions/0b1f6471-1bf0-4dda-aec3-111122223333/resourceGroups/myGroup`, or `/subscriptions/0b1f6471-1bf0-4dda-aec3-111122223333/resourceGroups/myGroup/providers/Microsoft.Compute/virtualMachines/myVM`.

A `permissions` block contains:

* `actions` - a list of actions supported by this role
* `not_actions` - a list of actions which are denied by this role

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Role Definition.
