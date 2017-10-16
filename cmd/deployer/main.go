package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

	/*{
	      "Records": [
	          {
	              "eventVersion": "2.0",
	              "eventTime": "2017-10-16T12:44:56.226Z",
	              "requestParameters": {
	                  "sourceIPAddress": "88.202.148.160"
	              },
	              "s3": {
	                  "configurationId": "tf-s3-lambda-005caa139621032f5d027c6ea5",
	                  "object": {
	                      "eTag": "963b9bc5f569b424fb9e8138bb10830b",
	                      "sequencer": "0059E4A9C806181953",
	                      "key": "cayley_notes",
	                      "size": 4437
	                  },
	                  "bucket": {
	                      "arn": "arn:aws:s3:::deployment-uploads",
	                      "name": "deployment-uploads",
	                      "ownerIdentity": {
	                          "principalId": "A3NCRLD2SUC3LM"
	                      }
	                  },
	                  "s3SchemaVersion": "1.0"
	              },
	              "responseElements": {
	                  "x-amz-id-2": "sV8yZMSDnraI7KuxSqc//yBlhRFcwux3FL3wS9wlyRXCH2SkG52q0DhPdGsAqhfY13N0gmhT25E=",
	                  "x-amz-request-id": "FCB9FE297224979F"
	              },
	              "awsRegion": "eu-west-1",
	              "eventName": "ObjectCreated:Put",
	              "userIdentity": {
	                  "principalId": "AWS:AIDAJLVR2WITNP4II2UD2"
	              },
	              "eventSource": "aws:s3"
	          }
	      ]
	  }
	*/

	s3Event := S3Event{}

	err := json.Unmarshal(evt, &s3Event)

	if err != nil {
		return "error", errors.Wrap(err, "error unmarshalling event json")
	}

	// assume auto deployment
	//policy := Policy{
	//	AutoDeploy: true,
	//}

	// read file from s3
	region := s3Event.Records[0].AwsRegion
	bucket := s3Event.Records[0].S3.Bucket.Name
	key := s3Event.Records[0].S3.Object.Key

	encoded, err := base64EncodedShaForS3File(bucket, key, region)

	if err != nil {
		return "error", err
	}

	log.Println("sha : ", encoded)

	return "ok", nil

}

func base64EncodedShaForS3File(bucket, key, region string) (string, error) {

	sess := session.New()
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
