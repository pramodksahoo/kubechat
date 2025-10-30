package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "unknown"
var Commit = "unknown"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of kubechat",
	Long:  `All software has a version (semantic at best). This is kubechat's'`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version:", Version)
		fmt.Println("commit:", Commit)
	},
}
