resource "tsuru_job" "scheduled_job" {
  name        = "my-scheduled-job"
  description = "Scheduled job example"
  pool        = "staging"
  team_owner  = "admin"

  container {
    image   = "ubuntu:latest"
    command = ["echo", "Hello from Tsuru Job!"]
  }

  schedule = "@every 1h"
}

resource "tsuru_job_deploy" "job_deploy" {
  job   = tsuru_job.scheduled_job.name
  image = "ubuntu:latest"
}

