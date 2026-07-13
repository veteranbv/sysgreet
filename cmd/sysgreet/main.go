package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/veteranbv/sysgreet/internal/ascii"
	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/bootstrap"
	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/render"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, bootstrap.ErrUserCanceled) {
			return
		}
		fmt.Fprintf(os.Stderr, "sysgreet: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	settings := parseFlags()

	if settings.Version {
		fmt.Printf("sysgreet %s (commit: %s, built: %s)\n", version, commit, date)
		return nil
	}
	if settings.Disable {
		return nil
	}
	if settings.ConfigPath != "" {
		// Reuse the SYSGREET_CONFIG plumbing so load and bootstrap agree
		// on the path.
		if err := os.Setenv("SYSGREET_CONFIG", settings.ConfigPath); err != nil {
			return err
		}
	}

	renderer, err := ascii.NewRenderer()
	if err != nil {
		return err
	}

	if settings.ListFonts {
		for _, font := range renderer.Fonts() {
			fmt.Println(font)
		}
		return nil
	}

	// Legacy Windows consoles need virtual terminal processing switched on
	// before any escape sequences are written; when that fails, fall back
	// to plain output rather than printing raw escapes.
	ansiOK := enableVirtualTerminal(os.Stdout)
	env := terminal.DetectEnv(os.Stdout, settings.NoColor || !ansiOK)

	cfg, err := loadConfig(ctx, settings)
	if err != nil {
		return err
	}
	env = render.ApplyConfig(env, cfg)
	if settings.Width > 0 {
		// The flag wins over both the detected width and layout.max_width.
		env.Width = settings.Width
	}

	if settings.Text != "" {
		return runTextMode(renderer, settings.Text, cfg, env)
	}

	buildEnv := env
	if settings.JSON {
		// Scripted output must not vary with terminal geometry or color
		// support; build against a neutral environment.
		buildEnv = terminal.Env{}
	}
	output, err := buildBanner(ctx, renderer, cfg, buildEnv, settings.Demo)
	if err != nil {
		return err
	}
	return printBanner(output, cfg, env, settings.JSON)
}

// loadConfig bootstraps a config on first run and loads it. --text, --demo,
// and --json are one-shot or scripted invocations that must never prompt,
// so bootstrap only runs in normal mode.
func loadConfig(ctx context.Context, settings runSettings) (config.Config, error) {
	if settings.Text == "" && !settings.Demo && !settings.JSON {
		if err := maybeBootstrap(ctx, settings); err != nil {
			return config.Config{}, err
		}
	}
	cfg, _, err := config.Load()
	if err != nil {
		return config.Config{}, err
	}
	if settings.Font != "" {
		cfg.ASCII.Font = settings.Font
	}
	return cfg, nil
}

func printBanner(output banner.Output, cfg config.Config, env terminal.Env, asJSON bool) error {
	if asJSON {
		doc, err := render.RenderJSON(output, cfg)
		if err != nil {
			return err
		}
		fmt.Println(doc)
		return nil
	}
	fmt.Println(render.NewRenderer(env).Render(output, cfg))
	return nil
}

type runSettings struct {
	PolicyFlag string
	ConfigPath string
	Font       string
	Width      int
	Disable    bool
	Demo       bool
	JSON       bool
	ListFonts  bool
	NoColor    bool
	Text       string
	Version    bool
}

