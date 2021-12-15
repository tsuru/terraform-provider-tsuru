resource "tsuru_service_instance_grant" "instance_grant" {
  service_name     = "service01"
  service_instance = "my-instance"
  team             = "mysupport-team"
}
