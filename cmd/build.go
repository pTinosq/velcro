package cmd

import (
	"github.com/spf13/cobra"
)

var isVerbose bool

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds your Velcro blog",
	Long:  `Builds your Velcro blog into a static site.`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
