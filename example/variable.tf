variable "host_provider" {
  type        = string
  description = "host-company"
  default     = "https://tsuru.mycompany.com"
}

variable "tsuru_app" {
  type = map(any)
  default = {
    "name" = "sample-app"
    "description" = "app created with terraform"
    "plan" = "c0.1m0.2"
    "pool" = "staging"
    "platform" = "python"
    "team_owner" = "admin"
  }
}

variable "labels" {
  type = map(any)
  default = {
    "label1" = "1"
    "io.tsuru/a"  = "2"
  }
}

variable "annotations" {
  type = map(any)
  default = {
    "annotation" = "value"
    "io.gcp/key"  = "2"
  }
}

variable restart_on_update{
  default = true
}

variable tsuru_app_grant{
  type        = list(string)
  description = "name-teams"
  default     = ["team-a", "team-b"]
}

variable tsuru_app_cname {
    type = string
    description = "host cname"
    default = "sample.tsuru.i.mycompany.com"
}

variable tsuru_app_router {
    type = map(any)
    default = {
      name = "internal-http-lb"
      "key" = "value"
  }
}

variable "tsuru_app_autoscale" {
  type = map(any)
  default = {
    "process" = "web"
    "min_units" = 3
    "max_units" = 10
    "cpu_average" = "800m"
  }
}