---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_volume_bind Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Tsuru Service Volume Bind
---

# tsuru_volume_bind (Resource)

Tsuru Service Volume Bind



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **app** (String) Application name
- **mount_point** (String) Name of service instance
- **volume** (String) Name of service kind

### Optional

- **id** (String) The ID of this resource.
- **read_only** (Boolean) restart app after applying
- **restart_on_update** (Boolean) restart app after applying
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)


