package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var (
	errUnsupportedFormat = errors.New("unsupported config format")
)

// Load returns the merged configuration, the path that was used, and an error if loading fails.
// Defaults are always applied; missing files are ignored.
func Load() (Config, string, error) {
	cfg := Default()
	candidatePaths := defaultConfigPaths()

	if custom := os.Getenv("SYSGREET_CONFIG"); custom != "" {
		candidatePaths = append([]string{custom}, candidatePaths...)
	}

	var usedPath string
	for _, p := range candidatePaths {
		if p == "" {
			continue
		}
		expanded := expandPath(p)
		info, err := os.Stat(expanded)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		data, err := os.ReadFile(expanded)
		if err != nil {
			return Config{}, "", fmt.Errorf("read config: %w", err)
		}

		var raw rawConfig
		switch strings.ToLower(filepath.Ext(expanded)) {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &raw); err != nil {
				return Config{}, "", fmt.Errorf("parse yaml config: %w", err)
			}
		case ".toml":
			if err := toml.Unmarshal(data, &raw); err != nil {
				return Config{}, "", fmt.Errorf("parse toml config: %w", err)
			}
		default:
			return Config{}, "", fmt.Errorf("%w: %s", errUnsupportedFormat, expanded)
		}

		mergeConfig(&cfg, raw)
		usedPath = expanded
		break
	}

	applyEnvOverrides(&cfg)
	return cfg, usedPath, nil
}

func defaultConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "sysgreet", "config.yaml"),
		filepath.Join(home, ".config", "sysgreet", "config.yml"),
		filepath.Join(home, ".config", "sysgreet", "config.toml"),
		filepath.Join(home, ".sysgreet.yaml"),
		filepath.Join(home, ".sysgreet.yml"),
		filepath.Join(home, ".sysgreet.toml"),
	}
}

func DefaultWritePath() string {
	if custom := os.Getenv("SYSGREET_CONFIG"); custom != "" {
		return expandPath(custom)
	}
	for _, candidate := range defaultConfigPaths() {
		if candidate == "" {
			continue
		}
		return expandPath(candidate)
	}
	return ""
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return os.ExpandEnv(p)
}

func mergeConfig(base *Config, override rawConfig) {
	mergeDisplay(base, override.Display)
	mergeASCII(base, override.ASCII)
	mergeLayout(base, override.Layout)
	mergeNetwork(base, override.Network)
	mergeMetadata(base, override.Version, override.CreatedAt)
}

func mergeDisplay(base *Config, display *rawDisplay) {
	if display == nil {
		return
	}
	if display.Hostname != nil {
		base.Display.Hostname = *display.Hostname
	}
	if display.OS != nil {
		base.Display.OS = *display.OS
	}
	if display.IPAddresses != nil {
		base.Display.IPAddresses = *display.IPAddresses
	}
	if display.RemoteIP != nil {
		base.Display.RemoteIP = *display.RemoteIP
	}
	if display.Uptime != nil {
		base.Display.Uptime = *display.Uptime
	}
	if display.User != nil {
		base.Display.User = *display.User
	}
	if display.Memory != nil {
		base.Display.Memory = *display.Memory
	}
	if display.Disk != nil {
		base.Display.Disk = *display.Disk
	}
	if display.Load != nil {
		base.Display.Load = *display.Load
	}
	if display.Datetime != nil {
		base.Display.Datetime = *display.Datetime
	}
	if display.LastLogin != nil {
		base.Display.LastLogin = *display.LastLogin
	}
}

func mergeASCII(base *Config, ascii *rawASCII) {
	if ascii == nil {
		return
	}
	if ascii.Font != nil && *ascii.Font != "" {
		base.ASCII.Font = *ascii.Font
	}
	if ascii.Color != nil && *ascii.Color != "" {
		base.ASCII.Color = *ascii.Color
	}
	if ascii.Gradient != nil && len(*ascii.Gradient) > 0 {
		base.ASCII.Gradient = append([]string{}, (*ascii.Gradient)...)
	}
	if ascii.Monochrome != nil {
		base.ASCII.Monochrome = *ascii.Monochrome
	}
}

func mergeLayout(base *Config, layout *rawLayout) {
	if layout == nil {
		return
	}
	if layout.Sections != nil && len(*layout.Sections) > 0 {
		base.Layout.Sections = append([]string{}, (*layout.Sections)...)
	}
	if layout.Compact != nil {
		base.Layout.Compact = *layout.Compact
	}
	if layout.MaxWidth != nil && *layout.MaxWidth >= 0 {
		base.Layout.MaxWidth = *layout.MaxWidth
	}
}

func mergeNetwork(base *Config, network *rawNetwork) {
	if network == nil {
		return
	}
	if network.ShowInterfaceNames != nil {
		base.Network.ShowInterfaceNames = *network.ShowInterfaceNames
	}
	if network.MaxInterfaces != nil {
		base.Network.MaxInterfaces = *network.MaxInterfaces
	}
}

func mergeMetadata(base *Config, version, createdAt *string) {
	if version != nil && *version != "" {
		base.Version = *version
	}
	if createdAt != nil && *createdAt != "" {
		base.CreatedAt = *createdAt
	}
}

