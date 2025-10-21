package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Velcro blog",
	Long:  `Initializes a new Velcro blog with a default template.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("⚙️ Initializing your Velcro blog...")
		slog.Info("⚙️ Initializing your Velcro blog...")

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
