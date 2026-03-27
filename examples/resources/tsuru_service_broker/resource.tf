# tsuru service broekr with basic_auth_config
resource "tsuru_service_broker" "test_broker" {
  name = "test_name"
  url  = "https://broker.example.com"
  config {
    insecure = false
    context = {
      test_context_1 = "TEST_CONTEXT_1"
      test_context_2 = "TEST_CONTEXT_2"
    }
    cache_expiration_seconds = 3600
    auth_config {
      basic_auth_config {
        username = "TEST_USERNAME"
        password = "TEST_PASSWORD"
      }
    }
  }
}

# tsuru service broker with bearer_config
resource "tsuru_service_broker" "test_broker" {
  name = "test_name"
  url  = "https://broker.example.com"
  config {
    insecure = true
    context = {
      test_context_1 = "TEST_CONTEXT_1"
      test_context_2 = "TEST_CONTEXT_2"
    }
    cache_expiration_seconds = 3600
    auth_config {
      bearer_config {
        token = "SENSITIVE_TOKEN"
      }
    }
  }
}
