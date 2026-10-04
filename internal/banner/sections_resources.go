package banner

import (
	"fmt"
	"math"

	"github.com/veteranbv/sysgreet/internal/collectors"
	"github.com/veteranbv/sysgreet/internal/config"
)

// ResourceSectionBuilder renders memory/disk/CPU information.
type ResourceSectionBuilder struct{}

// Key returns resource identifier.
func (ResourceSectionBuilder) Key() string { return "resources" }

// Enabled reports whether any resource displays are enabled.
func (ResourceSectionBuilder) Enabled(cfg config.Config) bool {
	return cfg.Display.Memory || cfg.Display.Disk || cfg.Display.Load
}

// Build renders each resource as a usage meter. Data keeps the raw
// numbers for JSON consumers.
func (ResourceSectionBuilder) Build(snap collectors.Snapshot, cfg config.Config) (Section, bool) {
	var items []Item
	data := map[string]any{}

	if mem := snap.Resources.Memory; cfg.Display.Memory && mem.Total > 0 {
		used := mem.Total - min(mem.Available, mem.Total)
		items = append(items, usageItem("Mem", used, mem.Total))
		data["memory_used_percent"] = percent(used, mem.Total)
	}

	if d := snap.Resources.Disk; cfg.Display.Disk && d.Total > 0 {
		items = append(items, usageItem("Disk", d.Used, d.Total))
		data["disk_used_percent"] = percent(d.Used, d.Total)
		if d.Path != "" {
			data["disk_path"] = d.Path
		}
	}

	if cpu := snap.Resources.CPU; cfg.Display.Load {
		switch cpu.Mode {
		case "usage":
			frac := cpu.Usage / 100
			items = append(items, Item{
				Label: "CPU",
				Value: fmt.Sprintf("%.0f%%", cpu.Usage),
				Meter: meter(frac),
				Level: levelFor(frac),
			})
			data["cpu_usage_percent"] = cpu.Usage
		case "load":
			items = append(items, loadItem(cpu))
			data["cpu_load_1"] = cpu.Load1
			data["cpu_load_5"] = cpu.Load5
			data["cpu_load_15"] = cpu.Load15
			if cpu.Cores > 0 {
				data["cpu_cores"] = cpu.Cores
			}
		}
	}

	return newSection("resources", "Resources", items, data)
}

func usageItem(label string, used, total uint64) Item {
	frac := float64(used) / float64(total)
	return Item{
		Label:  label,
		Value:  fmt.Sprintf("%d%%", percent(used, total)),
		Detail: bytePair(used, total),
		Meter:  meter(frac),
		Level:  levelFor(frac),
	}
}

// loadItem shows the 1-minute load against the core count: a load of 4 is
// idle on 32 cores and saturated on 4.
func loadItem(cpu collectors.CPUInfo) Item {
	it := Item{Label: "Load", Value: fmt.Sprintf("%.2f", cpu.Load1)}
	if cpu.Cores > 0 {
		frac := cpu.Load1 / float64(cpu.Cores)
		it.Meter = meter(frac)
		it.Level = levelFor(frac)
		it.Detail = fmt.Sprintf("%d cores", cpu.Cores)
		if cpu.Cores == 1 {
			it.Detail = "1 core"
		}
	}
	return it
}

var byteUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}

// bytePair formats used/total in the unit that suits total, e.g.
// "3.7/16.0 GiB".
func bytePair(used, total uint64) string {
	v := float64(total)
	i := 0
	for v >= 1024 && i < len(byteUnits)-1 {
		v /= 1024
		i++
	}
	scale := math.Pow(1024, float64(i))
	return fmt.Sprintf("%.1f/%.1f %s", float64(used)/scale, v, byteUnits[i])
}

func percent(part, total uint64) int {
	if total == 0 {
		return 0
	}
	return int(math.Round(float64(part) / float64(total) * 100))
}
