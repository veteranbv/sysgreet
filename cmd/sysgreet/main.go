package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/veteranbv/sysgreet/internal/ascii"
	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/bootstrap"
	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/render"
	"github.com/veteranbv/sysgreet/internal/terminal"
	"golang.org/x/term"
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

	if done, err := runUtilityMode(ctx, settings); done {
		return err
	}

	renderer, err := ascii.NewRenderer()
	if err != nil {
		return err
	}

	// Legacy Windows consoles need virtual terminal processing switched on
	// before any escape sequences are written; when that fails, fall back
	// to plain output rather than printing raw escapes.
	ansiOK := enableVirtualTerminal(os.Stdout)
	env := terminal.DetectEnv(os.Stdout, settings.NoColor || !ansiOK)

	cfg := loadConfig(settings)
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

// runUtilityMode handles every invocation that finishes without rendering a
// banner. It reports whether the run is done.
func runUtilityMode(ctx context.Context, settings runSettings) (bool, error) {
	if settings.Version {
		v, c, d := buildInfo()
		fmt.Printf("sysgreet %s (commit: %s, built: %s)\n", v, c, d)
		return true, nil
	}

	if settings.ConfigPath != "" {
		// Reuse the SYSGREET_CONFIG plumbing so load and bootstrap agree
		// on the path.
		if err := os.Setenv("SYSGREET_CONFIG", settings.ConfigPath); err != nil {
			return true, err
		}
	}

	switch {
	case settings.ListFonts:
		renderer, err := ascii.NewRenderer()
		if err != nil {
			return true, err
		}
		for _, font := range renderer.Fonts() {
			fmt.Println(font)
		}
		return true, nil
	case settings.InitConfig:
		return true, initConfig(ctx, settings)
	case settings.Disable || envTrue("SYSGREET_DISABLE"):
		// Silences output, never an explicit command like --init-config.
		return true, nil
	}

	bannerMode := settings.Text == "" && !settings.Demo && !settings.JSON
	if bannerMode && !settings.Force && nonInteractiveSSH() {
		// Shell rc files run for scp, rsync, sftp and `ssh host cmd` too;
		// banner bytes on stdout would corrupt those protocols.
		return true, nil
	}
	return false, nil
}

// loadConfig never fails: a broken config file costs a one-line warning,
// not the banner. A login banner that errors on every login is worse than
// one rendered with defaults.
func loadConfig(settings runSettings) config.Config {
	cfg, _, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sysgreet: ignoring config %v; using defaults\n", err)
	}
	if settings.Font != "" {
		cfg.ASCII.Font = settings.Font
	}
	return cfg
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
	InitConfig bool
	Force      bool
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
	initConfig := flag.Bool("init-config", false, "Write a starter config file and exit")
	policyFlag := flag.String("config-policy", "", "With --init-config, what to do with an existing config: prompt, keep, or overwrite")
	configPath := flag.String("config", "", "Path to a config file (overrides default lookup)")
	font := flag.String("font", "", "Font override for this run (see --list-fonts)")
	width := flag.Int("width", 0, "Assume this terminal width instead of detecting it")
	force := flag.Bool("force", false, "Print the banner even in a non-interactive SSH session")
	disable := flag.Bool("disable", false, "Print nothing and exit")
	demo := flag.Bool("demo", false, "Demo mode with 'SYSGREET' banner and fake data")
	jsonOut := flag.Bool("json", false, "Emit the banner as JSON for scripting")
	listFonts := flag.Bool("list-fonts", false, "List embedded fonts and exit")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	text := flag.String("text", "", "Render custom text as ASCII art (e.g., --text \"Tea Pot\")")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "sysgreet prints a login banner: the hostname in ASCII art plus system,")
		fmt.Fprintln(w, "network, and resource details.")
		fmt.Fprintln(w, "\nUsage: sysgreet [flags]\n\nFlags:")
		flag.PrintDefaults()
		fmt.Fprintln(w, "\nEnvironment variables:")
		fmt.Fprintln(w, "  SYSGREET_CONFIG          Config file path (same as --config)")
		fmt.Fprintln(w, "  SYSGREET_CONFIG_POLICY   Policy for --init-config (prompt|keep|overwrite)")
		fmt.Fprintln(w, "  SYSGREET_DISABLE         Print nothing when set to 1/true (fleet opt-out)")
		fmt.Fprintln(w, "  SYSGREET_DEBUG           Log collector errors to stderr")
		fmt.Fprintln(w, "  NO_COLOR                 Disable colored output (same as --no-color)")
		fmt.Fprintln(w, "  SYSGREET_DISPLAY_*, SYSGREET_ASCII_*, SYSGREET_LAYOUT_*, SYSGREET_NETWORK_*")
		fmt.Fprintln(w, "                           Override individual config keys (see README)")
		fmt.Fprintln(w, "\nIn a non-interactive SSH session (scp, rsync, `ssh host cmd`) sysgreet")
		fmt.Fprintln(w, "prints nothing so it cannot corrupt the transfer; use --force or `ssh -t`.")
	}
	flag.Parse()
	return runSettings{
		InitConfig: *initConfig,
		Force:      *force,
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
	interactive := isTerminal(os.Stdin)
	if os.Getenv("CI") != "" {
		interactive = false
	}
	if os.Getenv("SYSGREET_ASSUME_TTY") != "" {
		interactive = true
	}
	return interactive
}

// initConfig writes the starter config. It is the only path that writes to
// disk; a normal banner run has no side effects.
func initConfig(ctx context.Context, settings runSettings) error {
	cfgPath := config.DefaultWritePath()
	if cfgPath == "" {
		return errors.New("cannot determine config path: no home directory; pass --config")
	}
	if !config.SupportedPath(cfgPath) {
		return fmt.Errorf("cannot write %s: config files must end in .yaml, .yml, or .toml", cfgPath)
	}
	io := bootstrap.IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
	_, err := bootstrap.Bootstrap(ctx, cfgPath, io, bootstrap.Options{
		FlagPolicy:  settings.PolicyFlag,
		EnvPolicy:   os.Getenv("SYSGREET_CONFIG_POLICY"),
		Interactive: resolveInteractivity(),
	})
	return err
}

// nonInteractiveSSH reports a session sshd started for a command rather
// than a login shell: SSH variables are set but neither stdin nor stdout
// is a terminal.
func nonInteractiveSSH() bool {
	if os.Getenv("SSH_CONNECTION") == "" && os.Getenv("SSH_CLIENT") == "" {
		return false
	}
	if isTerminal(os.Stdin) || isTerminal(os.Stdout) {
		return false
	}
	// Only a pipe or socket can carry an scp/rsync/sftp stream. Output
	// redirected to a file (`ssh host 'sysgreet > /etc/motd'`) is wanted:
	// suppressing it would leave the shell's truncated, empty file.
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&(os.ModeNamedPipe|os.ModeSocket) != 0
}

func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

func envTrue(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
