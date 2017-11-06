lambda-deployer
---------------

[![CircleCI](https://circleci.com/gh/mdevilliers/lambda-deployer.svg?style=svg)](https://circleci.com/gh/mdevilliers/lambda-deployer)

Aim 
---

Deploy and update lambda functions easily and securely from either a developers machine or CI system. 


Goals
-----

- secure by default 
  - uploading from CI with a minimum set of AWS permissions 
  - sensitive configuration information is managed by terraform
- easy to integrate CI or the developer workflow
- integrate cleanly with existing terraform environments

Usage
-----

- create an AWS IAM user with the permissions to PutObject on a known S3 bucket
- deploy the `lambda-deployer` as an AWS lambda function 

TODO : INSERT IMAGE

- configure the uploader 

```
export AWS_ACCESS_KEY_ID=**************
export AWS_SECRET_ACCESS_KEY=***********************
```
- use the `uploader` executable 

```
            uploader up -b myS3Bucket \
                          -a myAlias \
                          -d "AUTOMATED DEPLOY" \
                          -e handler.Handle \
                          -r python2.7 \
                          -n myFunction /path/to/function.zip

```
- on upload to S3 the function.zip file contains metadata for the deployer to use

TODO : INSERT IMAGE

- the deployer will deploy the function
