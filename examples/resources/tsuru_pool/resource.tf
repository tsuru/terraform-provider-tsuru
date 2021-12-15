resource "tsuru_pool" "my-pool" {
  name = "my-pool"
  labels = {
    "my-label" = "value"
  }
}
