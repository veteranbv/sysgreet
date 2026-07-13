# Configuration Scenarios

## Default Configuration (Gradient Colors)

```yaml
ascii:
  font: "ANSI Regular"
  gradient: ["brightblue", "blue", "cyan", "brightcyan", "white"]
  monochrome: false

display:
  # All sections enabled by default
```

## Minimal Configuration (disable resources)

```yaml
display:
  memory: false
  disk: false
  load: false

ascii:
  font: "ANSI Regular"
  gradient: ["cyan", "blue"]  # Simple 2-color gradient
```

## Focus on Networking Only

```yaml
display:
  hostname: true
  os: true
  ip_addresses: true
  remote_ip: true
  uptime: false
  user: false
  memory: false
  disk: false
  load: false
  datetime: true
  last_login: false

network:
  show_interface_names: true
  max_interfaces: 5
```

## Narrow Terminals and tmux Panes

Sysgreet detects the terminal width at startup and degrades the banner
automatically, so nothing is required for split panes or phone SSH clients.
To cap the width below what the terminal reports (for example to keep logs
tidy), set:

```yaml
layout:
  max_width: 80
```

Useful companions:

- `sysgreet --width 60` previews what a 60-column session will see.
- `SYSGREET_LAYOUT_MAX_WIDTH=100` does the same per environment.
- On very tight widths the hostname renders as a single ruled line instead of
  multi-line art — it never wraps into garbage.

## Scripting with JSON

`sysgreet --json` emits the banner data as structured JSON (no art, no color,
no bootstrap prompts), with section `data` carrying raw values like
`memory_used_percent` for thresholding:

```bash
sysgreet --json | jq -r '.sections[] | select(.key=="resources").data.memory_used_percent'
```

## Troubleshooting Metric Discrepancies

If resource values appear inconsistent with native tools:

1. Run `go test -run TestResourceCollectorMatchesSystemStats ./test/integration/...` on the target OS.
2. Verify the user running `sysgreet` has permission to read filesystem metadata for the home directory.
3. Ensure no virtualization layers hide physical interfaces (VPN, container bridges). Adjust `network.max_interfaces` or disable specific sections if needed.
4. For Windows hosts, the CPU usage calculation relies on `cpu.PercentWithContext`. When the banner runs during login scripts, the first sampling window may be noisy; run the banner twice or lower the sampling interval via configuration if necessary.

## Non-interactive Bootstrap Policies

- `CI=1 bin/sysgreet --config-policy=keep` ensures automation never writes a config file (ideal for hosts that manage configs externally).
- `CI=1 SYSGREET_CONFIG_POLICY=overwrite bin/sysgreet` regenerates the default config on every run without prompting and keeps the latest backup beside the active file.
- If both a flag and environment variable are provided, the flag wins so one-off jobs can override fleet defaults.
