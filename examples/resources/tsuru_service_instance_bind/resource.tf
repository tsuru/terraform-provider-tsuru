resource "tsuru_service_instance_bind" "instance_bind" {
  service_name      = "service01"
  service_instance  = "my-instance"
  app               = "app01"
  restart_on_update = true
}
