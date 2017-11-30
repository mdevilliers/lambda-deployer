## Inputs

| Name                       | Description                                                                                     | Type   | Default | Required |
| ------                     | -------------                                                                                   | :----: | :-----: | :-----:  |
| application                | The name of the application                                                                     | string | -       | yes      |
| deployer_filepath          | File path to a zip file containing the deployer                                                 | string | -       | yes      |
| env_vars                   | A map of variables to be passed to the lambda function on deployment                            | map    | `<map>` | no       |
| environment                | The name of the environment                                                                     | string | -       | yes      |
| function_role_arn          | The arn of the role the function will be deployed with                                          | string | -       | yes      |
| maximum_unaliased_versions | The number of versions without an alias to keep. A function with an alias is a function in use. | string | `3`     | no       |
| s3_bucket_arn              | The arn of the S3 bucket to use for uploading applications from CI                              | string | -       | yes      |
| s3_bucket_id               | The id of the S3 bucket to use for uploading applications from CI                               | string | -       | yes      |

