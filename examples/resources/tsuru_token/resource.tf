resource "tsuru_token" "simple_token" {
  token_id = "my-simple-token"
  description = "My description"
  team = "team-dev"
  expires = "24h"
}