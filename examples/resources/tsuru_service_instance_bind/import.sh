# for apps
terraform import service_instance_bind.resource_name "service::instance::app"
# example
terraform import service_instance_bind.instance_bind "service01::my-instance::app01"

# for jobs
terraform import service_instance_bind.resource_name "service::instance::tsuru-job::job"
# example
terraform import service_instance_bind.instance_bind "service01::my-instance::tsuru-job::job01"
