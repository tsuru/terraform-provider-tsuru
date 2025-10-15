resource "tsuru_job" "my-job" {
  name        = "sample-job"
  description = "job created with terraform"
  plan        = "c0.1m0.1"
  team_owner  = "admin"
  pool        = "staging"
  schedule    = "0 0 1 * *"

  container {
    image   = "tsuru/scratch:latest"
    command = ["echo", "hello"]
  }

  metadata {
    labels = {
      "label1" = "value1"
    }
    annotations = {
      "annotation1" = "value1"
      "annotation2" = "value2"
    }
  }
}
