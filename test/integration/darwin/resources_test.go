//go:build darwin

package darwin

import (
	"context"
	"math"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/veteranbv/sysgreet/internal/collectors"
)

func TestResourceCollectorMatchesSystemStats(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-specific accuracy test")
	}
	ctx := context.Background()
	collector := collectors.NewResourceCollector()
	info, err := collector.CollectResources(ctx)
	if err != nil {
		t.Fatalf("CollectResources error: %v", err)
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		t.Fatalf("VirtualMemory error: %v", err)
	}
	if !withinTolerance(float64(info.Memory.Total), float64(vm.Total), 0.05) {
		t.Fatalf("memory total mismatch: got %d expected approx %d", info.Memory.Total, vm.Total)
	}

	if info.Disk.Path != "/System/Volumes/Data" {
		t.Fatalf("disk path %q: want the writable data volume, not the sealed system snapshot", info.Disk.Path)
	}
	usage, err := disk.UsageWithContext(ctx, info.Disk.Path)
	if err != nil {
		t.Fatalf("Disk usage error: %v", err)
	}
	// The collector reports the filesystem the way df does: usable
	// capacity is used plus available, excluding reserved blocks.
	if want := usage.Used + usage.Free; !withinTolerance(float64(info.Disk.Total), float64(want), 0.10) {
		t.Fatalf("disk total mismatch got %d expected approx %d", info.Disk.Total, want)
	}
}

func withinTolerance(observed, expected, tolerance float64) bool {
	if expected == 0 {
		return false
	}
	return math.Abs(observed-expected)/expected <= tolerance
}
