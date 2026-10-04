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

- **Single static binary** - Go 1.26+ to build, no CGO, no daemons, no service
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
# Via Go (requires Go 1.26+)
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

Run sysgreet only from interactive shells. Shell startup files also run for
`scp`, `rsync`, `sftp` and `ssh host cmd`, and anything printed there corrupts
those transfers. The snippets below guard against that; sysgreet also stays
silent on its own in a non-interactive SSH session, as a second line of
defense.

```bash
# Bash (~/.bashrc) or Zsh (~/.zshrc)
[[ $- == *i* ]] && command -v sysgreet >/dev/null && sysgreet
```

```fish
# Fish (~/.config/fish/config.fish)
status is-interactive; and type -q sysgreet; and sysgreet
```

```powershell
# PowerShell ($PROFILE); profiles only load for interactive sessions
if (Get-Command sysgreet -ErrorAction SilentlyContinue) { sysgreet }
```

To check a remote host without logging in, give the session a terminal
(`ssh -t pve1 sysgreet`) or pass `--force` (`ssh pve1 sysgreet --force`).

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
  sections: ["header", "system", "network", "resources"]

network:
  show_interface_names: true
  max_interfaces: 3
```

Environment variables override everything (e.g.
`SYSGREET_DISPLAY_MEMORY=false`, `SYSGREET_ASCII_FONT=standard`). See
[`configs/example.yaml`](configs/example.yaml) and
[`configs/example.toml`](configs/example.toml) for full references.

### Starter config

A normal run never writes files or prompts: with no config file, sysgreet
uses its built-in defaults. A broken config costs a one-line warning on
stderr and the banner renders with defaults; it never fails a login.

To get an editable starter config, run it once explicitly:

```bash
sysgreet --init-config        # writes ~/.config/sysgreet/config.yaml
```

- If the file already exists, an interactive run asks whether to
  `[K]eep`, `[O]verwrite` or `[C]ancel`. Without a terminal it keeps the
  existing file.
- `--config-policy keep|overwrite|prompt` (or `SYSGREET_CONFIG_POLICY`)
  decides ahead of time. The flag wins over the environment variable.
- Overwriting renames the old file to a timestamped `.bak` first. Backups are
  never deleted.

---

## What the banner shows

![Demo output](media/demo.jpg)

The hostname art comes first, then the OS line, then three sections laid out
side by side when the terminal is wide enough (see
[`docs/examples/default-output.md`](docs/examples/default-output.md) for real
output at 140, 80, 50 and 30 columns):

```text
System                                     Network                       Resources
  Uptime      4d 12h                         eth0        192.168.1.42      Mem   ██░░░░░░░░  23%  3.7/16.0 GiB
  User        demo                           tailscale0  100.101.42.7      Disk  █████████░  87%  412.0/476.0 GiB
  Time        Sun 04 Oct 01:32 UTC           From        192.168.1.20      Load  █░░░░░░░░░ 0.45  8 cores
  Last login  26h ago from 192.168.1.20
```

- **System** - Uptime, current user (bold red when you are root), local time,
  and your previous login from the system's login history (Linux).
- **Network** - The address carrying the default route first, then other
  physical interfaces, each labeled by interface name. `From` is the SSH
  client. Loopback, link-local, down interfaces, and container/VM bridges
  (Docker, libvirt, CNI, LXD, Incus, Podman) stay out of view.
- **Resources** - Usage meters for memory, the root filesystem (measured like
  `df`), and the 1-minute load against the core count. Meters turn yellow at
  75% and red at 90%. Windows shows realtime CPU usage instead of load.

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
go mod download
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

- CI (`.github/workflows/ci.yml`) runs `golangci-lint`, the test suite on
  Linux, macOS and Windows (plus the oldest supported Go), the race detector,
  `govulncheck`, and a startup check that fails if a full banner takes more
  than 250ms. Actions are pinned to commit SHAs and Dependabot keeps them and
  the Go modules current.
- To cut a release, run the **Tag Release** workflow from the Actions tab
  with a `vX.Y.Z` version (or push a `v*` tag manually). It tags `main` and
  hands off to the Release workflow.
- Releases use GoReleaser (`.goreleaser.yml`) with the latest Go release to
  build reproducible binaries for Linux, macOS and Windows (amd64 and arm64),
  plus checksums. Every archive gets a signed build-provenance attestation;
  verify a download with
  `gh attestation verify sysgreet_*.tar.gz --repo veteranbv/sysgreet`.
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
