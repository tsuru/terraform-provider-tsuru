resource "tsuru_app_unit" "other-unit" {
  app         = tsuru_app.my-app.name
  process     = "web"
  units_count = 20
}
