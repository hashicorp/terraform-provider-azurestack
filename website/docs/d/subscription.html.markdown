---
subcategory: "Base"
layout: "azurestack"
page_title: "Azure Resource Manager: azurestack_subscription"
description: |-
  Gets information about an existing Subscription.
---

# Data Source: azurestack_subscription

Use this data source to access information about an existing Subscription.

## Example Usage

```hcl
data "azurestack_subscription" "current" {
}

output "current_subscription_display_name" {
  value = data.azurestack_subscription.current.display_name
}
```

## Argument Reference

* `subscription_id` - (Optional) Specifies the ID of the subscription. If this argument is omitted, the subscription ID of the current Azure Resource Manager provider is used.

## Attributes Reference

* `id` - The ID of the subscription.
* `subscription_id` - The subscription GUID.
* `display_name` - The subscription display name.
* `tenant_id` - The subscription tenant ID.
* `state` - The subscription state. Possible values are Enabled, Warned, PastDue, Disabled, and Deleted.
* `location_placement_id` - The subscription location placement ID.
* `tags` - A mapping of tags assigned to the Subscription.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Subscription.
