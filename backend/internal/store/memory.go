package store

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kebab0o/sysdash/backend/internal/collect"
	"github.com/kebab0o/sysdash/backend/internal/types"
)

var ErrNotFound = errors.New("not found")

const retention = 30 * 24 * time.Hour
const ringCap = 50000

type Memory struct {
	mu sync.RWMutex

	items map[string]*types.Item

	cpuPoints  []collect.CPUPoint
	memPoints  []collect.MemPoint
	diskSeries map[string][]collect.DiskPoint
	diskIO     []collect.DiskIOPoint
	netIO      []collect.NetPoint

	lastCollector time.Time
}

func NewMemory() *Memory {
	return &Memory{
		items:      make(map[string]*types.Item),
		diskSeries: make(map[string][]collect.DiskPoint),
	}
}
func (m *Memory) now() time.Time { return time.Now().UTC() }

/* ===== Items CRUD ===== */
func (m *Memory) List() []*types.Item {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*types.Item, 0, len(m.items))
	for _, it := range m.items {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}
func (m *Memory) Get(id string) (*types.Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	it, ok := m.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *it
	return &cp, nil
}
func (m *Memory) Create(title, notes string) *types.Item {
	now := m.now()
	it := &types.Item{ID: uuid.NewString(), Title: title, Notes: notes, CreatedAt: now, UpdatedAt: now}
	m.mu.Lock()
	m.items[it.ID] = it
	m.mu.Unlock()
	return it
}
func (m *Memory) Update(id, title, notes string) (*types.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if title != "" {
		it.Title = title
	}
	it.Notes = notes
	it.UpdatedAt = m.now()
	return it, nil
}
func (m *Memory) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}

/* ===== Save points ===== */
func (m *Memory) SaveCPU(p collect.CPUPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cpuPoints = appendCapCPU(m.cpuPoints, p)
	return nil
}
func (m *Memory) SaveMem(p collect.MemPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memPoints = appendCapMem(m.memPoints, p)
	return nil
}
func (m *Memory) SaveDisk(p collect.DiskPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	series := append(m.diskSeries[p.Mount], p)
	if len(series) > ringCap {
		series = series[len(series)-ringCap:]
	}
	m.diskSeries[p.Mount] = series
	return nil
}
func (m *Memory) SaveDiskIO(p collect.DiskIOPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.diskIO = appendCapDiskIO(m.diskIO, p)
	return nil
}
func (m *Memory) SaveNet(p collect.NetPoint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.netIO = appendCapNet(m.netIO, p)
	return nil
}

func appendCapCPU(s []collect.CPUPoint, v collect.CPUPoint) []collect.CPUPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}
func appendCapMem(s []collect.MemPoint, v collect.MemPoint) []collect.MemPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}
func appendCapDiskIO(s []collect.DiskIOPoint, v collect.DiskIOPoint) []collect.DiskIOPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}
func appendCapNet(s []collect.NetPoint, v collect.NetPoint) []collect.NetPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}

/* ===== Queries ===== */
func (m *Memory) CPUSince(since time.Time) []collect.CPUPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []collect.CPUPoint
	for _, p := range m.cpuPoints {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}
func (m *Memory) MemSince(since time.Time) []collect.MemPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []collect.MemPoint
	for _, p := range m.memPoints {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}

type DiskSeries struct {
	Mount  string              `json:"mount"`
	Points []collect.DiskPoint `json:"points"`
}

func (m *Memory) DiskSince(since time.Time) []DiskSeries {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]DiskSeries, 0, len(m.diskSeries))
	for mount, series := range m.diskSeries {
		var pts []collect.DiskPoint
		for _, p := range series {
			if !p.At.Before(since) {
				pts = append(pts, p)
			}
		}
		if len(pts) > 0 {
			res = append(res, DiskSeries{Mount: mount, Points: pts})
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Mount < res[j].Mount })
	return res
}
func (m *Memory) DiskIOSince(since time.Time) []collect.DiskIOPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []collect.DiskIOPoint
	for _, p := range m.diskIO {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}
func (m *Memory) NetSince(since time.Time) []collect.NetPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []collect.NetPoint
	for _, p := range m.netIO {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}

/* ===== housekeeping ===== */
func (m *Memory) PruneOlderThan(cutoff time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cpu := m.cpuPoints[:0]
	for _, p := range m.cpuPoints {
		if !p.At.Before(cutoff) {
			cpu = append(cpu, p)
		}
	}
	m.cpuPoints = cpu
	mem := m.memPoints[:0]
	for _, p := range m.memPoints {
		if !p.At.Before(cutoff) {
			mem = append(mem, p)
		}
	}
	m.memPoints = mem
	for k, series := range m.diskSeries {
		dst := series[:0]
		for _, p := range series {
			if !p.At.Before(cutoff) {
				dst = append(dst, p)
			}
		}
		m.diskSeries[k] = dst
	}
	dio := m.diskIO[:0]
	for _, p := range m.diskIO {
		if !p.At.Before(cutoff) {
			dio = append(dio, p)
		}
	}
	m.diskIO = dio
	net := m.netIO[:0]
	for _, p := range m.netIO {
		if !p.At.Before(cutoff) {
			net = append(net, p)
		}
	}
	m.netIO = net
	return nil
}
func (m *Memory) PruneForRetention() { _ = m.PruneOlderThan(time.Now().Add(-retention)) }

func (m *Memory) LastCollector() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCollector
}
func (m *Memory) SetLastCollector(t time.Time) { m.mu.Lock(); m.lastCollector = t; m.mu.Unlock() }
