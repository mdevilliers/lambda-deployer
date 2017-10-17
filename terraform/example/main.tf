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

resource "aws_iam_role" "my_lambda_role" {
  name = "my_lambda_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

module "auto_deployer" {
  source = "../modules/lamda-deployer/"

  application       = "example-deployer"
  environment       = "staging"
  deployer_filepath = "../../deployer.zip"

  function_role_arn = "${aws_iam_role.my_lambda_role.arn}"
  s3_bucket_arn     = "${aws_s3_bucket.deployment_uploads.arn}"
  s3_bucket_id      = "${aws_s3_bucket.deployment_uploads.id}"
}

output "s3_bucket" {
  value = "${aws_s3_bucket.deployment_uploads.bucket_domain_name}"
}
