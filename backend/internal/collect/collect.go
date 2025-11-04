package collect

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type CPUPoint struct {
	At    time.Time `json:"t"`
	Usage float64   `json:"v"`
}
type MemPoint struct {
	At      time.Time `json:"t"`
	UsedPct float64   `json:"v"`
}
type DiskPoint struct {
	At      time.Time `json:"t"`
	Mount   string    `json:"mount"`
	UsedPct float64   `json:"usedPct"`
	UsedGB  float64   `json:"usedGB"`
	TotalGB float64   `json:"totalGB"`
}
type DiskIOPoint struct {
	At       time.Time `json:"t"`
	ReadMBs  float64   `json:"readMBs"`
	WriteMBs float64   `json:"writeMBs"`
}
type NetPoint struct {
	At    time.Time `json:"t"`
	RXkBs float64   `json:"rxKBs"`
	TXkBs float64   `json:"txKBs"`
}

type Sink interface {
	SaveCPU(CPUPoint) error
	SaveMem(MemPoint) error
	SaveDisk(DiskPoint) error
	SaveDiskIO(DiskIOPoint) error
	SaveNet(NetPoint) error
	PruneOlderThan(time.Time) error
	LastCollector() time.Time
	SetLastCollector(time.Time)
}

var (
	prevDiskRead  uint64
	prevDiskWrite uint64
	prevNetRx     uint64
	prevNetTx     uint64
	prevAt        time.Time
)

func Start(ctx context.Context, sink Sink, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	collectOnce(sink, interval) // prime immediately
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			collectOnce(sink, interval)
		}
	}
}

func collectOnce(sink Sink, interval time.Duration) {
	now := time.Now().UTC()

	if err := collectCPU(sink, now); err != nil {
		log.Printf("cpu: %v", err)
	}
	if err := collectMem(sink, now); err != nil {
		log.Printf("mem: %v", err)
	}
	if err := collectDiskUsage(sink, now); err != nil {
		log.Printf("disk: %v", err)
	}
	if err := collectDiskIO(sink, now, interval); err != nil {
		log.Printf("diskio: %v", err)
	}
	if err := collectNet(sink, now, interval); err != nil {
		log.Printf("net: %v", err)
	}

	sink.SetLastCollector(now)
}

func collectCPU(sink Sink, at time.Time) error {
	pcts, err := cpu.Percent(time.Second, false)
	if err != nil || len(pcts) == 0 {
		return err
	}
	return sink.SaveCPU(CPUPoint{At: at, Usage: pcts[0]})
}

func collectMem(sink Sink, at time.Time) error {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	return sink.SaveMem(MemPoint{At: at, UsedPct: vm.UsedPercent})
}

func collectDiskUsage(sink Sink, at time.Time) error {
	parts, err := disk.Partitions(true)
	if err != nil {
		return err
	}
	gb := func(b uint64) float64 { return float64(b) / (1024 * 1024 * 1024) }
	for _, p := range parts {
		u, err := disk.Usage(p.Mountpoint)
		if err != nil || u.Total == 0 {
			continue
		}
		_ = sink.SaveDisk(DiskPoint{
			At: at, Mount: p.Mountpoint, UsedPct: u.UsedPercent, UsedGB: gb(u.Used), TotalGB: gb(u.Total),
		})
	}
	return nil
}

// simple per-interval rate calc using deltas
func collectDiskIO(sink Sink, at time.Time, interval time.Duration) error {
	stats, err := disk.IOCounters()
	if err != nil || len(stats) == 0 {
		return err
	}
	var r, w uint64
	for _, v := range stats {
		r += v.ReadBytes
		w += v.WriteBytes
	}
	if !prevAt.IsZero() {
		sec := at.Sub(prevAt).Seconds()
		readMBs := (float64(r-prevDiskRead) / (1024 * 1024)) / sec
		writeMBs := (float64(w-prevDiskWrite) / (1024 * 1024)) / sec
		_ = sink.SaveDiskIO(DiskIOPoint{At: at, ReadMBs: readMBs, WriteMBs: writeMBs})
	} else {
		_ = sink.SaveDiskIO(DiskIOPoint{At: at, ReadMBs: 0, WriteMBs: 0})
	}
	prevDiskRead, prevDiskWrite, prevAt = r, w, at
	return nil
}

func collectNet(sink Sink, at time.Time, interval time.Duration) error {
	stats, err := net.IOCounters(false) // aggregate
	if err != nil || len(stats) == 0 {
		return err
	}
	s := stats[0]
	if !prevAt.IsZero() {
		sec := at.Sub(prevAt).Seconds()
		rxKBs := (float64(s.BytesRecv-prevNetRx) / 1024) / sec
		txKBs := (float64(s.BytesSent-prevNetTx) / 1024) / sec
		_ = sink.SaveNet(NetPoint{At: at, RXkBs: rxKBs, TXkBs: txKBs})
	} else {
		_ = sink.SaveNet(NetPoint{At: at, RXkBs: 0, TXkBs: 0})
	}
	prevNetRx, prevNetTx = s.BytesRecv, s.BytesSent
	return nil
}
