# Embedded FIGlet Fonts

The Sysgreet banner embeds multiple FIGlet fonts to ensure the CLI operates offline:

## Available Fonts

- **`ANSI Regular.flf`** (default) - Compact Unicode block characters with solid fills
- **`ANSI Shadow.flf`** - Unicode blocks with shadow effects
- **`Block.flf`** - Classic blocky style
- **`Blocks.flf`** - Variant of block style
- **`DOS Rebel.flf`** - Modern, clean DOS-style font
- **`Banner.flf`** - Classic wide banner style
- **`Basic.flf`** - Simple hash-mark style
- **`Small.flf`** - Narrow variant of standard, used automatically on tight terminals
- **`standard.flf`** - Classic FIGlet standard font
- **`slant.flf`** - Tall, angular slanted font

Run `sysgreet --list-fonts` to print this list from the binary itself, and
`sysgreet --text "Test" --font "ANSI Shadow"` to preview one without touching
your config.

All fonts originate from the FIGlet project (<http://www.figlet.org/>) and community repositories. They are licensed under the FIGlet Font License, which permits redistribution and embedding in binaries. License text is included at the top of each font file.

Fonts are stored in `assets/fonts/` and shipped with the binary via Go's `//go:embed` directive.

## Usage

Fonts can be configured in your config file:

```yaml
ascii:
  font: "ANSI Regular"  # Default - compact Unicode blocks
  # or "ANSI Shadow", "Banner", "Basic", "Block", "Blocks", "DOS Rebel",
  # "Small", "standard", "slant"
```

Or test fonts quickly:

```bash
sysgreet --text "Test" --font "Small"
```

## Width behavior

The configured font is a preference, not a promise. Sysgreet measures the art
against the terminal before printing; when it would overflow, it drops the
domain from the hostname, then tries `standard` and `Small`, and finally
renders a one-line ruled header. A misspelled font name falls back to the
default (`ANSI Regular`) — set `SYSGREET_DEBUG=1` to see a note when that
happens.
