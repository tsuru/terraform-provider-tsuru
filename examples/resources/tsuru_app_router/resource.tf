resource "tsuru_app_router" "other-router" {
  app  = tsuru_app.my-app.name
  name = "my-router"

  options = {
    "key" = "value"
  }
}
