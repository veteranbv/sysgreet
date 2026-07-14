package config

const SchemaVersion = "v1"

// DisplayConfig controls section visibility in the banner.
type DisplayConfig struct {
	Hostname    bool `yaml:"hostname" toml:"hostname"`
	OS          bool `yaml:"os" toml:"os"`
	IPAddresses bool `yaml:"ip_addresses" toml:"ip_addresses"`
	RemoteIP    bool `yaml:"remote_ip" toml:"remote_ip"`
	Uptime      bool `yaml:"uptime" toml:"uptime"`
	User        bool `yaml:"user" toml:"user"`
	Memory      bool `yaml:"memory" toml:"memory"`
	Disk        bool `yaml:"disk" toml:"disk"`
	Load        bool `yaml:"load" toml:"load"`
	Datetime    bool `yaml:"datetime" toml:"datetime"`
	LastLogin   bool `yaml:"last_login" toml:"last_login"`
}

// ASCIIConfig controls hostname ASCII-art rendering.
type ASCIIConfig struct {
	Font       string   `yaml:"font" toml:"font"`
	Color      string   `yaml:"color" toml:"color"`
	Gradient   []string `yaml:"gradient,omitempty" toml:"gradient,omitempty"`
	Monochrome bool     `yaml:"monochrome" toml:"monochrome"`
}

// LayoutConfig controls banner layout options.
type LayoutConfig struct {
	Compact bool `yaml:"compact" toml:"compact"`
	// MaxWidth caps the banner width in columns. Zero means use the
	// detected terminal width.
	MaxWidth int      `yaml:"max_width" toml:"max_width"`
	Sections []string `yaml:"sections" toml:"sections"`
}

// NetworkConfig configures network address display.
type NetworkConfig struct {
	ShowInterfaceNames bool `yaml:"show_interface_names" toml:"show_interface_names"`
	MaxInterfaces      int  `yaml:"max_interfaces" toml:"max_interfaces"`
}

// Config is the root configuration for sysgreet.
type Config struct {
	Display   DisplayConfig `yaml:"display" toml:"display"`
	ASCII     ASCIIConfig   `yaml:"ascii" toml:"ascii"`
	Layout    LayoutConfig  `yaml:"layout" toml:"layout"`
	Network   NetworkConfig `yaml:"network" toml:"network"`
	Version   string        `yaml:"version" toml:"version"`
	CreatedAt string        `yaml:"created_at" toml:"created_at"`
}

// Default returns the default configuration used when no config file is present.
func Default() Config {
	return Config{
		Display: DisplayConfig{
			Hostname:    true,
			OS:          true,
			IPAddresses: true,
			RemoteIP:    true,
			Uptime:      true,
			User:        true,
			Memory:      true,
			Disk:        true,
			Load:        true,
			Datetime:    true,
			LastLogin:   true,
		},
		ASCII: ASCIIConfig{
			Font:       "ANSI Regular",
			Color:      "",
			Gradient:   []string{"brightblue", "blue", "cyan", "brightcyan", "white"},
			Monochrome: false,
		},
		Layout: LayoutConfig{
			Compact:  false,
			Sections: []string{"header", "system", "network", "resources"},
		},
		Network: NetworkConfig{
			ShowInterfaceNames: true,
			MaxInterfaces:      3,
		},
		Version:   SchemaVersion,
		CreatedAt: "",
	}
}
