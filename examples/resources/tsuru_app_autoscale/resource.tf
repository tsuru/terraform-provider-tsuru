resource "tsuru_app_autoscale" "web" {
  app         = tsuru_app.my-app.name
  process     = "web"
  min_units   = 3
  max_units   = 10
  cpu_average = "60%"
}
