# Default Sysgreet Banner Snapshot

This reference output documents the constitutional gates enforced by the Sysgreet CLI.

## Constitution Gates Checklist

- ✅ Single-binary Go CLI (no external runtime dependencies)
- ✅ Cross-platform parity (Linux, macOS, Windows collectors tested)
- ✅ Startup < 50ms measured via `go test -bench Startup ./test/benchmarks`
- ✅ Unicode block ASCII art with gradient colors and monochrome fallback
- ✅ Offline operation with embedded assets and configuration defaults

## Sample Output (demo mode)

Run `sysgreet --demo` to see the default output with fake data:

```text
███████ ██    ██ ███████  ██████  ██████  ███████ ███████ ████████
██       ██  ██  ██      ██       ██   ██ ██      ██         ██
███████   ████   ███████ ██   ███ ██████  █████   █████      ██
     ██    ██         ██ ██    ██ ██   ██ ██      ██         ██
███████    ██    ███████  ██████  ██   ██ ███████ ███████    ██

Linux Server 6.8.0 (x86_64)

System
  Uptime: 4d 12h 33m
  User: demo /home/demo
  Time: Sat, 11 Oct 2025 14:32:13 EDT
  Last login: Sat, 11 Oct 2025 12:32:13 EDT (203.0.113.10)

Network
  Primary: 192.168.1.42 (eth0)
  Secondary: 10.8.0.2 (tun0)
  Remote: 203.0.113.5

Resources
  Mem: 12.3GB free / 16.0GB (23% used)
  Disk: 210.0GB used / 512.0GB (41% used)
  CPU Load: 0.45 0.52 0.60
```

Note: The banner uses ANSI Regular font with a blue-to-white gradient (brightblue → blue → cyan → brightcyan → white).

## Narrow Terminal Degradation

The same banner adapts to the available width instead of wrapping. At 50
columns the default font no longer fits, so sysgreet steps down to the
`Small` font:

```text
                                            _
  ___  _  _   ___  __ _   _ _   ___   ___  | |_
 (_-< | || | (_-< / _` | | '_| / -_) / -_) |  _|
 /__/  \_, | /__/ \__, | |_|   \___| \___|  \__|
       |__/       |___/
```

And when even that is too wide (long hostnames, very tight panes), the header
collapses to a single ruled line:

```text
═════════════ SYSGREET ═════════════
```

Preview any width with `sysgreet --width <n>`.

Use this snapshot for QA validation and regression testing until golden files are finalized.
