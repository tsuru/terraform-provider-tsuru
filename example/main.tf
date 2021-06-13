terraform {
  required_providers {
    tsuru = {
      source = "tsuru/tsuru"
    }
  }
}

resource "tsuru_app" "my-app" {
  name = var.tsuru_app["name"]

  description = var.tsuru_app["description"]

  plan = var.tsuru_app["plan"]
  pool = var.tsuru_app["pool"]
  platform = var.tsuru_app["platform"]

  team_owner = var.tsuru_app["team_owner"]

  tags = [ "a", "b" ]
  metadata {
    labels = {
      "label1" = var.labels["label1"]
      "io.tsuru/a" = var.labels["io.tsuru/a"]
    }
    annotations = {
      "annotation" = var.annotations["annotation"]
      "io.gcp/key" = var.annotations["io.gcp/key"]
    }
  }
  restart_on_update = true
}

resource "tsuru_app_grant" "app-permissions-team-a" {
  app = tsuru_app.my-app.name
  count = length(var.tsuru_app_grant)
  team = var.tsuru_app_grant[count.index]
}

resource "tsuru_app_cname" "app-extra-cname" {
  app = tsuru_app.my-app.name
  hostname = var.tsuru_app_cname
}

resource "tsuru_app_router" "other-router" {
  app = tsuru_app.my-app.name
  name = var.tsuru_app_router["name"]
  options = {
    "key" = var.tsuru_app_router["key"]
  }
}

resource "tsuru_app_autoscale" "web" {
  app = tsuru_app.my-app.name
  process = var.tsuru_app_autoscale["process"]
  min_units = var.tsuru_app_autoscale["min_units"]
  max_units = var.tsuru_app_autoscale["max_units"]
  cpu_average = var.tsuru_app_autoscale["cpu_average"]
}

resource "tsuru_app_env" "env" {
  app = tsuru_app.my-app.name
  restart_on_update = false

  environment_variable {
    name = "env1"
    value = "10"
  }

  environment_variable {
    name = "env2"
    sensitive_value = "12"
  }

  environment_variable {
    name = "env3"
    value = "12"
  }
}