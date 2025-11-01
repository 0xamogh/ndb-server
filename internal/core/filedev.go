package core

import (
	"context"
	"sync"

	"nbds3d/internal/store"
)

type FileDevice struct {
	name     string
	size     int64
	pageSize uint64
	st       *store.FSStore

	mu    sync.RWMutex
	pages map[uint64][]byte
	dirty map[uint64]bool
}

func NewFileDevice(name string, size int64, pageSize uint64, st *store.FSStore) *FileDevice {
	return &FileDevice{
		name:     name,
		size:     size,
		pageSize: pageSize,
		st:       st,
		pages:    make(map[uint64][]byte),
		dirty:    make(map[uint64]bool),
	}
}

func (f *FileDevice) Size() int64 { return f.size }

func (f *FileDevice) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= f.size {
		return 0, ErrOutOfBounds
	}
	if int64(len(p)) > f.size-off {
		p = p[:f.size-off]
	}
	n := 0
	for remaining := len(p); remaining > 0; {
		idx := uint64(off) / f.pageSize
		inPage := int(off % int64(f.pageSize))
		toCopy := int(f.pageSize) - inPage
		if toCopy > remaining {
			toCopy = remaining
		}

		f.mu.RLock()
		page := f.pages[idx]
		f.mu.RUnlock()

		if page == nil {
			addr := store.PageAddress{Export: f.name, Index: idx, Size: f.pageSize}
			data, err := f.st.ReadPage(context.Background(), addr)
			if err != nil {
				return n, err
			}
			f.mu.Lock()
			if existing := f.pages[idx]; existing == nil {
				f.pages[idx] = data
				page = data
			} else {
				page = existing
			}
			f.mu.Unlock()
		}

		copy(p[n:n+toCopy], page[inPage:inPage+toCopy])
		n += toCopy
		off += int64(toCopy)
		remaining -= toCopy
	}
	return n, nil
}

func (f *FileDevice) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 || off >= f.size {
		return 0, ErrOutOfBounds
	}
	if int64(len(p)) > f.size-off {
		p = p[:f.size-off]
	}
	n := 0
	for remaining := len(p); remaining > 0; {
		idx := uint64(off) / f.pageSize
		inPage := int(off % int64(f.pageSize))
		toCopy := int(f.pageSize) - inPage
		if toCopy > remaining {
			toCopy = remaining
		}

		f.mu.RLock()
		page := f.pages[idx]
		f.mu.RUnlock()

		if page == nil {
			page = make([]byte, int(f.pageSize))
			f.mu.Lock()
			if existing := f.pages[idx]; existing == nil {
				f.pages[idx] = page
			} else {
				page = existing
			}
			f.mu.Unlock()
		}

		copy(page[inPage:inPage+toCopy], p[n:n+toCopy])

		f.mu.Lock()
		f.dirty[idx] = true
		f.mu.Unlock()

		n += toCopy
		off += int64(toCopy)
		remaining -= toCopy
	}
	return n, nil
}

func (f *FileDevice) Flush() error {
	f.mu.RLock()
	dirtyCopy := make([]uint64, 0, len(f.dirty))
	for k := range f.dirty {
		dirtyCopy = append(dirtyCopy, k)
	}
	f.mu.RUnlock()

	for _, idx := range dirtyCopy {
		f.mu.RLock()
		page := f.pages[idx]
		f.mu.RUnlock()
		addr := store.PageAddress{Export: f.name, Index: idx, Size: f.pageSize}
		if err := f.st.WritePage(context.Background(), addr, page); err != nil {
			return err
		}
		f.mu.Lock()
		delete(f.dirty, idx)
		f.mu.Unlock()
	}
	return nil
}

func (f *FileDevice) Close() error { return nil }
