package bootstrap

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/veteranbv/sysgreet/internal/config"
)

// renderDefaultConfig marshals the default sysgreet configuration with
// metadata, in the format implied by the target path's extension so that a
// bootstrapped .toml config parses as TOML.
func renderDefaultConfig(now time.Time, path string) ([]byte, error) {
	cfg := config.Default()
	cfg.Version = config.SchemaVersion
	cfg.CreatedAt = now.UTC().Format(time.RFC3339)
	if strings.EqualFold(filepath.Ext(path), ".toml") {
		return toml.Marshal(cfg)
	}
	return yaml.Marshal(cfg)
}
