resource "tsuru_job_env" "my-envs" {
		job = "sample-job"
		environment_variables = {
			public_env = "public_value"
		}
		private_environment_variables = {
			private_env = "private_value"
		}
	}