func applyEnvOverrides(cfg *Config) {
	applyDisplayOverrides(&cfg.Display)
	applyASCIIOverrides(&cfg.ASCII)
	applyLayoutOverrides(&cfg.Layout)
	applyNetworkOverrides(&cfg.Network)
}

func applyDisplayOverrides(display *DisplayConfig) {
	boolFields := map[string]*bool{
		"SYSGREET_DISPLAY_HOSTNAME":     &display.Hostname,
		"SYSGREET_DISPLAY_OS":           &display.OS,
		"SYSGREET_DISPLAY_IP_ADDRESSES": &display.IPAddresses,
		"SYSGREET_DISPLAY_REMOTE_IP":    &display.RemoteIP,
		"SYSGREET_DISPLAY_UPTIME":       &display.Uptime,
		"SYSGREET_DISPLAY_USER":         &display.User,
		"SYSGREET_DISPLAY_MEMORY":       &display.Memory,
		"SYSGREET_DISPLAY_DISK":         &display.Disk,
		"SYSGREET_DISPLAY_LOAD":         &display.Load,
		"SYSGREET_DISPLAY_DATETIME":     &display.Datetime,
		"SYSGREET_DISPLAY_LAST_LOGIN":   &display.LastLogin,
	}
	for key, field := range boolFields {
		if v, ok := lookupBool(key); ok {
			*field = v
		}
	}
}

func applyASCIIOverrides(ascii *ASCIIConfig) {
	if font := os.Getenv("SYSGREET_ASCII_FONT"); font != "" {
		ascii.Font = font
	}
	if color := os.Getenv("SYSGREET_ASCII_COLOR"); color != "" {
		ascii.Color = color
	}
	if v, ok := lookupBool("SYSGREET_ASCII_MONOCHROME"); ok {
		ascii.Monochrome = v
	}
}

func applyLayoutOverrides(layout *LayoutConfig) {
	if v, ok := lookupBool("SYSGREET_LAYOUT_COMPACT"); ok {
		layout.Compact = v
	}
	if max := os.Getenv("SYSGREET_LAYOUT_MAX_WIDTH"); max != "" {
		if parsed, err := parseInt(max); err == nil && parsed >= 0 {
			layout.MaxWidth = parsed
		}
	}
	if sections := os.Getenv("SYSGREET_LAYOUT_SECTIONS"); sections != "" {
		parts := strings.Split(sections, ",")
		var cleaned []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			layout.Sections = cleaned
		}
	}
}

func applyNetworkOverrides(network *NetworkConfig) {
	if v, ok := lookupBool("SYSGREET_NETWORK_SHOW_INTERFACE_NAMES"); ok {
		network.ShowInterfaceNames = v
	}
	if max := os.Getenv("SYSGREET_NETWORK_MAX_INTERFACES"); max != "" {
		if parsed, err := parseInt(max); err == nil {
			network.MaxInterfaces = parsed
		}
	}
}

func lookupBool(key string) (bool, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, true
	}
}

func parseInt(input string) (int, error) {
	var value int
	_, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &value)
	return value, err
}

type rawConfig struct {
	Display   *rawDisplay `yaml:"display" toml:"display"`
	ASCII     *rawASCII   `yaml:"ascii" toml:"ascii"`
	Layout    *rawLayout  `yaml:"layout" toml:"layout"`
	Network   *rawNetwork `yaml:"network" toml:"network"`
	Version   *string     `yaml:"version" toml:"version"`
	CreatedAt *string     `yaml:"created_at" toml:"created_at"`
}

type rawDisplay struct {
	Hostname    *bool `yaml:"hostname" toml:"hostname"`
	OS          *bool `yaml:"os" toml:"os"`
	IPAddresses *bool `yaml:"ip_addresses" toml:"ip_addresses"`
	RemoteIP    *bool `yaml:"remote_ip" toml:"remote_ip"`
	Uptime      *bool `yaml:"uptime" toml:"uptime"`
	User        *bool `yaml:"user" toml:"user"`
	Memory      *bool `yaml:"memory" toml:"memory"`
	Disk        *bool `yaml:"disk" toml:"disk"`
	Load        *bool `yaml:"load" toml:"load"`
	Datetime    *bool `yaml:"datetime" toml:"datetime"`
	LastLogin   *bool `yaml:"last_login" toml:"last_login"`
}

type rawASCII struct {
	Font       *string   `yaml:"font" toml:"font"`
	Color      *string   `yaml:"color" toml:"color"`
	Gradient   *[]string `yaml:"gradient,omitempty" toml:"gradient,omitempty"`
	Monochrome *bool     `yaml:"monochrome" toml:"monochrome"`
}

type rawLayout struct {
	Compact  *bool     `yaml:"compact" toml:"compact"`
	MaxWidth *int      `yaml:"max_width" toml:"max_width"`
	Sections *[]string `yaml:"sections" toml:"sections"`
}

type rawNetwork struct {
	ShowInterfaceNames *bool `yaml:"show_interface_names" toml:"show_interface_names"`
	MaxInterfaces      *int  `yaml:"max_interfaces" toml:"max_interfaces"`
}
