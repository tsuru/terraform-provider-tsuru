---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_app Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Tsuru Application
---

# tsuru_app (Resource)

Tsuru Application



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) Application name
- **plan** (String) Plan
- **platform** (String) Platform
- **pool** (String) The name of pool
- **team_owner** (String) Application owner

### Optional

- **default_router** (String) Default router at creation of app
- **description** (String) Application description
- **id** (String) The ID of this resource.
- **metadata** (Block List, Max: 1) (see [below for nested schema](#nestedblock--metadata))
- **restart_on_update** (Boolean) Restart app after applying changes
- **tags** (List of String) Tags

### Read-Only

- **internal_address** (List of Object) (see [below for nested schema](#nestedatt--internal_address))
- **router** (List of Object) (see [below for nested schema](#nestedatt--router))

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

Optional:

- **annotations** (Map of String)
- **labels** (Map of String)


<a id="nestedatt--internal_address"></a>
### Nested Schema for `internal_address`

Read-Only:

- **domain** (String)
- **port** (Number)
- **process** (String)
- **protocol** (String)
- **version** (String)


<a id="nestedatt--router"></a>
### Nested Schema for `router`

Read-Only:

- **addresses** (List of String)
- **name** (String)
- **options** (Map of String)


