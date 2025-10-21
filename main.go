package main

import (
	"log/slog"
	"os"
	"velcro/cmd"

	"github.com/lmittmann/tint"
)

func main() {
	w := os.Stderr
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level: slog.LevelInfo,
			// Remove the time attribute from the output
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
		}),
	))

	cmd.Execute()
}
