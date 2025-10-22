# tsuru_app

Manages a Tsuru application.

## Basic example

```terraform
resource "tsuru_app" "simple" {
  name        = "my-app"
  description = "My application"
  plan        = "c0.1m0.2"
  pool        = "staging"
  platform    = "python"
  team_owner  = "my-team"
}
```

## Advanced example with processes

```terraform
resource "tsuru_app" "api" {
  name        = "web-api"
  description = "Web API with custom processes"
  plan        = "c0.2m0.5"
  pool        = "production"
  platform    = "go"
  team_owner  = "api-team"
  tags        = ["api", "backend"]

  metadata {
    labels = {
      "app.kubernetes.io/name": "web-api"
    }
    annotations = {
      "monitoring.tsuru.io/scrape": "true"
    }
  }

  process {
    name = "web"
    plan = "c0.5m1.0"
  }

  process {
    name = "worker"
    plan = "c0.2m0.5"
  }

  restart_on_update = true
}
```

## Arguments

### Required

- `name` - App name (must be unique)
- `plan` - Resource plan (CPU/memory allocation)
- `platform` - Platform (python, go, nodejs, etc)
- `pool` - Pool where the app runs
- `team_owner` - Team that owns the app

### Optional

- `custom_cpu_burst` - CPU burst factor
- `default_router` - Default router for the app
- `description` - App description
- `metadata` - Labels and annotations
- `process` - Per-process configuration
- `restart_on_update` - Restart app when config changes (default: false)
- `tags` - List of tags

### Read-only

- `cluster` - Cluster name where app is running
- `id` - Resource ID (same as name)
- `internal_address` - Internal addresses
- `router` - Router information

## Import

```bash
terraform import tsuru_app.my_app "app-name"
```

## Common issues

**404 Not Found when importing**
Check the app name and verify you have access to it.

**Plan shows unexpected changes after import**
Your Terraform config doesn't match the real app. Check metadata and process configs carefully.

## Related resources

- `tsuru_app_deploy`
- `tsuru_app_env`
- `tsuru_app_cname`
- `tsuru_app_autoscale`

