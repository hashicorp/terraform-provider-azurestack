---
subcategory: "Authorization"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_role_assignment"
description: |-
  Assigns a given Principal (User or Group) to a given Role.

---

# azurestack_role_assignment

Assigns a given Principal (User or Group) to a given Role.

## Example Usage (using a built-in Role)

```hcl
data "azurestack_subscription" "primary" {
}

data "azurestack_client_config" "example" {
}

resource "azurestack_role_assignment" "example" {
  scope                = data.azurestack_subscription.primary.id
  role_definition_name = "Reader"
  principal_id         = data.azurestack_client_config.example.object_id
}
```

## Example Usage (Custom Role & Service Principal)

```hcl
data "azurestack_subscription" "primary" {
}

data "azurestack_client_config" "example" {
}

resource "azurestack_role_definition" "example" {
  role_definition_id = "00000000-0000-0000-0000-000000000000"
  name               = "my-custom-role-definition"
  scope              = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["Microsoft.Resources/subscriptions/resourceGroups/read"]
    not_actions = []
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}

resource "azurestack_role_assignment" "example" {
  name               = "00000000-0000-0000-0000-000000000000"
  scope              = data.azurestack_subscription.primary.id
  role_definition_id = azurestack_role_definition.example.role_definition_resource_id
  principal_id       = data.azurestack_client_config.example.object_id
}
```

## Example Usage (Custom Role & User)

```hcl
data "azurestack_subscription" "primary" {
}

data "azurestack_client_config" "example" {
}

resource "azurestack_role_definition" "example" {
  role_definition_id = "00000000-0000-0000-0000-000000000000"
  name               = "my-custom-role-definition"
  scope              = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["Microsoft.Resources/subscriptions/resourceGroups/read"]
    not_actions = []
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}

resource "azurestack_role_assignment" "example" {
  name               = "00000000-0000-0000-0000-000000000000"
  scope              = data.azurestack_subscription.primary.id
  role_definition_id = azurestack_role_definition.example.role_definition_resource_id
  principal_id       = data.azurestack_client_config.example.object_id
}
```

## Example Usage (Custom Role & Management Group)

```hcl
data "azurestack_subscription" "primary" {
}

data "azurestack_client_config" "example" {
}

data "azurestack_management_group" "example" {
}

resource "azurestack_role_definition" "example" {
  role_definition_id = "00000000-0000-0000-0000-000000000000"
  name               = "my-custom-role-definition"
  scope              = data.azurestack_subscription.primary.id

  permissions {
    actions     = ["Microsoft.Resources/subscriptions/resourceGroups/read"]
    not_actions = []
  }

  assignable_scopes = [
    data.azurestack_subscription.primary.id,
  ]
}

resource "azurestack_role_assignment" "example" {
  name               = "00000000-0000-0000-0000-000000000000"
  scope              = data.azurestack_management_group.primary.id
  role_definition_id = azurestack_role_definition.example.role_definition_resource_id
  principal_id       = data.azurestack_client_config.example.object_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A unique UUID/GUID for this Role Assignment - one will be generated if not specified. Changing this forces a new resource to be created.

* `scope` - (Required) The scope at which the Role Assignment applies to, such as `/subscriptions/0b1f6471-1bf0-4dda-aec3-111122223333`, `/subscriptions/0b1f6471-1bf0-4dda-aec3-111122223333/resourceGroups/myGroup`, or `/subscriptions/0b1f6471-1bf0-4dda-aec3-111122223333/resourceGroups/myGroup/providers/Microsoft.Compute/virtualMachines/myVM`, or `/providers/Microsoft.Management/managementGroups/myMG`. Changing this forces a new resource to be created.

* `role_definition_id` - (Optional) The Scoped-ID of the Role Definition. Changing this forces a new resource to be created. Conflicts with `role_definition_name`.

* `role_definition_name` - (Optional) The name of a built-in Role. Changing this forces a new resource to be created. Conflicts with `role_definition_id`.

* `principal_id` - (Required) The ID of the Principal (User, Group or Service Principal) to assign the Role Definition to. Changing this forces a new resource to be created.

~> **NOTE:** The Principal ID is also known as the Object ID (ie not the "Application ID" for applications).

## Attributes Reference

The following attributes are exported:

* `id` - The Role Assignment ID.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Role Assignment.
* `update` - (Defaults to 30 minutes) Used when updating the Role Assignment.
* `read` - (Defaults to 5 minutes) Used when retrieving the Role Assignment.
* `delete` - (Defaults to 30 minutes) Used when deleting the Role Assignment.

## Import

Role Assignments can be imported using the `resource id`, e.g.

```shell
terraform import azurestack_role_assignment.example /subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Authorization/roleAssignments/00000000-0000-0000-0000-000000000000
```

~> **NOTE:** The format of `resource id` could be different for different kinds of `scope`:

- for scope `Subscription`, the id format is `/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Authorization/roleAssignments/00000000-0000-0000-0000-000000000000`
- for scope `Resource Group`, the id format is `/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Authorization/roleAssignments/00000000-0000-0000-0000-000000000000`
