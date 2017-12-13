package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	deployer "github.com/mdevilliers/lambda-deployer"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newUploadCommand() *cobra.Command {

	uploadCommand := &cobra.Command{
		Use:   "up",
		Short: "Upload a lambda function to S3.",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return errors.New("uploader up requires the path to a lambda package to upload e.g. './foo/bar.zip'")
			}

			pathToFile := args[0]

			if !fileExists(pathToFile) {
				return fmt.Errorf("%s does not exist", pathToFile)
			}

			_, fileName := filepath.Split(pathToFile)

			session, err := session.NewSession()

			if err != nil {
				return err
			}

			f, err := os.Open(pathToFile)

			if err != nil {
				return err
			}

			req := &s3.PutObjectInput{
				ACL:    aws.String("authenticated-read"),
				Body:   aws.ReadSeekCloser(f),
				Bucket: aws.String(_config.S3BucketName),
				Key:    aws.String(fileName),
				Metadata: map[string]*string{
					deployer.FunctionDescriptionTag: aws.String(_config.Description),
					deployer.FunctionNameTag:        aws.String(_config.FunctionName),
					deployer.FunctionHandlerTag:     aws.String(_config.Handler),
					deployer.FunctionRuntimeTag:     aws.String(_config.Runtime),
					deployer.FunctionMemorySizeTag:  aws.String(fmt.Sprintf("%d", _config.MemorySize)),
					deployer.FunctionTimeoutTag:     aws.String(fmt.Sprintf("%d", _config.Timeout)),
					deployer.FunctionAliasTag:       aws.String(_config.Alias),
				},
			}

			// get any default region if specified via an environmental variable
			// and add it to any configured regions specified via the application flags
			regionFromEnvironment := os.Getenv("AWS_REGION")

			if regionFromEnvironment != "" {
				if !stringInSlice(regionFromEnvironment, _config.Regions) {
					_config.Regions = append(_config.Regions, regionFromEnvironment)
				}
			}

			// upload to all regions
			for _, region := range _config.Regions {
				err = uploadToRegion(session, req, region)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error uploading to region : %s", region))
				}
			}

			return err

		},
	}

	_config.AddFlags(uploadCommand.Flags())

	uploadCommand.MarkFlagRequired("bucket")
	uploadCommand.MarkFlagRequired("name")
	uploadCommand.MarkFlagRequired("handler")
	uploadCommand.MarkFlagRequired("runtime")

	return uploadCommand
}

func uploadToRegion(session *session.Session, req *s3.PutObjectInput, region string) error {

	svc := s3.New(session, aws.NewConfig().WithRegion(region))

	_, err := svc.PutObject(req)

	return err

}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
