# Scheduled Job Example

This example demonstrates how to create and deploy a scheduled job that runs every hour.

## Usage

```bash
terraform init
terraform apply
```

The job will run automatically every hour. You can check its execution logs with:

```bash
tsuru job info -j my-scheduled-job
tsuru job log -j my-scheduled-job
```

## Resources used

- `tsuru_job` - Job definition with schedule
- `tsuru_job_deploy` - Job deployment

