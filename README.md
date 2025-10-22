# Terraform Provider Tsuru - Improved Documentation

Better docs for the terraform-provider-tsuru with practical examples, migration guides, detailed resource documentation, and step-by-step tutorials.

## Quick Links

**New to Terraform + Tsuru?** Start with [Getting Started](tutorials/01-getting-started.md)

**Migrating from manual management?** Read the [Migration Guide](guides/migration-guide.md)

**Need a reference?** Check the [Resource Documentation](resources/)

## Tutorials

| Tutorial | What you'll learn | Time |
|----------|-------------------|------|
| [Getting Started](tutorials/01-getting-started.md) | Initial setup and first app | 15 min |
| [Managing Apps](tutorials/02-managing-apps.md) | Environment variables, CNAMEs, autoscaling | 20 min |
| [Jobs](tutorials/03-jobs.md) | Scheduled and manual jobs | 15 min |
| [Infrastructure](tutorials/04-infrastructure.md) | Clusters, pools, and plans | 25 min |

## Examples

Ready-to-use examples for common scenarios:

- [Complete Web App](examples/complete-web-app/) - App with env vars, CNAME, and autoscaling
- [Scheduled Job](examples/scheduled-job/) - Job with cron schedule
- [Cluster Setup](examples/cluster-setup/) - Cluster and pool configuration

## Resource Documentation

Detailed docs for key resources:

- [tsuru_app](resources/app.md) - Application management
- [tsuru_cluster](resources/cluster.md) - Kubernetes cluster configuration

## Migration Guide

Moving from manual Tsuru management to Terraform? The [Migration Guide](guides/migration-guide.md) covers:

- Migration strategies (big bang vs gradual)
- Using `terraform import`
- Import order and validation
- Rollback procedures

## Contributing

Found an issue or want to add more examples? Contributions welcome.

## Additional Resources

- [Official Provider Docs](https://registry.terraform.io/providers/tsuru/tsuru/latest/docs)
- [Tsuru Documentation](https://docs.tsuru.io/)
- [Provider Repository](https://github.com/tsuru/terraform-provider-tsuru)

