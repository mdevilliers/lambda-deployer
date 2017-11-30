variable "application" {
  description = "The name of the application"
}

variable "environment" {
  description = "The name of the environment"
}

variable "s3_bucket_arn" {
  description = "The arn of the S3 bucket to use for uploading applications from CI"
}

variable "s3_bucket_id" {
  description = "The id of the S3 bucket to use for uploading applications from CI"
}

variable "deployer_filepath" {
  description = "File path to a zip file containing the deployer"
}

variable "function_role_arn" {
  description = "The arn of the role the function will be deployed with"
}

variable "env_vars" {
  type        = "map"
  description = "A map of variables to be passed to the lambda function on deployment"
  default     = {}
}

variable "maximum_unaliased_versions" {
  description = "The number of versions without an alias to keep. A function with an alias is a function in use."
  default     = 3
}
