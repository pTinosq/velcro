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
	err := buildPosts(cfg, opts)
	if err != nil {
		slog.Error("Failed to build posts", "error", err)
		return err
	}

	slog.Info("Building assets...")
	err = buildAssets(cfg, opts)
	if err != nil {
		slog.Error("Failed to build assets", "error", err)
		return err
	}

	slog.Info("Building scripts...")
	err = buildScripts(cfg, opts)
	if err != nil {
		slog.Error("Failed to build scripts", "error", err)
		return err
	}

	slog.Info("Building styles...")
	err = buildStyles(cfg, opts)
	if err != nil {
		slog.Error("Failed to build styles", "error", err)
		return err
	}

	slog.Info("Building pages...")
	err = buildPages(cfg, opts)
	if err != nil {
		slog.Error("Failed to build pages", "error", err)
		return err
	}

	return nil
}

func buildAssets(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	assetsDir := cfg.Dirs.Assets
	absoluteAssetsDir := filepath.Join(opts.RootDir, assetsDir)

	if _, err := os.Stat(absoluteAssetsDir); os.IsNotExist(err) {
		return nil
	}

	// Create output assets directory
	outputAssetsDir := filepath.Join(opts.RootDir, cfg.OutputDir, "assets")
	err := os.MkdirAll(outputAssetsDir, 0755)
	if err != nil {
		return err
	}

	return processDirectory(absoluteAssetsDir, outputAssetsDir, cfg, opts)
}

func buildScripts(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	scriptsDir := cfg.Dirs.Scripts
	absoluteScriptsDir := filepath.Join(opts.RootDir, scriptsDir)

	if _, err := os.Stat(absoluteScriptsDir); os.IsNotExist(err) {
		return nil
	}

	outputScriptsDir := filepath.Join(opts.RootDir, cfg.OutputDir, "scripts")
	err := os.MkdirAll(outputScriptsDir, 0755)
	if err != nil {
		return err
	}

	return processDirectory(absoluteScriptsDir, outputScriptsDir, cfg, opts)
}

func buildStyles(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	stylesDir := cfg.Dirs.Styles
	absoluteStylesDir := filepath.Join(opts.RootDir, stylesDir)

	// Check if styles directory exists
	if _, err := os.Stat(absoluteStylesDir); os.IsNotExist(err) {
		// Styles directory doesn't exist, nothing to copy
		return nil
	}

	// Create output styles directory
	outputStylesDir := filepath.Join(opts.RootDir, cfg.OutputDir, "styles")
	err := os.MkdirAll(outputStylesDir, 0755)
	if err != nil {
		return err
	}

	// Copy all styles using processDirectory
	return processDirectory(absoluteStylesDir, outputStylesDir, cfg, opts)
}

func buildPages(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	pagesDir := cfg.Dirs.Pages
	absolutePagesDir := filepath.Join(opts.RootDir, pagesDir)

	pages, err := os.ReadDir(absolutePagesDir)
	if err != nil {
		return err
	}

	for _, page := range pages {
		if page.IsDir() {
			sourcePageDir := filepath.Join(absolutePagesDir, page.Name())

			var outputPageDir string
			if page.Name() == "index" {
				// index page goes to the root of the output directory
				outputPageDir = filepath.Join(opts.RootDir, cfg.OutputDir)
			} else {
				// other pages go to {outputDir}/{pageName}
				outputPageDir = filepath.Join(opts.RootDir, cfg.OutputDir, page.Name())
			}

			err := os.MkdirAll(outputPageDir, 0755)
			if err != nil {
				return err
			}

			err = processDirectory(sourcePageDir, outputPageDir, cfg, opts)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

	// Check if this HTML file is from posts or pages directory
	absolutePostsDir := filepath.Join(opts.RootDir, cfg.Dirs.Posts)
	absolutePagesDir := filepath.Join(opts.RootDir, cfg.Dirs.Pages)

	isFromPostsOrPages := false
	if relPath, err := filepath.Rel(absolutePostsDir, src); err == nil && !strings.HasPrefix(relPath, "..") {
		isFromPostsOrPages = true
	}
	if !isFromPostsOrPages {
		if relPath, err := filepath.Rel(absolutePagesDir, src); err == nil && !strings.HasPrefix(relPath, "..") {
			isFromPostsOrPages = true
		}
	}

	// If from posts or pages, merge with base.html
	if isFromPostsOrPages {
		baseHTMLPath := filepath.Join(opts.RootDir, cfg.BaseHTML)
		baseContent, err := os.ReadFile(baseHTMLPath)
		if err != nil {
			return fmt.Errorf("failed to read base.html: %w", err)
		}

		pageContent := string(content)
		baseHTML := string(baseContent)

		// Extract <head> content from page/post (everything between <head> and </head>)
		headPattern := regexp.MustCompile(`(?i)<head(\s[^>]*)?>([\s\S]*?)</head>`)
		headMatches := headPattern.FindStringSubmatch(pageContent)
		var pageHeadContent string
		if len(headMatches) > 2 {
			pageHeadContent = headMatches[2]
		}

		// Extract <body> content from page/post (everything between <body> and </body>)
		bodyPattern := regexp.MustCompile(`(?i)<body(\s[^>]*)?>([\s\S]*?)</body>`)
		bodyMatches := bodyPattern.FindStringSubmatch(pageContent)
		var pageBodyContent string
		if len(bodyMatches) > 2 {
			pageBodyContent = bodyMatches[2]
		}

		// Merge head content into base.html's head
		if pageHeadContent != "" {
			baseHeadPattern := regexp.MustCompile(`(?i)(<head(\s[^>]*)?>)([\s\S]*?)(</head>)`)
			baseHTML = baseHeadPattern.ReplaceAllString(baseHTML, "${1}${3}"+pageHeadContent+"${4}")
		}

		// Merge body content into base.html's body (replace @content placeholder)
		if pageBodyContent != "" {
			contentIncludePattern := regexp.MustCompile(`<!--\s*include\s*=\s*"@content"\s*-->`)
			if !contentIncludePattern.MatchString(baseHTML) {
				return fmt.Errorf("base.html does not contain @content placeholder")
			}
			baseHTML = contentIncludePattern.ReplaceAllString(baseHTML, pageBodyContent)
		}

		content = []byte(baseHTML)
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
			slog.Debug("Processing component", "component", after)
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
