Example of integrating the lambda-deployer into a terraform module.

Steps are -
- create an S3 bucket - for the uploader to put lambda zip files in 
- create a user with the correct role for the uploader user 
- optionally create some variables to be passed to the configured function 
- output the keys for the uploader user

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| environment_variables | create some variables for the deployed lambda function defining them here means they not exposed to either a developers or CI environment | map | `<map>` | no |

## Outputs

| Name | Description |
|------|-------------|
| iam_access_key_id | the uploader access key |
| iam_access_key_secret | the uploader access key secret |
| s3_bucket | the s3 bucket for uploads |
| user_name | the uploader user name |

