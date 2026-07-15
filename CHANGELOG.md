# Changelog

## Unreleased

### Fixed

- **`--version` tells the truth for `go install` builds** - Binaries built outside GoReleaser reported `dev (commit: none, built: unknown)`; they now resolve the module version and VCS metadata from Go's embedded build info. GoReleaser-injected values still take precedence.

## v1.1.0

### Added

- **Width-aware rendering** - Sysgreet now detects the terminal width and guarantees the banner never wraps mid-glyph. When the configured font is too wide it steps down gracefully: drop the domain from the hostname (the full name moves to an info line), try a narrower font, and finally render a clean single-line ruled header. Narrow tmux panes and phone SSH sessions stay readable.
- **`Small` font** - A narrower FIGlet font (from the standard figlet distribution) embedded as the last art step before the plain-header fallback.
- **Truecolor gradients** - On terminals advertising `COLORTERM=truecolor`, gradient stops are interpolated into a smooth 24-bit fade. 16-color terminals keep the existing per-line cycling.
- **`--json` flag** - Emit the banner as structured JSON for scripting (section data includes raw percentages for thresholding in pipelines). Never prompts for bootstrap.
- **New flags** - `--font` (per-run font override), `--width` (assume a terminal width), `--no-color`, `--config` (explicit config path), `--list-fonts`.
- **`layout.max_width`** - Config key (and `SYSGREET_LAYOUT_MAX_WIDTH`) to cap banner width below the detected terminal size.
- **Windows legacy console support** - Virtual terminal processing is enabled before writing escapes; when the console refuses, output falls back to plain text instead of printing raw escape codes.

### Fixed

- **`NO_COLOR` now covers the whole banner** - Previously the hostname art kept its gradient with `NO_COLOR` set; escape sequences are now fully suppressed, including when output is piped or `TERM=dumb`.
- **Piped output is plain** - `sysgreet > motd.txt` no longer embeds ANSI codes; color is keyed off whether stdout is a terminal.
- **`--text` and `--demo` honor your config** - Both modes previously ignored the config file, so custom fonts and gradients silently didn't apply.
- **Compact mode is actually compact** - `layout.compact` now emits a true single line using the plain hostname instead of embedding the multi-line art.
- **Unknown fonts fall back deterministically** - A typo in `ascii.font` now falls back to the default font (logged under `SYSGREET_DEBUG`) instead of picking a random font every login.
- **Long body lines clip cleanly** - Info lines that exceed the terminal width are truncated with an ellipsis instead of wrapping.

### Changed

- **Collectors run concurrently** - All five collectors gather in parallel under a shared 250 ms deadline, so a slow metric source can never hang a login. Windows logins no longer pay the 100 ms CPU sample serially.
- **One color palette** - The ANSI color tables previously duplicated (and drifted) across packages now live in a single `internal/terminal` package alongside width and capability detection.

## v0.9.1

### Added

- **Gradient color support** - Banner lines can cycle through color gradients (default: brightblue → blue → cyan → brightcyan → white)
- **6 new Unicode block fonts** - ANSI Regular, ANSI Shadow, Block, Blocks, DOS Rebel, Basic for compact, professional banners
- **`--demo` flag** - Display 'SYSGREET' banner with realistic fake data, perfect for screenshots and demos
- **`--text` flag** - Render custom text as ASCII art (e.g., `--text "Production DB"`)
- **Visual padding** - Added blank line before banner output for better aesthetics

### Changed

- **Default font** - Changed from `slant` to `ANSI Regular` (compact Unicode blocks)
- **Default colors** - Now uses gradient instead of single random color
- **Banner style** - Unicode block characters (█) for tighter, more readable output
- **Bootstrap message** - Updated to reflect new defaults (ANSI Regular font with gradient)

### Documentation

- Updated README.md with hero image and demo screenshot
- Updated all example configs to show gradient configuration
- Documented all 8 available fonts in docs/examples/fonts.md
- Added special modes section to README (demo, text, disable)
- Updated quickstart guide with gradient and new flag examples

## v0.1.0

- Initial cross-platform sysgreet banner implementation
- Embedded FIGlet fonts with ASCII-art hostname rendering
- System, network, and resource collectors with graceful degradation
- YAML/TOML configuration support with optional monochrome mode
- GoReleaser pipeline and GitHub Actions release workflow
