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

