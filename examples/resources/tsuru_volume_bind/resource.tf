resource "tsuru_volume_bind" "volume_bind" {
  volume      = "volume01"
  app         = "sample-app"
  mount_point = "/var/www"
}
