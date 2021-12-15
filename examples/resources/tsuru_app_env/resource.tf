resource "tsuru_app_env" "env" {
  app               = tsuru_app.my-app.name
  restart_on_update = true

  environment_variables = {
    "ENV1" = "10"
    "ENV2" = "other value"
  }

  private_environment_variables = {
    "SECRET_ENV"  = data.google_secret_manager_secret_version.mysecret.secret_data
    "SECRET_ENV2" = data.myother-secret-manager.mysecret.value
  }
}
