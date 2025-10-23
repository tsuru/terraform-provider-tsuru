# Managing Applications

This tutorial covers environment variables, custom domains, and autoscaling.

## Environment variables

Add environment variables to your app:

```terraform
resource "tsuru_app_env" "app_config" {
  app = tsuru_app.first_app.name

  environment_variables = {
    "DATABASE_URL" = "postgres://user:pass@host:port/db"
    "DEBUG"        = "false"
  }
}
```

For secrets, use a secrets manager instead of hardcoding values.

## Custom domain

Add a CNAME to your app:

```terraform
resource "tsuru_app_cname" "app_domain" {
  app   = tsuru_app.first_app.name
  cname = "app.example.com"
}
```

Remember to configure your DNS to point to Tsuru's router.

## Autoscaling

Scale based on CPU and schedule:

```terraform
resource "tsuru_app_autoscale" "app_scaling" {
  app = tsuru_app.first_app.name

  min_units = 2
  max_units = 10

  cpu {
    target = "70%"
  }

  schedules = [
    {
      start        = "0 9 * * 1-5"
      end          = "0 18 * * 1-5"
      min_replicas = 3
      timezone     = "America/Sao_Paulo"
    }
  ]
}
```

This keeps 2-10 units, scales at 70% CPU, and guarantees 3 units during weekday business hours.

## Complete example

Your `main.tf` should now look like this:

```terraform
terraform {
  required_providers {
    tsuru = {
      source  = "tsuru/tsuru"
      version = "~> 2.17.0"
    }
  }
}

provider "tsuru" {}

resource "tsuru_app" "first_app" {
  name        = "my-first-app"
  description = "My first Terraform-managed app"
  platform    = "python"
  team_owner  = "my-team"
  pool        = "my-pool"
  plan        = "small"
}

resource "tsuru_app_env" "app_config" {
  app = tsuru_app.first_app.name

  environment_variables = {
    "DATABASE_URL" = "postgres://user:pass@host:port/db"
    "DEBUG"        = "false"
  }
}

resource "tsuru_app_cname" "app_domain" {
  app   = tsuru_app.first_app.name
  cname = "app.example.com"
}

resource "tsuru_app_autoscale" "app_scaling" {
  app = tsuru_app.first_app.name

  min_units = 2
  max_units = 10

  cpu {
    target = "70%"
  }

  schedules = [
    {
      start        = "0 9 * * 1-5"
      end          = "0 18 * * 1-5"
      min_replicas = 3
      timezone     = "America/Sao_Paulo"
    }
  ]
}
```

Run `terraform apply` to update your app with these new configurations.

