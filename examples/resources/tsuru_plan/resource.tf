resource "tsuru_plan" "plan1" {
  name    = "plan1"
  cpu     = "1" // or "100m" or "200%"
  memory  = "1Gi"
  default = true
}
