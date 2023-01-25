resource "tsuru_app_deploy" "my-deploy" {
  app   = tsuru_app.my-app.name
  image = "myrepository/my-app:0.1.0"
}
