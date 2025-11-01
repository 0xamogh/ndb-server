package core

import (
	"context"
	"errors"
	"sync"
)

var ErrOutOfBounds = errors.New("out of bounds")

type MemDevice struct {
	size     int64
	pageSize int64

	mu    sync.RWMutex
	pages map[uint64][]byte
	dirty map[uint64]struct{}
}

func NewMemDevice(size int64, pageSize int64) *MemDevice {
	return &MemDevice{
		size:     size,
		pageSize: pageSize,
		pages:    make(map[uint64][]byte),
		dirty:    make(map[uint64]struct{}),
	}
}

func (m *MemDevice) Size() int64 { return m.size }

func (m *MemDevice) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= m.size {
		return 0, ErrOutOfBounds
	}
	if int64(len(p)) > m.size-off {
		p = p[:m.size-off]
	}
	n := 0
	for remainingBytes := len(p); remainingBytes > 0; {
		currentIndex := uint64((off) / m.pageSize)
		inPage := int(off % m.pageSize)
		toCopy := int(m.pageSize) - inPage
		if toCopy > remainingBytes {
			toCopy = remainingBytes
		}

		m.mu.RLock()
		page := m.pages[currentIndex]
		m.mu.RUnlock()

		if page == nil {
			for i := 0; i < toCopy; i++ {
				p[n+i] = 0
			}
		} else {
			copy(p[n:n+toCopy], page[inPage:inPage+toCopy])
		}

		n += toCopy
		off += int64(toCopy)
		remainingBytes -= toCopy
	}
	return n, nil
}

func (m *MemDevice) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= m.size {
		return 0, ErrOutOfBounds
	}
	if int64(len(p)) > m.size-off {
		p = p[:m.size-off]
	}

	n := 0
	for remainingBytes := len(p); remainingBytes > 0; {
		currentIndex := uint64(off / m.pageSize)
		inPage := int(off % m.pageSize)
		toCopy := int(m.pageSize) - inPage
		if toCopy > remainingBytes {
			toCopy = remainingBytes
		}

		m.mu.RLock()
		page := m.pages[currentIndex]
		m.mu.RUnlock()

		if page == nil {
			page = make([]byte, int(m.pageSize))
			m.mu.Lock()
			// re-check in case another writer created it
			if existing := m.pages[currentIndex]; existing == nil {
				m.pages[currentIndex] = page
			} else {
				page = existing
			}
			m.mu.Unlock()
		}

		copy(page[inPage:inPage+toCopy], p[n:n+toCopy])

		n += toCopy
		off += int64(toCopy)
		remainingBytes -= toCopy
	}
	return n, nil
}

func (m *MemDevice) Flush(ctx context.Context) error {
	// RAM-only: mark clean. (Durable store will upload dirty pages here.)
	m.mu.Lock()
	for k := range m.dirty {
		delete(m.dirty, k)
	}
	m.mu.Unlock()
	return nil
}
