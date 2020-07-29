---
layout: "turbot"
title: "turbot"
template: Documentation
page_title: "Turbot: turbot_control"
nav:
  title: turbot_control
---

# Data Source: turbot\_control

This data source can be used to fetch information about a specific control.

## Example Usage

A simple example to extract the status of a control.

```hcl
data "turbot_control" "test" {
  id      = "112233445566"
}

output "json" {
  value = "${data.turbot_control.test}".state
}
```
Here is another example wherein, we can fetch control data using the control type and target resource.

```hcl
data "turbot_control" "example" {
  type      = "tmod:@turbot/aws-ec2#/control/types/instanceDiscovery"
  resource  = 'arn:aws::ap-northeast-1:112233445566'
}

output "json" {
  value = "${data.turbot_control.example}".state
}
```

## Argument Reference

* `id` - (Optional) The id of the control.
* `type` - (Optional) The type of the control.
* `resource` - (Optional) The unique identifier of the resource which the control is targeting.

**Note:** You must specify either the control id or the control type AND the resource.
## Attributes Reference

* `state` - The state of the control.
* `reason` - Message explaining the state of the control.
* `details` - Additional information regarding the control state.
* `tags` - Tags set on the control.