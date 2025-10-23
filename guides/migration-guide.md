# Migration Guide

Moving from manual Tsuru management to Terraform gives you version control, automation, and reproducibility for your infrastructure.

## Before you start

Take inventory of your current Tsuru resources. List everything you have:

```bash
tsuru app list
tsuru service instance list
tsuru pool list
tsuru cluster list
tsuru job list
```

Save this output - you'll need it to write your Terraform configs.

## Migration strategies

**Big Bang**
Import everything at once. Faster but riskier. Best for small environments or non-production.

**Gradual**
Import resources in phases. Safer for production. Start with one team or a few non-critical apps.

## How terraform import works

Terraform import brings existing resources under Terraform management without recreating them. You need to:

1. Write the resource definition in your `.tf` file
2. Run `terraform import` to link it to the real resource
3. Run `terraform plan` to verify everything matches

### Example: Importing an app

First, write the resource definition:

```terraform
resource "tsuru_app" "my_app" {
  name        = "my-app"
  description = "Imported app"
  plan        = "c0.1m0.2"
  pool        = "staging"
  platform    = "python"
  team_owner  = "admin"
}
```

Then import it:

```bash
terraform import tsuru_app.my_app "my-app"
```

Check if it worked:

```bash
terraform plan
```

If you see changes, your definition doesn't match reality. Adjust it until `terraform plan` shows no changes.

## Import order

Follow this sequence to avoid dependency issues:

1. Clusters and pools
2. Apps and their resources (env vars, CNAMEs, autoscale)
3. Services and bindings
4. Jobs

## Validation

After each import:

- Run `terraform plan` - should show zero changes
- Test the app/service to make sure it still works
- Commit your code

## Rollback

If something breaks:

- Revert to the previous git commit
- Run `terraform apply` to restore the old state
- Or remove the resource from Terraform with `terraform state rm`

## Checklist

- [ ] Inventory all Tsuru resources
- [ ] Choose migration strategy
- [ ] Set up Terraform with tsuru provider
- [ ] Write resource definitions
- [ ] Import clusters and pools
- [ ] Import apps and related resources
- [ ] Import services
- [ ] Import jobs
- [ ] Validate with `terraform plan` after each step
- [ ] Test apps functionality
- [ ] Backup `terraform.tfstate`

