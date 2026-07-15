# Sysgreet

[![Release](https://img.shields.io/github/v/release/veteranbv/sysgreet)](https://github.com/veteranbv/sysgreet/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/veteranbv/sysgreet)](https://golang.org/dl/)
[![License](https://img.shields.io/github/license/veteranbv/sysgreet)](LICENSE)

> Beautiful, low-latency system context for every terminal login.

![Sysgreet](media/sysgreet-padded.png)

Sysgreet keeps you oriented the moment a shell prompt appears. It prints the
hostname in ASCII art alongside a curated snapshot of operating system,
network, and resource telemetry, so you always know **which** machine you are on
and **whether** it is healthy. Built for managing home labs and fleets alike, it
remains lightweight, offline-friendly, and cross-platform across Linux, macOS,
and Windows.

---

## Why Sysgreet exists

I created Sysgreet while operating a growing home lab and juggling
multiple SSH sessions. I wanted a professional banner (_not_ a novelty) that
instantly answered three questions:

1. **Where am I logged in?** (Hostname, OS, architecture, remote source)
2. **Is this host behaving?** (Uptime, memory, disk, CPU trends)
3. **What network path am I on?** (Primary route, relevant secondary interfaces)

Sysgreet delivers those answers in under 50 ms without calling out to the
network or depending on external runtimes.

![Home Lab Example](media/homelab_server_example.jpg)

---

## Highlights

- **Single static binary** - Go 1.22+, no CGO, no daemons, no service
  dependencies.
- **Fits any terminal** - Sysgreet measures the terminal before printing and
  steps the banner down gracefully (shorter hostname, then a narrower font,
  then a clean one-line header) instead of wrapping ASCII art into garbage.
  Split tmux panes and phone SSH sessions stay readable.
- **Cross-platform parity** - Linux/macOS show load averages; Windows surfaces
  CPU usage (and legacy consoles get plain text instead of raw escape codes).
  Interface filtering avoids noisy virtual adapters everywhere.
- **Configurable yet optional** - YAML or TOML profiles toggle sections, pick
  fonts/colors, set layout order, and cap the interface list. Defaults "just
  work" with zero files.
- **Graceful degradation** - Missing metrics or SSH metadata simply fall back;
  the banner keeps rendering.
- **Performance-guarded** - Collectors run in parallel under a hard 250 ms
  deadline; the startup benchmark (<50 ms median, <80 ms p95) runs in CI and
  process RSS stays <15 MB.
- **Professional aesthetics** - Unicode block fonts with gradient colors —
  smooth 24-bit fades on truecolor terminals — plus automatic monochrome
  fallback and full `NO_COLOR` support.
- **Script-friendly** - `--json` emits the same data as structured JSON;
  piped output is always plain text.

---

## Quick start

### Install the binary

```bash
# Via Go (requires Go 1.22+)
go install github.com/veteranbv/sysgreet/cmd/sysgreet@latest

# Ensure Go's bin directory is in your PATH
# Add this to ~/.bashrc, ~/.zshrc, or equivalent if not already present:
export PATH="$HOME/go/bin:$PATH"

# Or download a release artifact (Linux/macOS/Windows, amd64 & arm64)
# https://github.com/veteranbv/sysgreet/releases
```

> _Tip:_ The binary runs entirely offline. Copy it between hosts without
> worrying about external assets.

### Update to latest version

```bash
# Via Go (silent on success)
go install github.com/veteranbv/sysgreet/cmd/sysgreet@latest

# Verify the update
sysgreet --version

# Or download the latest release
# https://github.com/veteranbv/sysgreet/releases
```

### Wire into your shell

| Shell            | Snippet                                                                                       |
|------------------|------------------------------------------------------------------------------------------------|
| Bash / Zsh       | `echo 'sysgreet' >> ~/.bashrc` (or `~/.zshrc`)                                                 |
| Fish             | `echo 'sysgreet' >> ~/.config/fish/config.fish`                                               |
| PowerShell       | `Add-Content $PROFILE 'sysgreet'`                                                             |
| Windows Terminal | Add `sysgreet` to your profile script so it runs after each session attaches                  |
| SSH `ForceCommand` | `ForceCommand /usr/local/bin/sysgreet && /bin/bash` (keeps banner even when no profile runs) |

**Special modes:**

```bash
# Demo mode - show 'SYSGREET' with fake data (perfect for screenshots)
sysgreet --demo

# Disable output (useful in CI/scripts)
sysgreet --disable

# Text mode - render any custom ASCII art on the fly
sysgreet --text "Production DB"
sysgreet --text "Coffee Break"
sysgreet --text "Deploy Day"

# JSON mode - structured output for scripts (no art, no prompts)
sysgreet --json | jq -r '.hostname'
```

**Useful flags:**

```bash
sysgreet --list-fonts          # Print the embedded fonts
sysgreet --font "ANSI Shadow"  # One-off font override
sysgreet --width 60            # Preview how a 60-column session renders
sysgreet --no-color            # Plain output (NO_COLOR works too)
sysgreet --config ~/alt.yaml   # Point at a specific config file
```

> **Tip:** Use `--text` to create custom banners for different environments, reminders, or just for fun. Great for distinguishing production boxes, marking maintenance windows, or adding personality to your terminals.

![Text mode example](media/text.png)
![Winning](media/test-winning.png)
---

## Configuration (optional)

Sysgreet looks for configuration in this order:

1. `--config` flag or `SYSGREET_CONFIG` environment variable (absolute or
   `~/` paths). An explicit path is exclusive — if the file is missing,
   sysgreet uses built-in defaults rather than silently reading another
   config.
2. `~/.config/sysgreet/config.yaml` (or `.yml`, `.toml`)
3. `~/.sysgreet.yaml` / `.toml`

Example YAML:

```yaml
# ~/.config/sysgreet/config.yaml
ascii:
  font: "ANSI Regular"
  gradient: ["brightblue", "blue", "cyan", "brightcyan", "white"]
  monochrome: false

display:
  hostname: true
  os: true
  ip_addresses: true
  remote_ip: true
  uptime: true
  user: true
  memory: true
  disk: true
  load: true
  datetime: true
  last_login: true

layout:
  compact: false
  max_width: 0 # cap banner width in columns; 0 = detected terminal width
  sections: ["header", "network", "system", "resources"]

network:
  show_interface_names: true
  max_interfaces: 4
```

Environment variables override everything (e.g.
`SYSGREET_DISPLAY_MEMORY=false`, `SYSGREET_ASCII_FONT=standard`). See
[`configs/example.yaml`](configs/example.yaml) and
[`configs/example.toml`](configs/example.toml) for full references.

### Bootstrap behaviour

- First run: sysgreet writes `~/.config/sysgreet/config.yaml` with curated defaults (all sections enabled, `ANSI Regular` font with blue-to-white gradient, metadata fields `created_at` and `version`).
- Existing config: sysgreet leaves the file untouched by default. Provide `--config-policy prompt` (or `SYSGREET_CONFIG_POLICY=prompt`) to surface the `[K]eep/[O]verwrite/[C]ancel` flow, or `overwrite` to regenerate the defaults (a timestamped `.bak` is created first).
- Non-interactive automation: use `--config-policy` or `SYSGREET_CONFIG_POLICY` to choose `prompt`, `keep`, or `overwrite`. When stdin is not a TTY (e.g. CI jobs), an explicit policy is required.
- Flags beat environment variables so scripts can override fleet defaults (`SYSGREET_CONFIG_POLICY=overwrite bin/sysgreet --config-policy=keep`).

---

## What the banner shows

![Demo output](media/demo.jpg)

- **System** - Hostname (ASCII art), OS name/version, architecture, uptime,
  active user + home, current time, last login when available.
- **Network** - Primary outbound interface based on routing table, filtered list
  of secondary physical interfaces, SSH remote IP (from `SSH_CONNECTION` or
  `SSH_CLIENT`). Loopback, link-local, Docker/VM, and down interfaces stay out of
  view by default.
- **Resources** - Memory, disk, and CPU metrics with highlight thresholds (≥75% in
  yellow, ≥90% in red). Windows surfaces realtime CPU usage; Unix hosts show load
  averages.

### Terminal width handling

The hostname art is only printed when it actually fits. On narrow terminals
sysgreet steps down automatically:

1. Full hostname in your configured font
2. Hostname without the domain (`pve1.home.lan` → `PVE1`, with the full name
   kept on an info line)
3. Progressively narrower fonts (`standard`, then `Small`)
4. A single ruled line — `═════ PVE1 ═════` — that fits any width

Set `layout.max_width` to cap the width below what the terminal reports, or
pass `--width` to preview a specific size.

---

## Performance guarantees

- **Startup** - `< 50 ms` median, `< 80 ms` p95 (validated by
  `go test -bench Startup ./test/benchmarks`)
- **Never hangs a login** - Collectors run concurrently under a shared 250 ms
  deadline; a stuck metric source just drops its section
- **Binary footprint** - `< 10 MB` for all release targets (GoReleaser checks)
- **Runtime memory** - `< 15 MB` RSS for default banner
- **No network activity** - All data collected locally, offline-safe

Enable `SYSGREET_DEBUG=1` to log collector errors without interrupting output.

---

## Development & contribution

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development guidelines, code standards, and workflow.

**Quick start:**

```bash
git clone https://github.com/veteranbv/sysgreet.git
cd sysgreet
go mod tidy
make test
make bench
```

**Common tasks:**

```bash
make fmt            # Format code
make lint           # Run linters (requires golangci-lint)
make test-coverage  # Run tests with coverage report
make build          # Build the binary
```

PRs are welcome. Please open an issue describing new collectors, layout ideas, or
platform-specific improvements before diving in.

---

## Release process

- CI (`.github/workflows/ci.yml`) runs `golangci-lint`, unit tests with race
  detection, integration tests, and validates startup performance (<80ms p95).
- To cut a release, run the **Tag Release** workflow from the Actions tab
  with a `vX.Y.Z` version (or push a `v*` tag manually). It tags `main` and
  hands off to the Release workflow.
- Releases use GoReleaser (`.goreleaser.yml`) to ship signed binaries for
  Linux/macOS (amd64/arm64) and Windows (amd64), plus checksums.
- `go install github.com/veteranbv/sysgreet@VERSION` is validated during the
  release workflow.

---

## Roadmap

- Extended GPU/storage telemetry for workstation profiles
- Pluggable section framework (e.g., Kubernetes context, vault status)
- Prebuilt Windows installer for enterprise onboarding

Ideas welcome. Open a discussion if a feature would make Sysgreet more useful for
your fleet.

---

## License

Sysgreet is licensed under the [Apache License 2.0](LICENSE).

Copyright © 2025 Henry Sowell
