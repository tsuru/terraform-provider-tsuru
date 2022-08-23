terraform import tsuru_app_autoscale.resource_name "app::process[::version]"

# example
terraform import tsuru_app_unit.other-unit "sample-app::web"
terraform import tsuru_app_unit.other-unit-with-version "sample-app::worker::2"
