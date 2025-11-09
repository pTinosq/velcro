package cmd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed init_template
var templateFS embed.FS

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Velcro blog",
	Long:  `Initializes a new Velcro blog with a default template.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			slog.Error("Please provide a name for your blog")
			return
		}
		blogName := args[0]

		slog.Debug("Validating blog name", "blogName", blogName)
		match, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", blogName)
		if !match {
			slog.Error("The blog name must contain only A-Z, a-z, 0-9, hyphens, and underscores")
			return
		}

		slog.Info("⚙️ Initializing your Velcro blog...")

		slog.Debug("Checking if blog directory exists", "blogName", blogName)
		if _, err := os.Stat(fmt.Sprintf("./%s", blogName)); err == nil {
			slog.Error("A folder with this name already exists")
			return
		}

		slog.Debug("Creating blog directory", "blogName", blogName)
		err := os.MkdirAll(fmt.Sprintf("./%s", blogName), 0755)
		if err != nil {
			slog.Error("Failed to create blog directory", "error", err)
			return
		}
		slog.Debug("Copying template files", "blogName", blogName)
		err = copyTemplateFiles(blogName)
		if err != nil {
			slog.Error("Failed to copy template files", "error", err)
			return
		}

		slog.Info("✅ Blog initialized successfully!\n")

		slog.Info("👀 Getting started with your new blog")
		slog.Info(fmt.Sprintf("1. cd ./%s", blogName))
		slog.Info("2. velcro build")
		slog.Info("3. velcro serve")
	},
}

func copyTemplateFiles(blogName string) error {
	destDir := filepath.Join(".", blogName)
	return fs.WalkDir(templateFS, "init_template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(d.Name(), ".gitkeep") {
			return nil
		}

		relPath, err := filepath.Rel("init_template", path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			slog.Debug("Creating directory", "path", destPath)
			return os.MkdirAll(destPath, 0755)
		}

		slog.Debug("Copying file", "from", path, "to", destPath)
		srcFile, err := templateFS.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func init() {
	rootCmd.AddCommand(initCmd)
}
