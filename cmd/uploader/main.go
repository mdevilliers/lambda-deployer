package main

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/mdevilliers/lambda-deployer/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// uploads zip files to s3 bucket adding some object metadata for the dployer to read
var RootCmd = &cobra.Command{
	Use: "uploader",
}

var _config = newConfig()

type config struct {
	S3BucketName string `envconfig:"S3BUCKET_NAME"`
	Description  string `envconfig:"FUNCTION_DESCRIPTION"`
	FunctionName string `envconfig:"FUNCTION_NAME"`
	Handler      string `envconfig:"FUNCTION_HANDLER"`
	Runtime      string `envconfig:"FUNCTION_RUNTIME"`
	MemorySize   int    `envconfig:"FUNCTION_MEMORY_SIZE" default:"128"`
	Timeout      int    `envconfig:"FUNCTION_TIMEOUT" default:"60"`
	Alias        string `envconfig:"FUNCTION_ALIAS" default:"$LATEST"`
}

func newConfig() *config {
	c := &config{}
	err := envconfig.Process("", c)

	if err != nil {
		panic("error processing environmental configuration")
	}

	return c
}

func (o *config) AddFlags(fs *pflag.FlagSet) {

	fs.StringVarP(&o.S3BucketName, "bucket", "b", o.S3BucketName, "bucket to upload to")
	fs.StringVarP(&o.Description, "desc", "d", o.Description, "description of function (optional)")
	fs.StringVarP(&o.FunctionName, "name", "n", o.FunctionName, "name of function")
	fs.StringVarP(&o.Handler, "handler", "e", o.Handler, "function handler")
	fs.StringVarP(&o.Runtime, "runtime", "r", o.Runtime, "runtime to use")
	fs.IntVarP(&o.MemorySize, "mem", "m", o.MemorySize, "memory to use")
	fs.IntVarP(&o.Timeout, "timeout", "t", o.Timeout, "timeout to use in seconds")
	fs.StringVarP(&o.Alias, "alias", "a", o.Alias, "alias to use")
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
