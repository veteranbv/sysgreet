package main

import "runtime/debug"

// buildInfo resolves the version metadata shown by --version. GoReleaser
// injects real values via ldflags for release binaries; go-install and
// plain go-build binaries keep the defaults, so fall back to the module
// build info Go embeds in every binary.
func buildInfo() (string, string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, date
	}
	return resolveBuildInfo(info, version, commit, date)
}

func resolveBuildInfo(info *debug.BuildInfo, v, c, d string) (string, string, string) {
	if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if c == "none" {
				c = s.Value
			}
		case "vcs.time":
			if d == "unknown" {
				d = s.Value
			}
		}
	}
	return v, c, d
}
