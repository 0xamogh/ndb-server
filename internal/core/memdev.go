package core

import (
	"context"
	"errors"
	"sync"

	"nbds3d/internal/store"
)

var (
	ErrOutOfBounds = errors.New("out of bounds")
)

type MemDevice struct {
	export   string
	size     int64
	pageSize uint64

	mu    sync.RWMutex
	pages map[uint64][]byte
	dirty map[uint64]bool

	st store.Store
}

func NewMemDevice(export string, size int64, pageSize uint64, st store.Store) *MemDevice {
	if pageSize == 0 {
		pageSize = 4096
	}
	return &MemDevice{
		export:   export,
		size:     size,
		pageSize: pageSize,
		pages:    make(map[uint64][]byte),
		dirty:    make(map[uint64]bool),
		st:       st,
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
		currentIndex := uint64(off / int64(m.pageSize))
		inPage := int(off % int64(m.pageSize))
		toCopy := int(m.pageSize) - inPage
		if toCopy > remainingBytes {
			toCopy = remainingBytes
		}

		m.mu.RLock()
		page := m.pages[currentIndex]
		m.mu.RUnlock()

		if page == nil && m.st != nil {
			buf, err := m.st.ReadPage(context.Background(), store.PageAddress{
				Export: m.export, Index: currentIndex, Size: m.pageSize,
			})
			if err != nil {
				return n, err
			}
			m.mu.Lock()
			if existing := m.pages[currentIndex]; existing == nil {
				page = buf
				m.pages[currentIndex] = page
			} else {
				page = existing
			}
			m.mu.Unlock()
		}

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
		currentIndex := uint64(off / int64(m.pageSize))
		inPage := int(off % int64(m.pageSize))
		toCopy := int(m.pageSize) - inPage
		if toCopy > remainingBytes {
			toCopy = remainingBytes
		}

		m.mu.RLock()
		page := m.pages[currentIndex]
		m.mu.RUnlock()

		if page == nil {
			if m.st != nil {
				buf, err := m.st.ReadPage(context.Background(), store.PageAddress{
					Export: m.export, Index: currentIndex, Size: m.pageSize,
				})
				if err != nil {
					return n, err
				}
				page = buf
			} else {
				page = make([]byte, int(m.pageSize))
			}
			m.mu.Lock()
			if existing := m.pages[currentIndex]; existing == nil {
				m.pages[currentIndex] = page
			} else {
				page = existing
			}
			m.mu.Unlock()
		}

		copy(page[inPage:inPage+toCopy], p[n:n+toCopy])

		m.mu.Lock()
		m.dirty[currentIndex] = true
		m.mu.Unlock()

		n += toCopy
		off += int64(toCopy)
		remainingBytes -= toCopy
	}
	return n, nil
}

func (m *MemDevice) Flush(ctx context.Context) error {
	if m.st == nil {
		m.mu.Lock()
		for k := range m.dirty {
			delete(m.dirty, k)
		}
		m.mu.Unlock()
		return nil
	}

	type item struct {
		idx  uint64
		data []byte
	}
	var batch []item

	m.mu.RLock()
	for idx := range m.dirty {
		if pg := m.pages[idx]; pg != nil {
			cp := make([]byte, len(pg))
			copy(cp, pg)
			batch = append(batch, item{idx: idx, data: cp})
		}
	}
	m.mu.RUnlock()

	for _, it := range batch {
		addr := store.PageAddress{Export: m.export, Index: it.idx, Size: m.pageSize}
		if err := m.st.WritePage(ctx, addr, it.data); err != nil {
			return err
		}
	}

	m.mu.Lock()
	for _, it := range batch {
		delete(m.dirty, it.idx)
	}
	m.mu.Unlock()

	return m.st.FlushExport(ctx, m.export)
}
