resource "tsuru_app" "my-app" {
  name        = "sample-app"
  description = "app created with terraform"
  plan        = "c0.1m0.2"
  pool        = "staging"
  platform    = "python"
  team_owner  = "admin"
  tags        = ["a", "b"]

  metadata {
    labels = {
      "label1"     = "1"
      "io.tsuru/a" = "2"
    }
    annotations = {
      "annotation" = "value"
      "io.gcp/key" = "something"
    }
  }

  restart_on_update = true
}
