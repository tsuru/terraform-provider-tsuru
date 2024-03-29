---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_volume Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Tsuru Service Volume
---

# tsuru_volume (Resource)

Tsuru Service Volume

## Example Usage

```terraform
resource "tsuru_volume" "volume" {
  name  = "volume01"
  owner = "my-team"
  plan  = "plan01"
  pool  = "pool01"
  options = {
    "key" = "value"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Volume name
- `owner` (String) Team owner of this volume
- `plan` (String)
- `pool` (String) Volume Pool

### Optional

- `options` (Map of String) Volume additional options
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import tsuru_volume.resource_name "name"

# example
terraform import tsuru_volume.volume "volume01"
```
