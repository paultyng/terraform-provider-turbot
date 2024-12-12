---
layout: "turbot"
title: "turbot"
template: Documentation
page_title: "Turbot: turbot_policy_pack"
nav:
  title: turbot_policy_pack
---

# turbot\_policy\_pack

`Turbot Policy Pack` allows resources from across the hierarchy to be organized together as a group.

## Example Usage

**Creating Your First Policy Pack**

```hcl
resource "turbot_policy_pack" "pack" {
  parent      = "tmod:@turbot/turbot#/"
  title       = "Demo Policy Pack"
  description = "My Demo Policy Pack"
  akas        = ["my-demo-policy-pack"]
}
```

## Argument Reference

The following arguments are supported:

- `title` - (Required) Short display name for the policy pack.
- `akas` - (Optional) Unique identifier of the resource.
- `description` - (Optional) Brief description of the purpose and details of the policy pack.
- `parent` - (Optional) The `id` or `aka` of the level at which the policy pack will be created. Defaults to `tmod:@turbot/turbot#/`. 

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `parent_akas` - A list of all `akas` for this policy pack’s parent resource.
- `id` - Unique identifier of the resource.

## Import

Policy Packs can be imported using the `id`. For example,

```
terraform import turbot_policy_pack.test 123456789012
```
