terraform import tsuru_certificate_issuer.resource_name "app::cname::issuer"

# example
terraform import tsuru_certificate_issuer.cert "my-app::my-domain.com::letsencrypt"