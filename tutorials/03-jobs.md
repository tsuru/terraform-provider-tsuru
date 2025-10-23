# Jobs and Scheduled Tasks

Jobs are tasks that run and finish, unlike apps that run continuously.

## Create a scheduled job

Add this to a new file `job.tf`:

```terraform
resource "tsuru_job" "daily_report" {
  name        = "daily-report-job"
  description = "Generates daily report"
  team_owner  = "my-team"
  pool        = "my-pool"
  plan        = "small"

  schedule = "0 8 * * *"

  container {
    image   = "my-registry/report-generator:v1.0"
    command = ["/app/generate-report", "--type=daily"]
  }
}
```

The schedule uses cron format. This job runs every day at 8am.

## Deploy the job

```terraform
resource "tsuru_job_deploy" "report_deploy" {
  job   = tsuru_job.daily_report.name
  image = "my-registry/report-generator:v1.0"
}
```

Apply the changes:

```bash
terraform apply
```

## Check job status

```bash
tsuru job info -j daily-report-job
tsuru job log -j daily-report-job
```

## Manual jobs

For jobs that run on demand, omit the schedule:

```terraform
resource "tsuru_job" "manual_task" {
  name        = "db-migration"
  description = "Database migration job"
  team_owner  = "my-team"
  pool        = "my-pool"
  plan        = "medium"

  container {
    image   = "my-registry/db-migrator:latest"
    command = ["/app/run-migrations"]
  }
}
```

Run it manually:

```bash
tsuru job run -j db-migration
```

This is useful for maintenance tasks or migrations that need manual triggering.

