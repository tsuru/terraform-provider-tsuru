---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_app_router Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Tsuru Application Router
---

# tsuru_app_router (Resource)

Tsuru Application Router

## Example Usage

```terraform
resource "tsuru_app_router" "other-router" {
  app  = tsuru_app.my-app.name
  name = "my-router"

  options = {
    "key" = "value"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) Application name
- `name` (String) Router name

### Optional

- `options` (Map of String) Application description
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import tsuru_app_router.resource_name "app::router_name"

# example
terraform import tsuru_app_grant.other-router "sample-app::my-router"
```
