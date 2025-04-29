terraform import tsuru_certificate_issuer.resource_name "app::cname::issuer"
# or
terraform import tsuru_certificate_issuer.resource_name "app::cname"

# example
terraform import tsuru_certificate_issuer.cert "my-app::my-domain.com::letsencrypt"
# or just (issuer will be discovered automatically)
terraform import tsuru_certificate_issuer.cert "my-app::my-domain.com"