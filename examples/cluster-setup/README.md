# Cluster Setup Example

This example shows how to add a Kubernetes cluster to Tsuru and associate it with a pool.

## Usage

```bash
terraform init
terraform apply
```

After applying, you'll have a new cluster `my-k8s-cluster` connected to the pool `my-pool`.

## Resources used

- `tsuru_cluster` - Kubernetes cluster
- `tsuru_pool` - Resource pool
- `tsuru_cluster_pool` - Cluster-pool association

