resource "tsuru_pool_constraint" "my-pool-constraint" {
	pool_expr = "my-pool"
	field = "router"
	values = [
		"load-balancer",
		"ingress"
	]
}
