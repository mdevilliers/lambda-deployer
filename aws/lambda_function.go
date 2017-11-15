package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	deployer "github.com/mdevilliers/lambda-deployer"
	"github.com/pkg/errors"
)

func CreateOrUpdateFunction(svc *lambda.Lambda, bucket, key, role string, meta deployer.FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	exists, err := functionExists(svc, meta.FunctionName)

	if err != nil {
		return nil, err
	}

	if exists {

		return updateLambdaFunction(svc, bucket, key, meta)

	}

	return newLambdaFunction(svc, bucket, key, role, meta)

}

func functionExists(svc *lambda.Lambda, name string) (bool, error) {

	req := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	_, err := svc.GetFunction(req)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeResourceNotFoundException:
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func updateLambdaFunction(svc *lambda.Lambda, s3Bucket, s3Key string, metadata deployer.FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	req := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(metadata.FunctionName),
		Publish:      aws.Bool(true),
		S3Bucket:     aws.String(s3Bucket),
		S3Key:        aws.String(s3Key),
	}

	resp, err := svc.UpdateFunctionCode(req)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

func newLambdaFunction(svc *lambda.Lambda, s3Bucket, s3Key, role string, metadata deployer.FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	req := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			S3Bucket: aws.String(s3Bucket),
			S3Key:    aws.String(s3Key),
		},
		Description:  aws.String(metadata.Description),
		FunctionName: aws.String(metadata.FunctionName),
		Handler:      aws.String(metadata.Handler),
		MemorySize:   aws.Int64(metadata.MemorySize),
		Publish:      aws.Bool(true),
		Role:         aws.String(role),
		Runtime:      aws.String(metadata.Runtime),
		Timeout:      aws.Int64(metadata.Timeout),
		Environment: &lambda.Environment{
			Variables: map[string]*string{},
		},
	}

	for k, v := range metadata.EnvVars {
		req.Environment.Variables[k] = aws.String(fmt.Sprintf("%v", v))
	}

	resp, err := svc.CreateFunction(req)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}
