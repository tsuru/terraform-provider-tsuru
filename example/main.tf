terraform {
  required_providers {
    tsuru = {
      source = "tsuru/tsuru"
      version = "0.1.9"
    }
  }
}

provider "tsuru" {
//  host = "https://tsuru.mycompany.com"
  host = "https://tsuru.globoi.com"
}

resource "tsuru_app" "my-app" {
  name = "sample-app-test-v2"

  description = "app created with terraform"

  plan = "small"
  pool = "dev"
  platform = "python"

  team_owner = "devops"

  tags = [ "a", "b" ]
  metadata {
    labels = {
      "label1" = "1"
      "io.tsuru/a" = "2"
    }
    annotations = {
      "annotation" = "value"
      "io.gcp/key" = "{something,2}"
    }
  }
  restart_on_update = true
}

resource "tsuru_app_deploy" "my-app" {
  app = tsuru_app.my-app.name
  image_url = "hello-world:latest"
  message = "hello world"
  override_old_versions = true
}

resource "tsuru_app_grant" "app-permissions-team-a" {
  app = tsuru_app.my-app.name
  team = "team-a"
}

resource "tsuru_app_grant" "app-permissions-team-b" {
  app = tsuru_app.my-app.name
  team = "team-b"
}

resource "tsuru_app_cname" "app-extra-cname" {
  app = tsuru_app.my-app.name
  hostname = "sample.tsuru.i.mycompany.com"
}

//resource "tsuru_app_router" "other-router" {
//  app = tsuru_app.my-app.name
//  name = "internal-http-lb"
//  options = {
//    "key" = "value"
//  }
//}

resource "tsuru_app_autoscale" "web" {
  app = tsuru_app.my-app.name
  process = "web"
  min_units = 3
  max_units = 10
  cpu_average = "800m"
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