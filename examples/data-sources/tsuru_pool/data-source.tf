data "tsuru_pool" "my-pool" {
  name = "my-pool"
}

output "my-pool-plans" {
  value = one([for c in data.tsuru_pool.my-pool.allowed : c.values if c.field == "plan"])
}
