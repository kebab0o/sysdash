package store

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kebab0o/sysdash/backend/internal/types"
)

var ErrNotFound = errors.New("not found")

const retention = 30 * 24 * time.Hour
const ringCap = 50000

type Memory struct {
	mu sync.RWMutex

	items map[string]*types.Item

	cpuPoints  []types.CPUPoint
	memPoints  []types.MemPoint
	diskSeries map[string][]types.DiskPoint
	diskIO     []types.DiskIOPoint
	netIO      []types.NetPoint

	logs  []LogEntry
	tasks map[string]*Task

	lastCollector time.Time
}

func NewMemory() *Memory {
	return &Memory{
		items:      make(map[string]*types.Item),
		diskSeries: make(map[string][]types.DiskPoint),
		tasks:      make(map[string]*Task),
	}
}
func (m *Memory) now() time.Time { return time.Now().UTC() }

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

func (m *Memory) SaveCPU(p types.CPUPoint) error {
	m.mu.Lock()
	m.cpuPoints = appendCapCPU(m.cpuPoints, p)
	m.mu.Unlock()
	return nil
}
func (m *Memory) SaveMem(p types.MemPoint) error {
	m.mu.Lock()
	m.memPoints = appendCapMem(m.memPoints, p)
	m.mu.Unlock()
	return nil
}
func (m *Memory) SaveDisk(p types.DiskPoint) error {
	m.mu.Lock()
	series := append(m.diskSeries[p.Mount], p)
	if len(series) > ringCap {
		series = series[len(series)-ringCap:]
	}
	m.diskSeries[p.Mount] = series
	m.mu.Unlock()
	return nil
}
func (m *Memory) SaveDiskIO(p types.DiskIOPoint) error {
	m.mu.Lock()
	m.diskIO = appendCapDiskIO(m.diskIO, p)
	m.mu.Unlock()
	return nil
}
func (m *Memory) SaveNet(p types.NetPoint) error {
	m.mu.Lock()
	m.netIO = appendCapNet(m.netIO, p)
	m.mu.Unlock()
	return nil
}

func appendCapCPU(s []types.CPUPoint, v types.CPUPoint) []types.CPUPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}
func appendCapMem(s []types.MemPoint, v types.MemPoint) []types.MemPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}
func appendCapDiskIO(s []types.DiskIOPoint, v types.DiskIOPoint) []types.DiskIOPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}
func appendCapNet(s []types.NetPoint, v types.NetPoint) []types.NetPoint {
	s = append(s, v)
	if len(s) > ringCap {
		return s[len(s)-ringCap:]
	}
	return s
}

func (m *Memory) CPUSince(since time.Time) []types.CPUPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []types.CPUPoint
	for _, p := range m.cpuPoints {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}
func (m *Memory) MemSince(since time.Time) []types.MemPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []types.MemPoint
	for _, p := range m.memPoints {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}

type DiskSeries struct {
	Mount  string            `json:"mount"`
	Points []types.DiskPoint `json:"points"`
}

func (m *Memory) DiskSince(since time.Time) []DiskSeries {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]DiskSeries, 0, len(m.diskSeries))
	for mount, series := range m.diskSeries {
		var pts []types.DiskPoint
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
func (m *Memory) DiskIOSince(since time.Time) []types.DiskIOPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []types.DiskIOPoint
	for _, p := range m.diskIO {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}
func (m *Memory) NetSince(since time.Time) []types.NetPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []types.NetPoint
	for _, p := range m.netIO {
		if !p.At.Before(since) {
			out = append(out, p)
		}
	}
	return out
}

func (m *Memory) PruneOlderThan(cutoff time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dstCPU := m.cpuPoints[:0]
	for _, p := range m.cpuPoints {
		if !p.At.Before(cutoff) {
			dstCPU = append(dstCPU, p)
		}
	}
	m.cpuPoints = dstCPU

	dstMem := m.memPoints[:0]
	for _, p := range m.memPoints {
		if !p.At.Before(cutoff) {
			dstMem = append(dstMem, p)
		}
	}
	m.memPoints = dstMem

	for k, series := range m.diskSeries {
		dst := series[:0]
		for _, p := range series {
			if !p.At.Before(cutoff) {
				dst = append(dst, p)
			}
		}
		m.diskSeries[k] = dst
	}

	dstIO := m.diskIO[:0]
	for _, p := range m.diskIO {
		if !p.At.Before(cutoff) {
			dstIO = append(dstIO, p)
		}
	}
	m.diskIO = dstIO

	dstNet := m.netIO[:0]
	for _, p := range m.netIO {
		if !p.At.Before(cutoff) {
			dstNet = append(dstNet, p)
		}
	}
	m.netIO = dstNet

	return nil
}
func (m *Memory) PruneForRetention() { _ = m.PruneOlderThan(time.Now().Add(-retention)) }

