# tsuru_cluster

Manages a Kubernetes cluster in Tsuru.

## Token authentication example

```terraform
resource "tsuru_cluster" "gke" {
  name        = "my-gke-cluster"
  provisioner = "kubernetes"
  
  kube_config {
    cluster {
      server  = "https://k8s.example.com"
      ca_cert = filebase64("~/.kube/ca.crt")
    }
    user {
      token = var.k8s_token
    }
  }

  pools   = ["pool1", "pool2"]
  default = true
}
```

## Certificate authentication example

```terraform
resource "tsuru_cluster" "on_prem" {
  name        = "on-premises-cluster"
  provisioner = "kubernetes"
  crio        = true

  kube_config {
    cluster {
      server = "https://k8s-api.internal"
    }
    user {
      client_certificate = filebase64("path/to/client.crt")
      client_key         = filebase64("path/to/client.key")
    }
  }
}
```

## Arguments

### Required

- `name` - Cluster name (must be unique)
- `provisioner` - Provisioner type (currently only "kubernetes")
- `kube_config` - Kubernetes connection config

### Optional

- `crio` - Use CRI-O as container runtime (default: false)
- `custom_data` - Custom data map
- `default` - Make this the default cluster (default: false)
- `pools` - List of pools to associate with this cluster

### kube_config block

**cluster block:**
- `server` - Kubernetes API server URL
- `ca_cert` - CA certificate (base64)
- `insecure_skip_tls_verify` - Skip TLS verification (not recommended for production)

**user block:**
- `token` - Bearer token for auth
- `client_certificate` - Client cert (base64)
- `client_key` - Client key (base64)

## Import

```bash
terraform import tsuru_cluster.my_cluster "cluster-name"
```

## Common issues

**Connection errors**
Check that the server URL is correct and accessible from where Tsuru runs.

**Authentication failures**
Verify your token/certs are valid and have the right RBAC permissions in Kubernetes.

**Pool association fails**
Make sure the pools exist before associating them. Consider using `tsuru_cluster_pool` resource instead.

## Related resources

- `tsuru_pool`
- `tsuru_cluster_pool`

