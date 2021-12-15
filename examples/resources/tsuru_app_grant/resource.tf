resource "tsuru_app_grant" "app-permissions-team-a" {
  app  = tsuru_app.my-app.name
  team = "colab-team"
}
