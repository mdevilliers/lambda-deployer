package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

			svc := s3.New(session, aws.NewConfig())

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
					"Function-Description": aws.String(_config.Description),
					"Function-Name":        aws.String(_config.FunctionName),
					"Function-Handler":     aws.String(_config.Handler),
					"Function-Runtime":     aws.String(_config.Runtime),
					"Function-Memory-Size": aws.String(fmt.Sprintf("%d", _config.MemorySize)),
					"Function-Timeout":     aws.String(fmt.Sprintf("%d", _config.Timeout)),
					"Function-Alias":       aws.String(_config.Alias),
				},
			}

			_, err = svc.PutObject(req)

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

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
