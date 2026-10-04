package collectors

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// DefaultResourceCollector gathers memory, disk, and CPU metrics.
type DefaultResourceCollector struct{}

// NewResourceCollector instantiates the default resource collector.
func NewResourceCollector() ResourceCollector {
	return DefaultResourceCollector{}
}

// CollectResources implements ResourceCollector with graceful degradation.
func (DefaultResourceCollector) CollectResources(ctx context.Context) (ResourceInfo, error) {
	var info ResourceInfo

	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		info.Memory = MemoryInfo{Total: vm.Total, Available: vm.Available}
	} else {
		recordError("memory", err)
	}

	path := diskPath()
	if usage, err := disk.UsageWithContext(ctx, path); err == nil {
		// gopsutil's Free is space available to ordinary users and Used
		// excludes root-reserved blocks, matching df; Total would count
		// the reserve and overstate usage by its size.
		info.Disk = DiskInfo{Path: path, Total: usage.Used + usage.Free, Used: usage.Used}
	} else {
		recordError("disk", err)
	}

	if runtime.GOOS == "windows" {
		values, err := cpu.PercentWithContext(ctx, time.Millisecond*100, false)
		if err != nil {
			recordError("cpu", err)
		} else if len(values) > 0 {
			info.CPU = CPUInfo{Usage: values[0], Cores: runtime.NumCPU(), Mode: "usage"}
		}
	} else {
		avg, err := load.AvgWithContext(ctx)
		if err != nil {
			recordError("cpu", err)
		} else {
			info.CPU = CPUInfo{Load1: avg.Load1, Load5: avg.Load5, Load15: avg.Load15, Cores: runtime.NumCPU(), Mode: "load"}
		}
	}
	return info, nil
}

// diskPath is the filesystem the banner reports: the root filesystem, or the
// system drive on Windows. It holds the OS and usually the logs, so it is
// the one that fills up and breaks the host. On macOS "/" is the sealed,
// read-only system snapshot; the writable data volume is what fills up.
func diskPath() string {
	switch runtime.GOOS {
	case "windows":
		if drive := os.Getenv("SystemDrive"); drive != "" {
			return drive + `\`
		}
		return `C:\`
	case "darwin":
		if info, err := os.Stat(macDataVolume); err == nil && info.IsDir() {
			return macDataVolume
		}
	}
	return "/"
}

const macDataVolume = "/System/Volumes/Data"
