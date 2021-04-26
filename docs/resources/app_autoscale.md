---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_app_autoscale Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Tsuru Application Autoscale
---

# tsuru_app_autoscale (Resource)

Tsuru Application Autoscale



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **app** (String) Application name
- **cpu_average** (String) cpu average
- **max_units** (Number) maximum number of units
- **process** (String) Name of service instance

### Optional

- **id** (String) The ID of this resource.
- **min_units** (Number) minimum number of units
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)

