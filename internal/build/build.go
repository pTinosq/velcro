package build

import (
	"log/slog"
	"os"
	"path/filepath"
	"velcro/internal/siteconfig"
)

type BuildOptions struct {
	RootDir string
}

func Run(cfg *siteconfig.SiteConfig, opts *BuildOptions) error {
	// 1. build the posts
	slog.Info("Building posts...")
	err := buildPosts(cfg, opts)
	if err != nil {
		return err
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
			slog.Info("Building post", "post", post.Name())
		}
	}

	return nil
}
