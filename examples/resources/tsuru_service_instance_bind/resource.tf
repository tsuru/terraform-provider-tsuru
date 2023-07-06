resource "tsuru_service_instance_bind" "app_bind" {
  service_name      = "service01"
  service_instance  = "my-instance"
  app               = "app01"
  restart_on_update = true
}

resource "tsuru_service_instance_bind" "job_bind" {
  service_name      = "service01"
  service_instance  = "my-instance"
  job               = "job01"
  restart_on_update = true
}