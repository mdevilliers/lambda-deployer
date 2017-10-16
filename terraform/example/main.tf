terraform {
  required_version = ">= 0.10, < 1.0"
}

provider "aws" {
  version = "~> 0.1"
  region  = "eu-west-1"
}


resource "aws_s3_bucket" "deployment_uploads" {
  bucket = "deployment-uploads"
}

module "auto_deployer" {
  source = "../modules/lamda-deployer/"

  application = "my-lambda-function-deployer"
  environment = "staging"
  deployer_filepath = "../../deployer.zip"
  s3_bucket_arn = "${aws_s3_bucket.deployment_uploads.arn}"
  s3_bucket_id = "${aws_s3_bucket.deployment_uploads.id}"
}

output "s3_bucket" {
    value = "${aws_s3_bucket.deployment_uploads.bucket_domain_name}"
}
