terraform {
  required_version = ">= 0.10, < 1.0"
}

provider "aws" {
  version = "~> 0.1"
  region  = "eu-west-1"
}

# create a bucket for uploads
resource "aws_s3_bucket" "deployment_uploads" {
  bucket = "deployment-uploads"
}

# create a role for the deployed functions to use
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

# we need a service account for the uploader
resource "aws_iam_user" "uploader" {
  name = "srv_lambda_uploader_${aws_s3_bucket.deployment_uploads.bucket_domain_name}"
}

# generate keys for uploader service account 
resource "aws_iam_access_key" "uploader" {
  user = "${aws_iam_user.uploader.name}"
}

# grant user access to the bucket
resource "aws_s3_bucket_policy" "bucket_policy" {
  bucket = "${aws_s3_bucket.deployment_uploads.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "${aws_iam_user.uploader.arn}"
      },
      "Action": [ "s3:*"],
      "Resource": [
        "${aws_s3_bucket.deployment_uploads.arn}",
        "${aws_s3_bucket.deployment_uploads.arn}/*"
      ]
    }
  ]
}
EOF
}

module "auto_deployer" {
  source = "../modules/lamda-deployer/"

  application       = "example"
  environment       = "staging"
  deployer_filepath = "../../deployer.zip"

  function_role_arn = "${aws_iam_role.my_lambda_role.arn}"
  s3_bucket_arn     = "${aws_s3_bucket.deployment_uploads.arn}"
  s3_bucket_id      = "${aws_s3_bucket.deployment_uploads.id}"
}

// the s3 bucket for uploads
output "s3_bucket" {
  value = "${aws_s3_bucket.deployment_uploads.bucket_domain_name}"
}

// the uploader user name
output "user_name" {
  value = "${aws_iam_user.uploader.name}"
}

// the uploader access key
output "iam_access_key_id" {
  value = "${aws_iam_access_key.uploader.id}"
}

//the uploader access key secret
output "iam_access_key_secret" {
  value = "${aws_iam_access_key.uploader.secret}"
}
