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

	slog.Info("Building component assets...")
	err = buildComponentAssets(cfg, opts)
	if err != nil {
		slog.Error("Failed to build component assets", "error", err)
		return err
	}

	slog.Info("Resolving paths...")
	err = resolvePaths(cfg, opts)
	if err != nil {
		slog.Error("Failed to resolve paths", "error", err)
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

func buildComponentAssets(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	componentsDir := filepath.Join(opts.RootDir, cfg.Dirs.Components)

	// Check if components directory exists
	if _, err := os.Stat(componentsDir); os.IsNotExist(err) {
		return nil
	}

	// Read components directory
	components, err := os.ReadDir(componentsDir)
	if err != nil {
		return err
	}

	// Track component names (without extension)
	componentNames := make(map[string]bool)

	for _, component := range components {
		if component.IsDir() {
			continue
		}

		name := component.Name()
		ext := filepath.Ext(name)
		baseName := strings.TrimSuffix(name, ext)

		if ext == ".html" {
			componentNames[baseName] = true
		}
	}

	// Copy component CSS and JS files to output
	outputStylesDir := filepath.Join(opts.RootDir, cfg.OutputDir, "styles")
	outputScriptsDir := filepath.Join(opts.RootDir, cfg.OutputDir, "scripts")

	for componentName := range componentNames {
		// Copy CSS file if it exists
		cssPath := filepath.Join(componentsDir, componentName+".css")
		if _, err := os.Stat(cssPath); err == nil {
			err := os.MkdirAll(outputStylesDir, 0755)
			if err != nil {
				return err
			}
			dstCSSPath := filepath.Join(outputStylesDir, componentName+".css")
			info, err := os.Stat(cssPath)
			if err != nil {
				return err
			}
			err = copyFile(cssPath, dstCSSPath, info.Mode())
			if err != nil {
				return err
			}
		}

		// Copy JS file if it exists
		jsPath := filepath.Join(componentsDir, componentName+".js")
		if _, err := os.Stat(jsPath); err == nil {
			err := os.MkdirAll(outputScriptsDir, 0755)
			if err != nil {
				return err
			}
			dstJSPath := filepath.Join(outputScriptsDir, componentName+".js")
			info, err := os.Stat(jsPath)
			if err != nil {
				return err
			}
			err = copyFile(jsPath, dstJSPath, info.Mode())
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

	// Extract page/post identifier for data-page processing
	var currentPageID string
	if isFromPostsOrPages {
		// Extract the page/post name from the source path
		if relPath, err := filepath.Rel(absolutePostsDir, src); err == nil && !strings.HasPrefix(relPath, "..") {
			// It's a post - get the post folder name
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > 0 {
				currentPageID = parts[0]
			}
		} else if relPath, err := filepath.Rel(absolutePagesDir, src); err == nil && !strings.HasPrefix(relPath, "..") {
			// It's a page - get the page folder name
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > 0 {
				currentPageID = parts[0]
			}
		}
	}

	visited := make(map[string]bool)
	componentAssets := make(map[string]bool) // Track component CSS/JS files
	processed, err := processIncludes(string(content), cfg, opts, visited, filepath.Dir(src), currentPageID, componentAssets)
	if err != nil {
		return err
	}

	// Inject component CSS and JS files into the HTML
	processed = injectComponentAssets(processed, componentAssets, cfg, opts)

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

func processIncludes(content string, cfg *siteconfig.SiteConfig, opts *BuildOptions, visited map[string]bool, currentDir string, currentPageID string, componentAssets map[string]bool) (string, error) {
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

			// Check for associated CSS and JS files
			componentCSSPath := filepath.Join(componentsDir, componentName+".css")
			componentJSPath := filepath.Join(componentsDir, componentName+".js")

			if _, err := os.Stat(componentCSSPath); err == nil {
				// CSS file exists, track it for injection
				componentAssets["css:"+componentName] = true
			}

			if _, err := os.Stat(componentJSPath); err == nil {
				// JS file exists, track it for injection
				componentAssets["js:"+componentName] = true
			}

			processedComponent, err := processIncludes(string(componentContent), cfg, opts, visited, componentsDir, currentPageID, componentAssets)
			if err != nil {
				delete(visited, componentKey)
				return "", err
			}

			// Process data-page attributes in the component
			processedComponent = processDataPageAttributes(processedComponent, currentPageID)

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

func injectComponentAssets(content string, componentAssets map[string]bool, cfg *siteconfig.SiteConfig, opts *BuildOptions) string {
	if len(componentAssets) == 0 {
		return content
	}

	// Collect CSS and JS links
	var cssLinks []string
	var jsLinks []string

	for asset := range componentAssets {
		if after, ok := strings.CutPrefix(asset, "css:"); ok {
			componentName := after
			// Use @styles path that will be resolved later
			cssLinks = append(cssLinks, `<link rel="stylesheet" href="@styles/`+componentName+`.css">`)
		} else if after0, ok0 := strings.CutPrefix(asset, "js:"); ok0 {
			componentName := after0
			// Use @scripts path that will be resolved later
			jsLinks = append(jsLinks, `<script src="@scripts/`+componentName+`.js"></script>`)
		}
	}

	// Inject CSS into <head> (before </head>)
	if len(cssLinks) > 0 {
		headPattern := regexp.MustCompile(`(?i)(</head>)`)
		cssInjection := "    " + strings.Join(cssLinks, "\n    ") + "\n"
		content = headPattern.ReplaceAllString(content, cssInjection+"$1")
	}

	// Inject JS before </body>
	if len(jsLinks) > 0 {
		bodyPattern := regexp.MustCompile(`(?i)(</body>)`)
		jsInjection := "    " + strings.Join(jsLinks, "\n    ") + "\n"
		content = bodyPattern.ReplaceAllString(content, jsInjection+"$1")
	}

	return content
}

func processDataPageAttributes(content string, currentPageID string) string {
	// Pattern to match tags with data-page attribute
	// Matches: <tag ... data-page="value" ...> or <tag data-page="value" ...>
	dataPagePattern := regexp.MustCompile(`(<[^>]*\s+)data-page\s*=\s*"([^"]*)"([^>]*>)`)

	return dataPagePattern.ReplaceAllStringFunc(content, func(match string) string {
		submatch := dataPagePattern.FindStringSubmatch(match)
		if len(submatch) >= 4 {
			beforeAttr := submatch[1]
			dataPageValue := submatch[2]
			afterAttr := submatch[3]

			// Check if data-page value matches current page ID
			if currentPageID != "" && dataPageValue == currentPageID {
				// Check if class attribute already exists in the tag
				fullTag := beforeAttr + afterAttr
				classPattern := regexp.MustCompile(`class\s*=\s*"([^"]*)"`)
				if classMatch := classPattern.FindString(fullTag); classMatch != "" {
					// Extract existing class value
					classSubmatch := classPattern.FindStringSubmatch(fullTag)
					if len(classSubmatch) > 1 {
						existingClass := classSubmatch[1]
						if !strings.Contains(existingClass, "active") {
							newClass := existingClass + " active"
							// Replace the class attribute
							result := classPattern.ReplaceAllString(fullTag, `class="`+newClass+`"`)
							return result
						}
						// active already exists, just remove data-page
						return fullTag
					}
				}
				// No class attribute, add it before the closing >
				return beforeAttr + `class="active"` + afterAttr
			}
			// Doesn't match or no currentPageID, just remove data-page attribute
			return beforeAttr + afterAttr
		}
		return match
	})
}

func resolvePaths(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	outputDir := filepath.Join(opts.RootDir, cfg.OutputDir)

	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process .html, .css, and .js files
		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".css" && ext != ".js" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Get the directory of the current file relative to outputDir
		relPath, err := filepath.Rel(outputDir, path)
		if err != nil {
			return err
		}
		currentDir := filepath.Dir(relPath)
		if currentDir == "." {
			currentDir = ""
		}

		contentStr := string(content)
		modified := false

		// Helper function to calculate relative path from current file to target
		calculateRelativePath := func(targetPath string) string {
			// If currentDir is empty (file is at root), targetPath is already relative
			if currentDir == "" {
				return targetPath
			}
			// Calculate relative path from current file's directory to target
			rel, err := filepath.Rel(currentDir, targetPath)
			if err != nil {
				return targetPath
			}
			// Normalize path separators for web (use forward slashes)
			return filepath.ToSlash(rel)
		}

		// Resolve @assets paths
		assetsPattern := regexp.MustCompile(`@assets(/[^"'\s>)]*)`)
		if assetsPattern.MatchString(contentStr) {
			contentStr = assetsPattern.ReplaceAllStringFunc(contentStr, func(match string) string {
				submatch := assetsPattern.FindStringSubmatch(match)
				if len(submatch) > 1 {
					targetPath := "assets" + submatch[1]
					return calculateRelativePath(targetPath)
				}
				return match
			})
			modified = true
		}

		// Resolve @posts paths
		postsPattern := regexp.MustCompile(`@posts(/[^"'\s>)]*)`)
		if postsPattern.MatchString(contentStr) {
			contentStr = postsPattern.ReplaceAllStringFunc(contentStr, func(match string) string {
				submatch := postsPattern.FindStringSubmatch(match)
				if len(submatch) > 1 {
					targetPath := "posts" + submatch[1]
					return calculateRelativePath(targetPath)
				}
				return match
			})
			modified = true
		}

		// Resolve @styles paths
		stylesPattern := regexp.MustCompile(`@styles(/[^"'\s>)]*)`)
		if stylesPattern.MatchString(contentStr) {
			contentStr = stylesPattern.ReplaceAllStringFunc(contentStr, func(match string) string {
				submatch := stylesPattern.FindStringSubmatch(match)
				if len(submatch) > 1 {
					targetPath := "styles" + submatch[1]
					return calculateRelativePath(targetPath)
				}
				return match
			})
			modified = true
		}

		// Resolve @scripts paths
		scriptsPattern := regexp.MustCompile(`@scripts(/[^"'\s>)]*)`)
		if scriptsPattern.MatchString(contentStr) {
			contentStr = scriptsPattern.ReplaceAllStringFunc(contentStr, func(match string) string {
				submatch := scriptsPattern.FindStringSubmatch(match)
				if len(submatch) > 1 {
					targetPath := "scripts" + submatch[1]
					return calculateRelativePath(targetPath)
				}
				return match
			})
			modified = true
		}

		// Resolve @pages paths (special handling: pages/index -> root, pages/{other} -> {other})
		pagesPattern := regexp.MustCompile(`@pages(/[^"'\s>)]*)`)
		if pagesPattern.MatchString(contentStr) {
			contentStr = pagesPattern.ReplaceAllStringFunc(contentStr, func(match string) string {
				submatch := pagesPattern.FindStringSubmatch(match)
				if len(submatch) > 1 {
					pagePath := submatch[1]
					pagePath = strings.TrimPrefix(pagePath, "/")
					parts := strings.Split(pagePath, "/")

					var targetPath string
					if len(parts) > 0 && parts[0] == "index" {
						// pages/index/... -> root level
						if len(parts) > 1 {
							targetPath = strings.Join(parts[1:], "/")
						} else {
							targetPath = "index.html"
						}
					} else if len(parts) > 0 {
						// pages/{other}/... -> {other}/...
						targetPath = pagePath
					} else {
						return match
					}

					return calculateRelativePath(targetPath)
				}
				return match
			})
			modified = true
		}

		if modified {
			return os.WriteFile(path, []byte(contentStr), info.Mode())
		}

		return nil
	})
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
