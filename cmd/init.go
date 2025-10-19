package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize something",
	Long:  `Initialize command that logs hello world.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing Velcro...")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
