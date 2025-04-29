resource "tsuru_certificate_issuer" "cert" {
  app    = tsuru_app.my-app.name
  cname  = "myapp.mydomain.com"
  issuer = "letsencrypt"
}

# Example with app cname added separately
resource "tsuru_app_cname" "app-cname" {
  app      = tsuru_app.my-app.name
  hostname = "anotherapp.mydomain.com"
}

resource "tsuru_certificate_issuer" "cert-for-cname" {
  app    = tsuru_app.my-app.name
  cname  = tsuru_app_cname.app-cname.hostname
  issuer = "letsencrypt"
}