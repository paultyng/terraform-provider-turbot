---
layout: "turbot"
title: "turbot"
template: Documentation
page_title: "Turbot: turbot_turbot_directory"
nav:
  title: turbot_turbot_directory
---

# turbot\_turbot\_directory

The `Turbot Directory` resource adds support for authentication directories. It is used to create and delete directory settings.

## Example Usage

**Creating Your First Turbot Directory**

```hcl
resource "turbot_turbot_directory" "test" {
	parent              = "tmod:@turbot/turbot#/"
  	title               = "provider_test_refactor"
  	description         = "test directory"
  	profile_id_template = "{{profile.email}}"
  	server              = "test"
	tags = {
		dev = "prod"
	}
}
```

## Argument Reference

The following arguments are supported:

- `parent` - (Required) ID or `aka` of the parent resource.
- `profile_id_template` - (Required) A template to generate profile id for users authenticated through a Turbot Guardrails directory. For example, email id of the user.
- `title` - (Required) Short descriptive name for the directory.
- `server` - (Required)
- `description` - (Optional) Brief description of the purpose and details of the directory.
- `tags` - (Optional) Labels that can be used to manage, group, categorize, search, and save metadata for the directory.

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `parent_akas` - A list of all `akas` for this directory's parent resource.
- `status` - Status of the Turbot Guardrails directory, which defaults to `ACTIVE`. Probable options are `ACTIVE`, `INACTIVE` and `NEW`.
- `id` - Unique identifier of the Turbot Guardrails directory.

## Import

Turbot Guardrails Directories can be imported using the `id`. For example,

```
terraform import turbot_turbot_directory.test 123456789012
```