func (m *Memory) LastCollector() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCollector
}
func (m *Memory) SetLastCollector(t time.Time) { m.mu.Lock(); m.lastCollector = t; m.mu.Unlock() }

type LogEntry struct {
	At    time.Time `json:"t"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

func (m *Memory) addLog(level, msg string) {
	m.logs = append(m.logs, LogEntry{At: time.Now().UTC(), Level: level, Msg: msg})
	if len(m.logs) > ringCap {
		m.logs = m.logs[len(m.logs)-ringCap:]
	}
}
func (m *Memory) ListLogs(limit int, filter string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	filter = strings.ToLower(strings.TrimSpace(filter))
	n := len(m.logs)
	out := make([]LogEntry, 0, min(limit, n))
	for i := n - 1; i >= 0 && len(out) < limit; i-- {
		if filter != "" && !strings.Contains(strings.ToLower(m.logs[i].Msg), filter) {
			continue
		}
		out = append(out, m.logs[i])
	}
	return out
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Task struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	EveryMinutes int       `json:"everyMinutes"`
	LastRun      time.Time `json:"lastRun"`
	Status       string    `json:"status"`
	Enabled      bool      `json:"enabled"`
}

func (m *Memory) ListTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func (m *Memory) CreateTask(name string, every int) *Task {
	t := &Task{ID: uuid.NewString(), Name: name, EveryMinutes: every, Enabled: true}
	m.mu.Lock()
	m.tasks[t.ID] = t
	m.addLog("INFO", "task created: "+name)
	m.mu.Unlock()
	return t
}
func (m *Memory) DeleteTask(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.tasks, id)
	m.addLog("INFO", "task deleted: "+t.Name)
	return nil
}

func (m *Memory) RunTaskNow(id string) error {
	m.mu.Lock()
	t, ok := m.tasks[id]
	m.mu.Unlock()
	if !ok {
		return ErrNotFound
	}

	name := strings.ToLower(strings.TrimSpace(t.Name))
	var err error
	switch {
	case strings.Contains(name, "dns"):
		err = flushDNS()
	case strings.Contains(name, "temp"), strings.Contains(name, "cache"):
		err = clearTemp()
	default:
		err = clearTemp()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		t.Status = "ERR"
		t.LastRun = time.Now().UTC()
		m.addLog("ERROR", "task failed: "+t.Name+" ("+err.Error()+")")
		return err
	}
	t.Status = "OK"
	t.LastRun = time.Now().UTC()
	m.addLog("INFO", "task ran: "+t.Name)
	return nil
}

func (m *Memory) StartScheduler(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.mu.RLock()
				now := time.Now().UTC()
				for id, t := range m.tasks {
					if !t.Enabled || t.EveryMinutes <= 0 {
						continue
					}
					if t.LastRun.IsZero() || now.Sub(t.LastRun) >= time.Duration(t.EveryMinutes)*time.Minute {
						go func(id string) { _ = m.RunTaskNow(id) }(id)
					}
				}
				m.mu.RUnlock()
			case <-stop:
				return
			}
		}
	}()
}

func clearTemp() error {
	base := os.TempDir()
	entries, err := os.ReadDir(base)
	if err != nil {
		return err
	}
	for _, e := range entries {
		full := filepath.Join(base, e.Name())
		if e.IsDir() {
			switch strings.ToLower(e.Name()) {
			case "temp", "cache", "tmp":
				_ = os.RemoveAll(full)
			}
		} else {
			_ = os.Remove(full)
		}
	}
	_ = runtime.GOOS
	return nil
}

func flushDNS() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("ipconfig", "/flushdns")
		return cmd.Run()
	case "darwin":
		cmd := exec.Command("sh", "-c", "dscacheutil -flushcache; killall -HUP mDNSResponder")
		return cmd.Run()
	default:
		if err := exec.Command("sh", "-c", "resolvectl flush-caches || resolvectl reload").Run(); err == nil {
			return nil
		}
		if err := exec.Command("sh", "-c", "nscd -i hosts").Run(); err == nil {
			return nil
		}
		return errors.New("dns flush not supported without elevated permissions")
	}
}
