package cmd

import (
	"fmt"

	deployer "github.com/mdevilliers/lambda-deployer"
	"github.com/spf13/cobra"
)

var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(deployer.VersionString())
		return nil
	},
}
