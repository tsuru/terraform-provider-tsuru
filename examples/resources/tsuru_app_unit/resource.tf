resource "tsuru_app_unit" "other-unit" {
  app         = tsuru_app.my-app.name
  process     = "web"
  units_count = 20
}

resource "tsuru_app_unit" "other-unit-with-version" {
  app         = tsuru_app.my-app.name
  process     = "worker"
  version     = 2
  units_count = 10
}
