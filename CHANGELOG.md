# Changelog

## Unreleased

### Changed

- **Dashboard layout** - Sections render as aligned label/value columns and sit side by side when the terminal is wide enough: at 140 columns the info panel is 5 lines instead of 17, and at 80 columns System and Network share a row. Piped output still stacks. Memory, disk, and load render as usage meters colored by threshold; titles take the art's first gradient color and labels are dimmed.
- **Narrow terminals degrade gracefully** - Details drop and meters shrink before any value is clipped, and no line exceeds the width at any size.
- **`--json` items** - Each section now also carries structured `items` (label, value, detail, meter, level). `lines` and `data` are unchanged in shape, and the System section's `data` now has exact values (`uptime_seconds`, RFC 3339 `time` and `last_login`, `last_login_from`); the `lines` text follows the new format (`Mem: 23% 3.7/16.0 GiB`).
- **A login banner never breaks a login** - Normal runs no longer write a config file or prompt. On a fresh host, `ssh host cmd`, cron, and Ansible runs previously exited with `config policy required` on every invocation until someone logged in with a terminal; they now render from built-in defaults. Writing the starter config is an explicit `sysgreet --init-config` (it refuses extensions the loader cannot read, and refuses to guess a location when there is no home directory), and `--config-policy`/`SYSGREET_CONFIG_POLICY` now only apply to it.
- **Broken config degrades instead of failing** - A config that cannot be read or parsed (including an explicit `--config` path that is missing or a directory) prints one warning line and the banner renders with defaults (environment overrides still apply), exiting 0.

### Fixed

- **Silent in non-interactive SSH sessions** - Shell startup files also run for scp, rsync, sftp, and `ssh host cmd`; a banner on stdout corrupts those transfers. When SSH variables are set, neither stdin nor stdout is a terminal, and stdout is a pipe or socket, sysgreet now prints nothing. Redirecting to a file (`ssh host 'sysgreet > /etc/motd'`) still works. `--force` or `ssh -t` overrides it. The README's shell snippets now include the standard interactive-shell guard, and the `ForceCommand` example, which dropped the client's command and broke scp, sftp, and rsync, is gone.
- **Config backups are never deleted** - Overwriting the config pruned every older backup, so a second overwrite lost the user's original file. All backups are now kept, and two overwrites in the same second no longer share a name.
- **Invalid boolean environment values are ignored** - `SYSGREET_DISPLAY_MEMORY=ture` used to turn the setting off; it now leaves it unchanged.
- **Policy errors name the bad value** - `invalid config policy value "replace" (want prompt, keep, or overwrite)`.
- **Last login was the current session** - It read the list of currently logged-in sessions and picked the newest, which is the shell that just started, so it always said "just now". It now reads the login history (`/var/log/wtmp`) and reports the login before the current session, including in tmux panes and new terminal tabs, which have no history record of their own. It is omitted where no history exists (macOS, Windows, wtmpdb-only systems) instead of being wrong.
- **Disk usage was overstated** - Blocks reserved for root were counted as used, overstating usage by the reserve (5 points on a default ext4 disk, 66 points in one container: 88% shown against 22% in `df`). It now matches `df`, and measures the root filesystem (the writable data volume on macOS, the system drive on Windows) rather than the home directory's.
- **OS name joined the distribution and its family** - "Ubuntu Debian 24.04", "Rocky Rhel", "Darwin Standalone workstation". Linux now uses the distribution's own name from `/etc/os-release` ("Ubuntu 24.04.4 LTS") without repeating the version; macOS shows "macOS 15.0".
- **Primary address was a name-based guess** - A libvirt or CNI bridge could outrank the real Wi-Fi address. The primary is now the address carrying the default route (a route lookup, no packets sent), container and VM bridges are filtered, IPv6-only hosts show their global address, and an interface's aliases no longer fill the other slots.
- **Load had no scale** - Load is now shown against the core count, so 4.03 on 4 cores reads as saturated.
- **`display.hostname: false` did nothing** - It now omits the art.
- **`layout.sections: [header]` sorted sections alphabetically** - It now keeps the default order.
- **Hostname survives a partial system probe failure** - It falls back to `os.Hostname()` instead of rendering "UNKNOWN".
- **SSH clients over dual-stack sshd** - `::ffff:203.0.113.9` now shows as `203.0.113.9`.
- **Art padding** - Fonts that pad with U+2001 left a ragged edge and a blank row under the art; both are trimmed.

### Added

- `SYSGREET_DISABLE=1` silences the banner for a user or host without editing shell startup files. Explicit commands such as `--init-config` and `--list-fonts` still run.
- `--force` prints the banner even in a non-interactive SSH session.
- `--help` now describes the tool and lists the environment overrides.

### Build and release

- **Go 1.26 or later is required to build from source.** Release binaries are built with the latest Go release, so they ship with current standard-library security fixes; v1.2.1 was built with Go 1.22, which has known vulnerabilities reachable from sysgreet.
- **No more cgo warning on `go install` for macOS** - gopsutil v4 dropped the `go-m1cpu` dependency that printed a C compiler warning during install.
- **Reproducible release builds** - The same tag now produces byte-identical binaries (`-trimpath`, commit-based timestamps).
- **Build provenance** - Each release is attested with GitHub's build provenance; verify a download with `gh attestation verify <file> --repo veteranbv/sysgreet`.
- Windows release archives are `.zip`.
- CI runs on Linux, macOS and Windows, adds the race detector and `govulncheck`, pins actions to commit SHAs, and fails if a full banner takes longer than 250ms. Dependabot keeps actions and modules current.

## v1.2.1

### Fixed

- **`--version` tells the truth for `go install` builds** - Binaries built outside GoReleaser reported `dev (commit: none, built: unknown)`; they now resolve the module version and VCS metadata from Go's embedded build info. GoReleaser-injected values still take precedence.

## v1.2.0

_Also tagged as v1.1.0. Both tags point at the same commit and ship identical builds._

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
