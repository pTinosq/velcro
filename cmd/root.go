package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var isVerbose bool

var rootCmd = &cobra.Command{
	Use:   "velcro",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logger with correct level based on verbose flag
		w := os.Stderr
		var level slog.Level
		if isVerbose {
			level = slog.LevelDebug
		} else {
			level = slog.LevelInfo
		}

		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level: level,
				// Remove the time attribute from the output
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey && len(groups) == 0 {
						return slog.Attr{}
					}
					return a
				},
			}),
		))
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVarP(&isVerbose, "verbose", "v", false, "verbose output")
}
