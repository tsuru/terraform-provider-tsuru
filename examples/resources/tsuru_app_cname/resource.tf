resource "tsuru_app_cname" "app-extra-cname" {
  app      = tsuru_app.my-app.name
  hostname = "mydomain.com"
}
