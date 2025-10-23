# Getting Started

This tutorial walks you through setting up the Tsuru provider and creating your first app.

## Setup

Create a new directory and a `main.tf` file:

```bash
mkdir tsuru-terraform
cd tsuru-terraform
```

Add the provider configuration to `main.tf`:

```terraform
terraform {
  required_providers {
    tsuru = {
      source  = "tsuru/tsuru"
      version = "~> 2.17.0"
    }
  }
}

provider "tsuru" {}
```

## Authentication

Set your Tsuru credentials as environment variables:

```bash
export TSURU_TARGET="https://tsuru.example.com"
export TSURU_TOKEN="your-auth-token"
```

Don't put credentials in your Terraform files.

## Create an app

Add this to your `main.tf`:

```terraform
resource "tsuru_app" "first_app" {
  name        = "my-first-app"
  description = "My first Terraform-managed app"
  platform    = "python"
  team_owner  = "my-team"
  pool        = "my-pool"
  plan        = "small"
}
```

Update `team_owner`, `pool`, and `plan` to match your Tsuru environment.

## Apply

```bash
terraform init
terraform plan
terraform apply
```

Type `yes` when prompted. Your app is now created in Tsuru.

## Deploy

Creating the app doesn't deploy code. Deploy using the Tsuru client:

```bash
echo "web: python app.py" > Procfile
echo "print('Hello!')" > app.py
tsuru app deploy -a my-first-app -f .
```

Check the app:

```bash
tsuru app info -a my-first-app
tsuru app log -a my-first-app
```

## Next steps

Now that you have a basic app, you can add:
- Environment variables with `tsuru_app_env`
- Custom domains with `tsuru_app_cname`
- Autoscaling with `tsuru_app_autoscale`

Check out the next tutorial to learn more.

