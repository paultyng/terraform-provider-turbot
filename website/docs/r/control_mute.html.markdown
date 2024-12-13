---
layout: "turbot"
title: "turbot"
template: Documentation
page_title: "Turbot: turbot_control_mute"
nav:
  title: turbot_control_mute
---

# turbot\_control\_mute

`Turbot Control Mute` allows to mute a control on a given resource.

## Example Usage

**Mute a Control**

```hcl
resource "turbot_control_mute" "mute_control" {
  control_id   = "330102006163524"
  note         = "Muting the control"
  to_timestamp = "2024-12-18T12:54:07.000Z"
  until_states = ["alarm"]
}
```

## Argument Reference

The following arguments are supported:

- `control_id` - (Optional) The `id` or `aka` of the control to mute. **Note:** Either `control_id` or the combination of `control_type` and `resource` must be provided.
- `control_type` - (Optional) The `id` or `aka` of the control type to be muted. Must be used in combination with `resource` if `control_id` is not provided.
- `note` - (Optional) A `note` explaining the reason for muting the control.
- `resource` - (Optional) The `id` or `aka` of the resource where the control is available. Must be used in combination with `control_type` if `control_id` is not provided.
- `to_timestamp` - (Optional) A timestamp, in ISO8601 format specifying when the mute should end.
- `until_states` - (Optional) A list of control states where the mute will not apply.

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `id` - Unique identifier of the control to mute.
- `mute_state` - The mute state of the specified control.

## Import

Control Mute can be imported using the `id`. For example,

```
terraform import turbot_control_mute.test 123456789012
```
