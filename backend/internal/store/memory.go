package store

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kebab0o/sysdash/backend/internal/types"
)

var ErrNotFound = errors.New("not found")

type Memory struct {
	mu    sync.RWMutex
	items map[string]*types.Item
}

func NewMemory() *Memory {
	return &Memory{items: make(map[string]*types.Item)}
}

func (m *Memory) List() []*types.Item {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*types.Item, 0, len(m.items))
	for _, it := range m.items {
		out = append(out, it)
	}
	return out
}

func (m *Memory) Get(id string) (*types.Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	it, ok := m.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	return it, nil
}

func (m *Memory) Create(title, notes string) *types.Item {
	now := time.Now().UTC()
	it := &types.Item{
		ID:        uuid.NewString(),
		Title:     title,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
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
	it.UpdatedAt = time.Now().UTC()
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
