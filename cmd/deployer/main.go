package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	deployer "github.com/mdevilliers/lambda-deployer"
	aws_helper "github.com/mdevilliers/lambda-deployer/aws"
	"github.com/pkg/errors"
)

func main() {
	// DO NOTHING
}

// Policy holds information for the deployer to implement
type Policy struct {
	// MaximumUntaggedVersions is the maximum untagged versions of a lambda function
	// we want to keep. Tagged versions are never deleted.
	MaximumUntaggedVersions int
}

// S3Event struct captures the JSON structure of the event passed when a new
// object is created in S3
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

// Handle is called when ever an object is written to S3 via the uploader.
// We assume this is always a lambda function zip file and that AWS Lambda will error
// if the file is not of a correct format.
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
		return "error", errors.Wrap(err, "error un-marshaling event json")
	}

	session, err := session.NewSession()

	if err != nil {
		return "error", err
	}

	lambdaSvc := lambda.New(session, aws.NewConfig())
	s3Svc := s3.New(session, aws.NewConfig())

	bucket := s3Event.Records[0].S3.Bucket.Name
	key := s3Event.Records[0].S3.Object.Key

	meta, err := getMetadata(s3Svc, bucket, key)

	if err != nil {
		return "error", errors.Wrap(err, "error reading metadata from s3 object")
	}

	// create or update the lambda function
	conf, err := aws_helper.CreateOrUpdateFunction(lambdaSvc, bucket, key, role, meta)

	if err != nil {
		return "error", errors.Wrap(err, "error creating or updating lambda function")
	}

	// update, create the alias
	err = aws_helper.CreateOrUpdateAlias(lambdaSvc, conf, meta)

	if err != nil {
		return "error", errors.Wrap(err, "error creating or updating alias")
	}

	return "ok", nil

}

// getMetadata parses the S3 object metadata
func getMetadata(svc *s3.S3, s3Bucket, s3Key string) (deployer.FunctionMetadata, error) {

	req := &s3.HeadObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
	}

	resp, err := svc.HeadObject(req)

	if err != nil {
		return deployer.FunctionMetadata{}, err
	}

	memorySize, err := strconv.ParseInt(*(resp.Metadata["Function-Memory-Size"]), 10, 64)

	if err != nil {
		return deployer.FunctionMetadata{}, errors.Wrap(err, "cannot parse function-memory-size")
	}

	timeout, err := strconv.ParseInt(*(resp.Metadata["Function-Timeout"]), 10, 64)

	if err != nil {
		return deployer.FunctionMetadata{}, errors.Wrap(err, "cannot parse function-timeout")
	}

	meta := deployer.FunctionMetadata{
		Description:  *(resp.Metadata["Function-Description"]),
		FunctionName: *(resp.Metadata["Function-Name"]),
		Handler:      *(resp.Metadata["Function-Handler"]),
		Runtime:      *(resp.Metadata["Function-Runtime"]),
		MemorySize:   int64(memorySize),
		Timeout:      int64(timeout),
		Alias:        *(resp.Metadata["Function-Alias"]),
		EnvVars:      map[string]interface{}{},
	}

	// add in any environmental variables set in the terraform
	envVars := os.Getenv("DEPLOYER_FUNCTION_ENV_VARS")

	err = json.Unmarshal([]byte(envVars), &meta.EnvVars)

	if err != nil {
		return deployer.FunctionMetadata{}, errors.Wrap(err, "error un-marshaling envionmental vars")
	}

	return meta, nil

}
