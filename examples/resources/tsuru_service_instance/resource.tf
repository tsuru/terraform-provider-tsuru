resource "tsuru_service_instance" "my_reverse_proxy" {
  service_name = "rpaasv2"
  name         = "my-reverse-proxy"
  owner        = "my-team"
  description  = "My Reverse Proxy"
  pool         = "some-pool"
  plan         = "c2m2"
  tags         = ["tag_a", "tag_b"]
  parameters = {
    "value"      = "10"
    "otherValue" = "false"
  }
  wait_for_up_status = true
}
