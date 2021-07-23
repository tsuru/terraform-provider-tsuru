terraform {
  required_providers {
    tsuru = {
      source = "tsuru/tsuru"
    }
  }
}

provider "tsuru" {
  host = "https://tsuru.mycompany.com"
}

resource "tsuru_app" "my-app" {
  name = "sample-app"

  description = "app created with terraform"

  plan     = "c0.1m0.2"
  pool     = "staging"
  platform = "python"

  team_owner = "admin"

  tags = ["a", "b"]
  metadata {
    labels = {
      "label1"     = "1"
      "io.tsuru/a" = "2"
    }
    annotations = {
      "annotation" = "value"
      "io.gcp/key" = "{something,2}"
    }
  }
  restart_on_update = true
}

resource "tsuru_app_grant" "app-permissions-team-a" {
  app  = tsuru_app.my-app.name
  team = "team-a"
}

resource "tsuru_app_grant" "app-permissions-team-b" {
  app  = tsuru_app.my-app.name
  team = "team-b"
}

resource "tsuru_app_cname" "app-extra-cname" {
  app      = tsuru_app.my-app.name
  hostname = "sample.tsuru.i.mycompany.com"
}

resource "tsuru_app_router" "other-router" {
  app  = tsuru_app.my-app.name
  name = "internal-http-lb"
  options = {
    "key" = "value"
  }
}

resource "tsuru_app_autoscale" "web" {
  app         = tsuru_app.my-app.name
  process     = "web"
  min_units   = 3
  max_units   = 10
  cpu_average = "800m"
}

resource "tsuru_app_env" "env" {
  app               = tsuru_app.my-app.name
  restart_on_update = false

  environment_variables = {
    "ENV1" = "10"
  }

  private_environment_variables = {
    "ENV2" = "12"
    "ENV3" = "12"
  }
}
