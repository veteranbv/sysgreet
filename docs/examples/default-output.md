# Banner Layout Reference

Every block below is real output from `sysgreet --demo --no-color --width N`,
regenerated from the binary. The demo snapshot is fixed fake data, so the
layout is the only thing that changes between widths. The exact body layout
at the piped, 80-column and 140-column widths is also pinned by
`internal/render/testdata/*.golden`.

## Wide terminal (140 columns)

Sections sit side by side, so the whole banner fits in a handful of lines.

```text
███████ ██    ██ ███████  ██████  ██████  ███████ ███████ ████████
██       ██  ██  ██      ██       ██   ██ ██      ██         ██
███████   ████   ███████ ██   ███ ██████  █████   █████      ██
     ██    ██         ██ ██    ██ ██   ██ ██      ██         ██
███████    ██    ███████  ██████  ██   ██ ███████ ███████    ██

Ubuntu 24.04.4 LTS (amd64)

System                                     Network                       Resources
  Uptime      4d 12h                         eth0        192.168.1.42      Mem   ██░░░░░░░░  23%  3.7/16.0 GiB
  User        demo                           tailscale0  100.101.42.7      Disk  █████████░  87%  412.0/476.0 GiB
  Time        Sun 04 Oct 01:32 UTC           From        192.168.1.20      Load  █░░░░░░░░░ 0.45  8 cores
  Last login  26h ago from 192.168.1.20
```

## Standard terminal (80 columns)

Sections wrap to a new row when the next one would not fit.

```text
███████ ██    ██ ███████  ██████  ██████  ███████ ███████ ████████
██       ██  ██  ██      ██       ██   ██ ██      ██         ██
███████   ████   ███████ ██   ███ ██████  █████   █████      ██
     ██    ██         ██ ██    ██ ██   ██ ██      ██         ██
███████    ██    ███████  ██████  ██   ██ ███████ ███████    ██

Ubuntu 24.04.4 LTS (amd64)

System                                     Network
  Uptime      4d 12h                         eth0        192.168.1.42
  User        demo                           tailscale0  100.101.42.7
  Time        Sun 04 Oct 01:32 UTC           From        192.168.1.20
  Last login  26h ago from 192.168.1.20

Resources
  Mem   ██░░░░░░░░  23%  3.7/16.0 GiB
  Disk  █████████░  87%  412.0/476.0 GiB
  Load  █░░░░░░░░░ 0.45  8 cores
```

## Narrow pane (50 columns)

The art steps down to a narrower font and sections stack.

```text
════════════════════ SYSGREET ════════════════════

Ubuntu 24.04.4 LTS (amd64)

System
  Uptime      4d 12h
  User        demo
  Time        Sun 04 Oct 01:32 UTC
  Last login  26h ago from 192.168.1.20

Network
  eth0        192.168.1.42
  tailscale0  100.101.42.7
  From        192.168.1.20

Resources
  Mem   ██░░░░░░░░  23%  3.7/16.0 GiB
  Disk  █████████░  87%  412.0/476.0 GiB
  Load  █░░░░░░░░░ 0.45  8 cores
```

## Very narrow (30 columns)

The art becomes a single ruled line; meters shrink and details drop before
any value is clipped. No line ever exceeds the width.

```text
══════════ SYSGREET ══════════

Ubuntu 24.04.4 LTS (amd64)

System
  Uptime      4d 12h
  User        demo
  Time        Sun 04 Oct 01:3…
  Last login  26h ago from 19…

Network
  eth0        192.168.1.42
  tailscale0  100.101.42.7
  From        192.168.1.20

Resources
  Mem   ██░░░░░░░░  23%
  Disk  █████████░  87%
  Load  █░░░░░░░░░ 0.45
```

## Reading the banner

- **Meters** show used/total for memory and disk, and the 1-minute load
  against the core count. They turn yellow at 75% and red at 90%; the
  numbers after the bar are the same data in GiB.
- **Disk** is the root filesystem (the system drive on Windows), measured the
  way `df` does: blocks reserved for root are excluded.
- **Network** lists the address carrying the default route first, then other
  physical interfaces, each labeled with its interface name. `From` is the
  SSH client.
- **Last login** is your previous login from the system's login history, not
  the session you are in.
- **User** turns bold red when you are root.

Without a terminal (piped output, `> motd.txt`) sections always stack, so the
output is stable for scripts. `--json` gives the same data structured.
