package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/mdevilliers/lambda-deployer/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// uploads zip files to s3 bucket adding some object metadata for the deployer to read
var RootCmd = &cobra.Command{
	Use: "uploader",
}

var _config = newConfig()

type config struct {
	S3BucketName string   `envconfig:"S3BUCKET_NAME"`
	Description  string   `envconfig:"FUNCTION_DESCRIPTION"`
	FunctionName string   `envconfig:"FUNCTION_NAME"`
	Handler      string   `envconfig:"FUNCTION_HANDLER"`
	Runtime      string   `envconfig:"FUNCTION_RUNTIME"`
	MemorySize   int      `envconfig:"FUNCTION_MEMORY_SIZE" default:"128"`
	Timeout      int      `envconfig:"FUNCTION_TIMEOUT" default:"60"`
	Alias        string   `envconfig:"FUNCTION_ALIAS" default:"$LATEST"`
	Regions      []string `envconfig:"ADDITIONAL_AWS_REGIONS"`
}

func newConfig() *config {
	c := &config{
		Regions: []string{},
	}
	err := envconfig.Process("", c)

	if err != nil {
		panic("error processing environmental configuration")
	}

	return c
}

func (o *config) AddFlags(fs *pflag.FlagSet) {

	fs.StringVarP(&o.S3BucketName, "bucket", "b", o.S3BucketName, "AWS S3 bucket to upload function to")
	fs.StringVarP(&o.Description, "desc", "d", o.Description, "Function description (optional)")
	fs.StringVarP(&o.FunctionName, "name", "n", o.FunctionName, "Function name")
	fs.StringVarP(&o.Handler, "handler", "e", o.Handler, "Function handler")
	fs.StringVarP(&o.Runtime, "runtime", "r", o.Runtime, "AWS Lambda Runtime to use")
	fs.IntVarP(&o.MemorySize, "mem", "m", o.MemorySize, "Memory to use")
	fs.IntVarP(&o.Timeout, "timeout", "t", o.Timeout, "Function timeout to use in seconds")
	fs.StringVarP(&o.Alias, "alias", "a", o.Alias, "Function Alias to use")
	fs.StringSliceVarP(&o.Regions, "region", "g", o.Regions, "Additional AWS Regions to upload lambda function to")
}

func init() {
	RootCmd.AddCommand(newUploadCommand())
	RootCmd.AddCommand(cmd.VersionCommand)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
