---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tsuru_cluster_pool Resource - terraform-provider-tsuru"
subcategory: ""
description: |-
  Resource used to assign pools into clusters
---

# tsuru_cluster_pool (Resource)

Resource used to assign pools into clusters

## Example Usage

```terraform
resource "tsuru_cluster_pool" "cluster-pool" {
  cluster = "my-pool"
  pool    = "my-cluster"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster` (String) The name of defined cluster
- `pool` (String) The name of defined pool

### Optional

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
terraform import tsuru_cluster_pool.resource_name "cluster/pool"

# example
terraform import tsuru_cluster_pool.cluster-pool "my-cluster/my-pool"
```
