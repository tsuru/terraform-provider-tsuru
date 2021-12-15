resource "tsuru_webhook" "webhook1" {
  name        = "webhook1"
  description = "my event"
  team_owner  = "myteam"

  event_filter {
    target_types = [
      "target01",
      "target02",
    ]
    target_values = [
      "targetvalue01",
      "targetvalue02",
    ]

    kind_types = [
      "kind_type"
    ]

    kind_names = [
      "kind_name"
    ]

    error_only   = false
    success_only = true
  }

  url       = "http://blah.io/webhook"
  proxy_url = "http://myproxy.com"
  headers = {
    "X-Token" = "my-token"
  }

  method   = "POST"
  body     = "body-test"
  insecure = true
}
