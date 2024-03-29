---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_volume_bind Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Tsuru Service Volume Bind
---

# tsuru_volume_bind (Resource)

Tsuru Service Volume Bind

## Example Usage

```terraform
resource "tsuru_volume_bind" "volume_bind" {
  volume      = "volume01"
  app         = "sample-app"
  mount_point = "/var/www"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) Application name
- `mount_point` (String) Name of service instance
- `volume` (String) Name of service kind

### Optional

- `read_only` (Boolean) restart app after applying (default = false)
- `restart_on_update` (Boolean) restart app after applying (default = true)
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
terraform import tsuru_volume_bind.resource_name "app::volume::mount_point"

# example
terraform import tsuru_volume_bind.volume_bind "sample-app::volume01::/var/www"
```
