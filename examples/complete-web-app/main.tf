resource "tsuru_app" "web_app" {
  name        = "complete-web-app"
  description = "Complete web application example"
  plan        = "c0.1m0.2"
  pool        = "staging"
  platform    = "python"
  team_owner  = "admin"
  tags        = ["terraform", "example", "web"]

  metadata {
    labels = {
      "app"       = "web-app"
      "component" = "frontend"
    }
    annotations = {
      "owner" = "team-alpha"
    }
  }

  restart_on_update = true
}

resource "tsuru_app_env" "app_env" {
  app = tsuru_app.web_app.name

  environment_variables = {
    "DATABASE_URL" = "postgres://user:password@host:port/dbname"
    "CACHE_URL"    = "redis://host:port"
  }

  no_restart = false
}

resource "tsuru_app_cname" "app_cname" {
  app   = tsuru_app.web_app.name
  cname = "www.complete-web-app.com"
}

resource "tsuru_app_autoscale" "app_autoscale" {
  app = tsuru_app.web_app.name

  min_units = 2
  max_units = 5

  cpu {
    target = "80%"
  }

  schedules = [
    {
      start        = "0 8 * * 1-5"
      end          = "0 20 * * 1-5"
      min_replicas = 3
    }
  ]
}

