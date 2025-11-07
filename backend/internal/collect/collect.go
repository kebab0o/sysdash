package collect

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/kebab0o/sysdash/backend/internal/types"
)

type Saver interface {
	SaveCPU(types.CPUPoint) error
	SaveMem(types.MemPoint) error
	SaveDisk(types.DiskPoint) error
	SaveDiskIO(types.DiskIOPoint) error
	SaveNet(types.NetPoint) error
	SetLastCollector(time.Time)
}

var (
	mu sync.Mutex

	lastDiskReadBytes  uint64
	lastDiskWriteBytes uint64
	lastDiskAt         time.Time
	haveDiskBaseline   bool

	lastNetRxBytes  uint64
	lastNetTxBytes  uint64
	lastNetAt       time.Time
	haveNetBaseline bool
)

func Start(ctx context.Context, s Saver, period time.Duration) {
	mu.Lock()
	lastDiskAt = time.Now()
	lastNetAt = lastDiskAt
	haveDiskBaseline = false
	haveNetBaseline = false
	mu.Unlock()

	t := time.NewTicker(period)
	defer t.Stop()

	sample(s, period)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			sample(s, period)
		}
	}
}

func sample(s Saver, period time.Duration) {
	now := time.Now().UTC()

	if vals, err := cpu.Percent(0, false); err == nil && len(vals) > 0 {
		_ = s.SaveCPU(types.CPUPoint{At: now, V: vals[0]})
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		_ = s.SaveMem(types.MemPoint{At: now, V: vm.UsedPercent})
	}

	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			mount := normalizeMount(p.Mountpoint)
			if mount == "" {
				continue
			}
			if u, err := disk.Usage(mount); err == nil && u.Total > 0 {
				_ = s.SaveDisk(types.DiskPoint{
					At:      now,
					Mount:   mount,
					UsedPct: u.UsedPercent,
					UsedGB:  bytesToGB(u.Used),
					TotalGB: bytesToGB(u.Total),
				})
			}
		}
	}

	rd, wd := diskIOMetrics(period)
	_ = s.SaveDiskIO(types.DiskIOPoint{At: now, ReadMBs: rd, WriteMBs: wd})

	rx, tx := netIOMetrics(period)
	_ = s.SaveNet(types.NetPoint{At: now, RxKBs: rx, TxKBs: tx})

	s.SetLastCollector(now)
}

func diskIOMetrics(period time.Duration) (readMBs float64, writeMBs float64) {
	stats, err := disk.IOCounters()
	if err != nil || len(stats) == 0 {
		return 0, 0
	}
	var totalRead, totalWrite uint64
	for _, v := range stats {
		totalRead += v.ReadBytes
		totalWrite += v.WriteBytes
	}

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	sec := now.Sub(lastDiskAt).Seconds()
	if sec <= 0 {
		sec = period.Seconds()
	}

	if !haveDiskBaseline {
		lastDiskReadBytes = totalRead
		lastDiskWriteBytes = totalWrite
		lastDiskAt = now
		haveDiskBaseline = true
		return 0, 0
	}

	rd := float64(totalRead-lastDiskReadBytes) / (1024.0 * 1024.0) / sec
	wd := float64(totalWrite-lastDiskWriteBytes) / (1024.0 * 1024.0) / sec
	if rd < 0 {
		rd = 0
	}
	if wd < 0 {
		wd = 0
	}

	lastDiskReadBytes = totalRead
	lastDiskWriteBytes = totalWrite
	lastDiskAt = now
	return rd, wd
}

func netIOMetrics(period time.Duration) (rxKBs float64, txKBs float64) {
	stats, err := net.IOCounters(false)
	if err != nil || len(stats) == 0 {
		return 0, 0
	}
	rx := stats[0].BytesRecv
	tx := stats[0].BytesSent

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	sec := now.Sub(lastNetAt).Seconds()
	if sec <= 0 {
		sec = period.Seconds()
	}

	if !haveNetBaseline {
		lastNetRxBytes = rx
		lastNetTxBytes = tx
		lastNetAt = now
		haveNetBaseline = true
		return 0, 0
	}

	rxRate := float64(rx-lastNetRxBytes) / 1024.0 / sec
	txRate := float64(tx-lastNetTxBytes) / 1024.0 / sec
	if rxRate < 0 {
		rxRate = 0
	}
	if txRate < 0 {
		txRate = 0
	}

	lastNetRxBytes = rx
	lastNetTxBytes = tx
	lastNetAt = now
	return rxRate, txRate
}

func bytesToGB(b uint64) float64 { return float64(b) / 1024.0 / 1024.0 / 1024.0 }

func normalizeMount(m string) string {
	if m == "" {
		return ""
	}

	if runtime.GOOS == "windows" {
		m = strings.TrimSpace(m)
		m = filepath.Clean(m)
		if len(m) >= 2 && strings.HasSuffix(m, `:`) {
			m += `\`
		}
	}
	return m
}
