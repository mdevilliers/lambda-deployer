package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
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
	//policy := Policy{
	//	AutoDeploy: true,
	//}

	session, err := session.NewSession()

	if err != nil {
		return "error", err
	}

	// read file from s3
	region := s3Event.Records[0].AwsRegion
	bucket := s3Event.Records[0].S3.Bucket.Name
	key := s3Event.Records[0].S3.Object.Key

	encoded, err := base64EncodedShaForS3File(session, bucket, key, region)

	if err != nil {
		return "error", err
	}

	log.Println("sha : ", encoded)

	meta := FunctionMetadata{
		Description:  "description",
		FunctionName: "lambda_rules",
		Handler:      "index.handler",
		Runtime:      "nodejs4.3",
	}

	err = deployLambdaFunction(session, bucket, key, role, meta)

	if err == nil {
		return "error", err
	}

	return "ok", nil

}

func base64EncodedShaForS3File(sess *session.Session, bucket, key, region string) (string, error) {

	svc := s3.New(sess, aws.NewConfig().WithRegion(region))

	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	log.Println("request : ", req)

	results, err := svc.GetObject(req)

	if err != nil {
		return "", errors.Wrap(err, "error loading file from s3 API")
	}
	defer results.Body.Close()

	sha := sha256.New()

	if _, err := io.Copy(sha, results.Body); err != nil {
		return "", errors.Wrap(err, "error reading file from s3")
	}

	shaSum := sha.Sum(nil)
	encoded := base64.StdEncoding.EncodeToString(shaSum[:])
	return encoded, nil

}

func deployLambdaFunction(sess *session.Session, s3Bucket, s3Key, role string, metadata FunctionMetadata) error {

	svc := lambda.New(sess, aws.NewConfig())

	req := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			S3Bucket: aws.String(s3Bucket),
			S3Key:    aws.String(s3Key),
		},
		Description:  aws.String(metadata.Description),
		FunctionName: aws.String(metadata.FunctionName),
		Handler:      aws.String(metadata.Handler),
		MemorySize:   aws.Int64(128),
		Publish:      aws.Bool(true),
		Role:         aws.String(role),
		Runtime:      aws.String(metadata.Runtime),
		Timeout:      aws.Int64(15),
	}

	result, err := svc.CreateFunction(req)

	log.Println("create function : ", result, err)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
