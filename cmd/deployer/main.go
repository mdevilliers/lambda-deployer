package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	deployer "github.com/mdevilliers/lambda-deployer"
	"github.com/pkg/errors"
)

func main() {
	// DO NOTHING
}

type Policy struct {
	AutoDeploy bool
}

type FunctionMetadata struct {
	Description  string
	FunctionName string
	Handler      string
	Runtime      string
	MemorySize   int64
	Timeout      int64
	Alias        string
}

type S3Event struct {
	Records []struct {
		EventVersion      string    `json:"eventVersion"`
		EventTime         time.Time `json:"eventTime"`
		RequestParameters struct {
			SourceIPAddress string `json:"sourceIPAddress"`
		} `json:"requestParameters"`
		S3 struct {
			ConfigurationID string `json:"configurationId"`
			Object          struct {
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
				Key       string `json:"key"`
				Size      int    `json:"size"`
			} `json:"object"`
			Bucket struct {
				Arn           string `json:"arn"`
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalID string `json:"principalId"`
				} `json:"ownerIdentity"`
			} `json:"bucket"`
			S3SchemaVersion string `json:"s3SchemaVersion"`
		} `json:"s3"`
		ResponseElements struct {
			XAmzID2       string `json:"x-amz-id-2"`
			XAmzRequestID string `json:"x-amz-request-id"`
		} `json:"responseElements"`
		AwsRegion    string `json:"awsRegion"`
		EventName    string `json:"eventName"`
		UserIdentity struct {
			PrincipalID string `json:"principalId"`
		} `json:"userIdentity"`
		EventSource string `json:"eventSource"`
	} `json:"Records"`
}

func Handle(evt json.RawMessage, ctx *runtime.Context) (string, error) {

	log.Println("deployer : ", deployer.VersionString())
	log.Println("handle event : ", string(evt))

	role := os.Getenv("DEPLOYER_FUNCTION_ROLE_ARN")

	if role == "" {
		return "error", errors.New("DEPLOYER_FUNCTION_ROLE_ARN not set")
	}

	s3Event := S3Event{}

	err := json.Unmarshal(evt, &s3Event)

	if err != nil {
		return "error", errors.Wrap(err, "error unmarshalling event json")
	}

	// assume auto deployment
	// policy := Policy{
	//	AutoDeploy: true,
	//}

	session, err := session.NewSession()

	if err != nil {
		return "error", err
	}

	svc := lambda.New(session, aws.NewConfig())

	//region := s3Event.Records[0].AwsRegion
	bucket := s3Event.Records[0].S3.Bucket.Name
	key := s3Event.Records[0].S3.Object.Key

	// TODO : make this dynamic
	meta := FunctionMetadata{
		Description:  "description",
		FunctionName: "lambda_rules",
		Handler:      "index.handler",
		Runtime:      "nodejs4.3",
		MemorySize:   128,
		Timeout:      15,
		Alias:        "xxx-latest",
	}

	var functionConfiguration *lambda.FunctionConfiguration

	// update, create the function
	exists, err := functionExists(svc, meta.FunctionName)

	if err != nil {
		return "error", err
	}

	if exists {

		functionConfiguration, err = updateLambdaFunction(svc, bucket, key, meta)

		if err != nil {
			return "error", err
		}

	} else {

		functionConfiguration, err = deployLambdaFunction(svc, bucket, key, role, meta)

		if err != nil {
			return "error", err
		}
	}

	// update, create the alias
	exists, err = aliasExists(svc, meta.FunctionName, meta.Alias)

	if exists {

		err = updateAlias(svc, meta.FunctionName, meta.Alias, *functionConfiguration.Version)

		if err != nil {
			return "error", err
		}

	} else {

		err = createAlias(svc, meta.FunctionName, meta.Alias, *functionConfiguration.Version)

		if err != nil {
			return "error", err
		}

	}

	return "ok", nil

}

func functionExists(svc *lambda.Lambda, name string) (bool, error) {

	req := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	resp, err := svc.GetFunction(req)

	log.Println("GetFunction : ", resp, err)

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

func updateAlias(svc *lambda.Lambda, functionName, aliasName, functionVersion string) error {

	req := &lambda.UpdateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(functionVersion),
	}

	resp, err := svc.UpdateAlias(req)

	log.Println("UpdateAlias : ", resp, err)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}

func createAlias(svc *lambda.Lambda, functionName, aliasName, functionVersion string) error {

	req := &lambda.CreateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(functionVersion),
	}

	resp, err := svc.CreateAlias(req)

	log.Println("CreateAlias : ", resp, err)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}

func aliasExists(svc *lambda.Lambda, functionName, aliasName string) (bool, error) {

	req := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	}

	resp, err := svc.GetAlias(req)

	log.Println("GetAlias : ", resp, err)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeResourceNotFoundException:
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}

func updateLambdaFunction(svc *lambda.Lambda, s3Bucket, s3Key string, metadata FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	req := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(metadata.FunctionName),
		Publish:      aws.Bool(true),
		S3Bucket:     aws.String(s3Bucket),
		S3Key:        aws.String(s3Key),
	}

	resp, err := svc.UpdateFunctionCode(req)

	log.Println("UpdateFunctionCode : ", resp, err)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

func deployLambdaFunction(svc *lambda.Lambda, s3Bucket, s3Key, role string, metadata FunctionMetadata) (*lambda.FunctionConfiguration, error) {

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
	}

	resp, err := svc.CreateFunction(req)

	log.Println("CreateFunction : ", resp, err)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}
