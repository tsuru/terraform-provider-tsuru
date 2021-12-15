resource "tsuru_router" "test_router" {
  name            = "test_router"
  type            = "router"
  readiness_gates = ["gate1", "gate2"]
  config          = <<-EOT
    url: "testing"
    headers:
      "x-my-header": test
    EOT
}
