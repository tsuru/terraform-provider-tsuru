resource "tsuru_cluster" "test_cluster" {
  name              = "test_cluster"
  tsuru_provisioner = "kubernetes"
  default           = true
  http_proxy        = "http://myproxy.io:3128"
  kube_config {
    cluster {
      server                     = "https://mycluster.local"
      tls_server_name            = "mycluster.local"
      insecure_skip_tls_verify   = true
      certificate_authority_data = "server-cert"
    }
    user {
      auth_provider {
        name = "tsuru"
        config = {
          "tsuru-flag-01" = "result"
        }
      }
      client_certificate_data = "client-cert"
      client_key_data         = "client-key"
      token                   = "token"
      username                = "username"
      password                = "password"
      exec {
        api_version = "api-version"
        command     = "tsuru"
        args        = ["cluster", "login"]
        env {
          name  = "TSURU_TOKEN"
          value = "token"
        }
      }
    }
  }
}
