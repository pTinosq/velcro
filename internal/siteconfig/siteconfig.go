package siteconfig

import (
	"log/slog"

	"github.com/BurntSushi/toml"
)

type Dirs struct {
	Root       string `toml:"root"`
	Pages      string `toml:"pages"`
	Posts      string `toml:"posts"`
	Assets     string `toml:"assets"`
	Styles     string `toml:"styles"`
	Scripts    string `toml:"scripts"`
	Components string `toml:"components"`
}

type SiteConfig struct {
	BaseHTML    string `toml:"base_html"`
	OutputDir   string `toml:"output_dir"`
	Dirs        Dirs   `toml:"dirs"`
	DraftPrefix string `toml:"draft_prefix"`
}

func LoadSiteConfig(path string) (*SiteConfig, error) {
	var config SiteConfig

	slog.Info("Loading site config from", "path", path)
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