func parseFlags() runSettings {
	policyFlag := flag.String("config-policy", "", "Config bootstrap policy: prompt, keep, or overwrite")
	configPath := flag.String("config", "", "Path to a config file (overrides default lookup)")
	font := flag.String("font", "", "Font override for this run (see --list-fonts)")
	width := flag.Int("width", 0, "Assume this terminal width instead of detecting it")
	disable := flag.Bool("disable", false, "Disable sysgreet output")
	demo := flag.Bool("demo", false, "Demo mode with 'SYSGREET' banner and fake data")
	jsonOut := flag.Bool("json", false, "Emit the banner as JSON for scripting")
	listFonts := flag.Bool("list-fonts", false, "List embedded fonts and exit")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	text := flag.String("text", "", "Render custom text as ASCII art (e.g., --text \"Tea Pot\")")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "\nEnvironment variables:")
		fmt.Fprintln(flag.CommandLine.Output(), "  SYSGREET_CONFIG          Config file path (same as --config)")
		fmt.Fprintln(flag.CommandLine.Output(), "  SYSGREET_CONFIG_POLICY   Config bootstrap policy (prompt|keep|overwrite)")
		fmt.Fprintln(flag.CommandLine.Output(), "  SYSGREET_ASSUME_TTY      Force interactive prompts (testing/support)")
		fmt.Fprintln(flag.CommandLine.Output(), "  NO_COLOR                 Disable colored output (same as --no-color)")
		fmt.Fprintln(flag.CommandLine.Output(), "  CI                      When set, disables interactive prompts by default")
		fmt.Fprintln(flag.CommandLine.Output(), "\nBootstrap:")
		fmt.Fprintln(flag.CommandLine.Output(), "  First run writes curated defaults (ANSI Regular font with gradient, metadata).")
		fmt.Fprintln(flag.CommandLine.Output(), "  Existing configs stay untouched unless you opt in via config policy.")
	}
	flag.Parse()
	return runSettings{
		PolicyFlag: *policyFlag,
		ConfigPath: *configPath,
		Font:       *font,
		Width:      *width,
		Disable:    *disable,
		Demo:       *demo,
		JSON:       *jsonOut,
		ListFonts:  *listFonts,
		NoColor:    *noColor,
		Text:       *text,
		Version:    *showVersion,
	}
}

func buildBanner(ctx context.Context, renderer *ascii.Renderer, cfg config.Config, env terminal.Env, demo bool) (banner.Output, error) {
	if demo {
		hostBanner, err := banner.New(collectors.Providers{}, renderer, banner.BuildersForConfig(cfg))
		if err != nil {
			return banner.Output{}, err
		}
		return hostBanner.BuildWithSnapshot(collectors.DemoSnapshot(), cfg, env), nil
	}

	providers := collectors.Providers{
		System:    collectors.NewSystemCollector(),
		Network:   collectors.NewNetworkCollector(cfg.Network.MaxInterfaces),
		Resources: collectors.NewResourceCollector(),
		Session:   collectors.NewSessionCollector(),
		LastLogin: collectors.NewLastLoginCollector(),
	}
	hostBanner, err := banner.New(providers, renderer, banner.BuildersForConfig(cfg))
	if err != nil {
		return banner.Output{}, err
	}
	output, _, err := hostBanner.Build(ctx, cfg, env)
	return output, err
}

func runTextMode(renderer *ascii.Renderer, text string, cfg config.Config, env terminal.Env) error {
	art, err := renderer.Render(text, ascii.RenderOptions{
		Font:       cfg.ASCII.Font,
		Color:      cfg.ASCII.Color,
		Gradient:   cfg.ASCII.Gradient,
		Monochrome: cfg.ASCII.Monochrome,
		MaxWidth:   env.Width,
		Profile:    env.Profile,
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n%s\n\n", art.Text)
	return nil
}

func resolveInteractivity() bool {
	interactive := isInteractive()
	if os.Getenv("CI") != "" {
		interactive = false
	}
	if os.Getenv("SYSGREET_ASSUME_TTY") != "" {
		interactive = true
	}
	return interactive
}

func maybeBootstrap(ctx context.Context, settings runSettings) error {
	cfgPath := config.DefaultWritePath()
	if cfgPath == "" {
		return nil
	}
	policyEnv := os.Getenv("SYSGREET_CONFIG_POLICY")
	info, statErr := os.Stat(cfgPath)
	policyProvided := settings.PolicyFlag != "" || policyEnv != ""
	configMissing := errors.Is(statErr, os.ErrNotExist)
	configIsDir := statErr == nil && info.IsDir()
	if statErr != nil && !configMissing {
		return fmt.Errorf("stat config: %w", statErr)
	}
	if !policyProvided && !configMissing && !configIsDir {
		return nil
	}
	io := bootstrap.IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
	_, err := bootstrap.Bootstrap(ctx, cfgPath, io, bootstrap.Options{
		FlagPolicy:  settings.PolicyFlag,
		EnvPolicy:   policyEnv,
		Interactive: resolveInteractivity(),
	})
	return err
}

func isInteractive() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
