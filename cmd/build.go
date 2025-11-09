package cmd

import (
	"log/slog"
	"path/filepath"
	"velcro/internal/siteconfig"

	"github.com/spf13/cobra"
)

// var isVerbose bool

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds your Velcro blog",
	Long:  `Builds your Velcro blog into a static site.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			slog.Error("Please provide a path to your site config")
			return
		}
		siteConfigPath := args[0]
		siteConfigPath = filepath.Join(siteConfigPath, "site.config.toml")

		config, err := siteconfig.LoadSiteConfig(siteConfigPath)
		if err != nil {
			slog.Error("Failed to load site config", "error", err)
			return
		}
		slog.Info("Site config loaded successfully", "config", config)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
