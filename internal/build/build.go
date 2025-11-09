package build

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"velcro/internal/siteconfig"
)

type BuildOptions struct {
	RootDir string
}

func Run(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	slog.Info("Building posts...")
	return buildPosts(cfg, opts)
}

func buildPosts(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	postsDir := cfg.Dirs.Posts
	absolutePostsDir := filepath.Join(opts.RootDir, postsDir)

	posts, err := os.ReadDir(absolutePostsDir)
	if err != nil {
		return err
	}

	for _, post := range posts {
		if post.IsDir() {
			// Create the folder in the output directory
			outputPostDir := filepath.Join(opts.RootDir, cfg.OutputDir, "posts", post.Name())
			err := os.MkdirAll(outputPostDir, 0755)
			if err != nil {
				return err
			}

			sourcePostDir := filepath.Join(absolutePostsDir, post.Name())
			err = processDirectory(sourcePostDir, outputPostDir, cfg, opts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processDirectory(src, dst string, cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			err := os.MkdirAll(dstPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(dstPath), 0755)
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".html") {
				err = processHTMLFile(path, dstPath, cfg, opts)
				if err != nil {
					return err
				}
			} else {
				err = copyFile(path, dstPath, info.Mode())
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func processHTMLFile(src, dst string, cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	visited := make(map[string]bool)
	processed, err := processIncludes(string(content), cfg, opts, visited, filepath.Dir(src))
	if err != nil {
		return err
	}

	err = validateHTML(processed, src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, []byte(processed), 0644)
}

func validateHTML(content, filePath string) error {
	headOpenPattern := regexp.MustCompile(`(?i)<head(\s[^>]*)?>`)
	headClosePattern := regexp.MustCompile(`(?i)</head>`)
	bodyOpenPattern := regexp.MustCompile(`(?i)<body(\s[^>]*)?>`)
	bodyClosePattern := regexp.MustCompile(`(?i)</body>`)

	hasHeadOpen := headOpenPattern.MatchString(content)
	hasHeadClose := headClosePattern.MatchString(content)
	hasBodyOpen := bodyOpenPattern.MatchString(content)
	hasBodyClose := bodyClosePattern.MatchString(content)

	if hasHeadOpen && !hasHeadClose {
		slog.Warn("Unclosed <head> tag detected", "file", filePath)
	}

	if hasBodyOpen && !hasBodyClose {
		slog.Warn("Unclosed <body> tag detected", "file", filePath)
	}

	if !hasHeadOpen {
		slog.Warn("Missing <head> tag", "file", filePath)
	}

	return nil
}

func processIncludes(content string, cfg *siteconfig.SiteConfig, opts *BuildOptions, visited map[string]bool, currentDir string) (string, error) {
	includePattern := regexp.MustCompile(`<!--\s*include\s*=\s*"(@[^"]+)"\s*-->`)
	var result strings.Builder
	lastIndex := 0

	for _, match := range includePattern.FindAllStringSubmatchIndex(content, -1) {
		result.WriteString(content[lastIndex:match[0]])

		includePath := content[match[2]:match[3]]

		if includePath == "@content" {
			result.WriteString(content[match[0]:match[1]])
			lastIndex = match[1]
			continue
		}

		if after, ok := strings.CutPrefix(includePath, "@components/"); ok {
			componentName, _ := strings.CutSuffix(after, ".html")

			componentsDir := filepath.Join(opts.RootDir, cfg.Dirs.Components)
			componentHTMLPath := filepath.Join(componentsDir, componentName+".html")

			componentKey := componentHTMLPath
			if visited[componentKey] {
				return "", fmt.Errorf("circular include detected: component %q is included multiple times", componentName)
			}

			visited[componentKey] = true

			componentContent, err := os.ReadFile(componentHTMLPath)
			if err != nil {
				delete(visited, componentKey)
				return "", fmt.Errorf("failed to read component %q: %w", componentName, err)
			}

			processedComponent, err := processIncludes(string(componentContent), cfg, opts, visited, componentsDir)
			if err != nil {
				delete(visited, componentKey)
				return "", err
			}

			delete(visited, componentKey)

			result.WriteString(processedComponent)
			lastIndex = match[1]
		} else {
			result.WriteString(content[match[0]:match[1]])
			lastIndex = match[1]
		}
	}

	result.WriteString(content[lastIndex:])
	return result.String(), nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
