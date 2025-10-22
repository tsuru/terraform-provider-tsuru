# Complete Web Application Example

This example shows how to set up a production-ready web application on Tsuru with environment variables, custom domain, and autoscaling.

## What's included

This configuration creates an app with environment variables for database and cache connections, adds a custom CNAME, and configures autoscaling based on CPU usage and business hours.

## Usage

```bash
terraform init
terraform apply
```

After applying, your app will be available at `www.complete-web-app.com` with automatic scaling between 2-5 units based on CPU load, maintaining at least 3 units during weekday business hours (8am-8pm).

## Resources used

- `tsuru_app` - Base application
- `tsuru_app_env` - Environment variables
- `tsuru_app_cname` - Custom domain
- `tsuru_app_autoscale` - Scaling configuration

