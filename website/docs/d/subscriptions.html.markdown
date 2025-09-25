---
subcategory: "Base"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_subscriptions"
description: |-
  Get information about the available subscriptions.
---

# Data Source: azurestack_subscriptions

Use this data source to access information about all the Subscriptions currently available.

## Example Usage

```hcl
data "azurestack_subscriptions" "available" {
}

output "available_subscriptions" {
  value = data.azurestack_subscriptions.available.subscriptions
}

output "first_available_subscription_display_name" {
  value = data.azurestack_subscriptions.available.subscriptions[0].display_name
}
```

## Argument Reference

* `display_name_prefix` - (Optional) A case-insensitive prefix which can be used to filter on the `display_name` field
* `display_name_contains` - (Optional) A case-insensitive value which must be contained within the `display_name` field, used to filter the results

## Attributes Reference

* `subscriptions` - One or more `subscription` blocks as defined below.

The `subscription` block contains:

* `id` - The ID of this subscription.
* `subscription_id` - The subscription GUID.
* `display_name` - The subscription display name.
* `tenant_id` - The subscription tenant ID.
* `state` - The subscription state. Possible values are Enabled, Warned, PastDue, Disabled, and Deleted.
* `location_placement_id` - The subscription location placement ID.
* `tags` - A mapping of tags assigned to the resource.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the subscriptions.
