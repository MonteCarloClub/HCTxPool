// Package metrics provides general system and process level metrics collection.
package metrics

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rcrowley/go-metrics"
)

// MetricsEnabledFlag is the CLI flag name to use to enable metrics collections.
var MetricsEnabledFlag = "metrics"

// enabled is the flag specifying if metrics are enable or not.
var enabled = false

// Init enables or disables the metrics system. Since we need this to run before
// any other code gets to create meters and timers, we'll actually do an ugly hack
// and peek into the command line args for the metrics flag.
func init() {
	for _, arg := range os.Args {
		if strings.TrimLeft(arg, "-") == MetricsEnabledFlag {
			klog.Infof("Enabling metrics collection")
			enabled = true
		}
	}
}

// NewMeter create a new metrics Meter, either a real one of a NOP stub depending
// on the metrics flag.
func NewMeter(name string) metrics.Meter {
	if !enabled {
		return new(metrics.NilMeter)
	}
	return metrics.GetOrRegisterMeter(name, metrics.DefaultRegistry)
}

// NewTimer create a new metrics Timer, either a real one of a NOP stub depending
// on the metrics flag.
func NewTimer(name string) metrics.Timer {
	if !enabled {
		return new(metrics.NilTimer)
	}
	return metrics.GetOrRegisterTimer(name, metrics.DefaultRegistry)
}

// CollectProcessMetrics periodically collects various metrics about the running
// process.
func CollectProcessMetrics(refresh time.Duration) {
	// Short circuit if the metrics system is disabled
	if !enabled {
		return
	}
	// Create the various data collectors
	memstats := make([]*runtime.MemStats, 2)
	diskstats := make([]*DiskStats, 2)
	for i := 0; i < len(memstats); i++ {
		memstats[i] = new(runtime.MemStats)
		diskstats[i] = new(DiskStats)
	}
	// Define the various metrics to collect
	memAllocs := metrics.GetOrRegisterMeter("system/memory/allocs", metrics.DefaultRegistry)
	memFrees := metrics.GetOrRegisterMeter("system/memory/frees", metrics.DefaultRegistry)
	memInuse := metrics.GetOrRegisterMeter("system/memory/inuse", metrics.DefaultRegistry)
	memPauses := metrics.GetOrRegisterMeter("system/memory/pauses", metrics.DefaultRegistry)

	var diskReads, diskReadBytes, diskWrites, diskWriteBytes metrics.Meter
	if err := ReadDiskStats(diskstats[0]); err == nil {
		diskReads = metrics.GetOrRegisterMeter("system/disk/readcount", metrics.DefaultRegistry)
		diskReadBytes = metrics.GetOrRegisterMeter("system/disk/readdata", metrics.DefaultRegistry)
		diskWrites = metrics.GetOrRegisterMeter("system/disk/writecount", metrics.DefaultRegistry)
		diskWriteBytes = metrics.GetOrRegisterMeter("system/disk/writedata", metrics.DefaultRegistry)
	} else {
		klog.Error("failed to read disk metrics: %v", err)
	}
	// Iterate loading the different stats and updating the meters
	for i := 1; ; i++ {
		runtime.ReadMemStats(memstats[i%2])
		memAllocs.Mark(int64(memstats[i%2].Mallocs - memstats[(i-1)%2].Mallocs))
		memFrees.Mark(int64(memstats[i%2].Frees - memstats[(i-1)%2].Frees))
		memInuse.Mark(int64(memstats[i%2].Alloc - memstats[(i-1)%2].Alloc))
		memPauses.Mark(int64(memstats[i%2].PauseTotalNs - memstats[(i-1)%2].PauseTotalNs))

		if ReadDiskStats(diskstats[i%2]) == nil {
			diskReads.Mark(int64(diskstats[i%2].ReadCount - diskstats[(i-1)%2].ReadCount))
			diskReadBytes.Mark(int64(diskstats[i%2].ReadBytes - diskstats[(i-1)%2].ReadBytes))
			diskWrites.Mark(int64(diskstats[i%2].WriteCount - diskstats[(i-1)%2].WriteCount))
			diskWriteBytes.Mark(int64(diskstats[i%2].WriteBytes - diskstats[(i-1)%2].WriteBytes))
		}
		time.Sleep(refresh)
	}
}
