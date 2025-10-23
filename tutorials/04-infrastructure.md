# Infrastructure and Clusters

Manage your Tsuru infrastructure: clusters, pools, and resource plans.

## Add a Kubernetes cluster

Create `infrastructure.tf`:

```terraform
resource "tsuru_cluster" "prod" {
  name        = "prod-cluster"
  provisioner = "kubernetes"

  kube_config {
    cluster {
      server  = "https://k8s.prod.example.com"
      ca_cert = filebase64("~/.kube/prod-ca.crt")
    }
    user {
      token = var.k8s_token
    }
  }

  default = true
}
```

Use variables for sensitive data like tokens.

## Create pools

```terraform
resource "tsuru_pool" "production" {
  name        = "production"
  provisioner = "kubernetes"
}

resource "tsuru_pool" "staging" {
  name        = "staging"
  provisioner = "kubernetes"
}
```

## Connect clusters to pools

```terraform
resource "tsuru_cluster_pool" "prod_to_prod" {
  cluster = tsuru_cluster.prod.name
  pool    = tsuru_pool.production.name
}

resource "tsuru_cluster_pool" "prod_to_staging" {
  cluster = tsuru_cluster.prod.name
  pool    = tsuru_pool.staging.name
}
```

Now the prod cluster can run apps from both pools.

## Define resource plans

```terraform
resource "tsuru_plan" "small" {
  name   = "small"
  cpu    = "100m"
  memory = "256Mi"
}

resource "tsuru_plan" "medium" {
  name   = "medium"
  cpu    = "500m"
  memory = "1Gi"
}

resource "tsuru_plan" "large" {
  name   = "large"
  cpu    = "2000m"
  memory = "4Gi"
}
```

Apps can now reference these plans by name.

## Apply and verify

```bash
terraform apply
```

Check what was created:

```bash
tsuru cluster list
tsuru pool list
tsuru plan list
```

Managing infrastructure as code makes it easy to replicate environments and keep everything in version control.

