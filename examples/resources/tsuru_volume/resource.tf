resource "tsuru_volume" "volume" {
  name  = "volume01"
  owner = "my-team"
  plan  = "plan01"
  pool  = "pool01"
  options = {
    "key" = "value"
  }
}
