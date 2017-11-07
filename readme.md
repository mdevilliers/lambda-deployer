lambda-deployer
---------------

[![CircleCI](https://circleci.com/gh/mdevilliers/lambda-deployer.svg?style=svg)](https://circleci.com/gh/mdevilliers/lambda-deployer)

Aim
---

Deploy and update lambda functions easily and securely from either a developers machine or CI system.

Goals
-----

- secure by default
  - upload and configure lambda functions with a minimum set of AWS permissions.
  - sensitive configuration information e.g. database connection credentials are not exposed to either a CI system, developer machine or (shock horror) Github!
- easy to integrate CI or the developer workflow
- integrate cleanly with existing AWS environments

Usage
-----

- download a [release](https://github.com/thingful/daas/releases)
- create an AWS S3 bucket to handle the uploads of deployment packages.
- create an AWS IAM role with the permissions to PutObject on the S3 bucket. This users credentials will be used to upload packages to S3.
- create an AWS IAM role with the permissions your lambda function needs. This users credentials will be used to run your lambda function.
- deploy the `lambda-deployer` as an AWS lambda function using the terraform module.

```
module "auto_deployer" {
  source = "git@github.com:mdevilliers/lambda-deployer//terraform/modules/lamda-deployer"

  application       = "${var.application}" // name of your application
  environment       = "${var.environment}" // logical environment e.g. production
  deployer_filepath = "../../etc/deployer.zip" // path to the deployer zip file

  function_role_arn = "${aws_iam_role.lambda_exec_role.arn}" // arn of the AWS IAM role your function needs
  s3_bucket_arn     = "${aws_s3_bucket.deployment_uploads.arn}" // arn of the AWS S3 bucket to monitor for uploads
  s3_bucket_id      = "${aws_s3_bucket.deployment_uploads.id}" // name of the AWS S3 bucket bucket to monitor for uploads

  env_vars = {
    variables = {
      FOO          = "BAR" // variables to configure the lambda function with
    }
  }
}

```

There is an example terraform package using the terraform [module](https://github.com/mdevilliers/lambda-deployer/tree/master/terraform)

TODO : INSERT IMAGE

- download and configure the uploader with the credentials of the upload user, the name of the S3 bucket and properties for your lambda function.

```
export AWS_ACCESS_KEY_ID=**************
export AWS_SECRET_ACCESS_KEY=***********************

uploader up -b myS3Bucket \
            -a myAlias \
            -d "AUTOMATED DEPLOY" \
            -e handler.Handle \
            -r python2.7 \
            -n myFunction /path/to/function.zip

```

On upload to S3 the function.zip file contains metadata for the deployer to use

![metadata]( docs/metadata.jpg)

- the deployer will deploy the function

![function]( docs/function.jpg)

