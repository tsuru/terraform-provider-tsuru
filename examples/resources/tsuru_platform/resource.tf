resource "tsuru_platform" "python" {
  name     = "python"
  disabled = false
  dockerfile = <<-EOT
    FROM tsuru/python:latest
    EOT
}

# `image` is a convenience alternative to `dockerfile`: it generates a
# Dockerfile with a single `FROM ${image}` instruction. `image` and
# `dockerfile` are mutually exclusive.
resource "tsuru_platform" "nodejs" {
  name  = "nodejs"
  image = "tsuru/nodejs:latest"
}
