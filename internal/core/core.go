package core

import (
	"errors"
	"sync"
)

type Device interface {
	ReadAt(p []byte, off int64) (int, error)
	WriteAt(p []byte, off int64) (int, error)
	Flush() error
	Size() int64
	Close() error
}

var (
	regMu sync.Mutex
	reg   = map[string]*ramDev{}
)

func OpenOrCreate(name string, size uint64) (Device, error) {
	regMu.Lock()
	defer regMu.Unlock()
	if d, ok := reg[name]; ok {
		return d, nil
	}
	d := newRam(size)
	reg[name] = d
	return d, nil
}

type ramDev struct {
	mu   sync.RWMutex
	data []byte
}

func newRam(size uint64) *ramDev {
	return &ramDev{data: make([]byte, size)}
}

func (d *ramDev) Size() int64  { return int64(len(d.data)) }
func (d *ramDev) Close() error { return nil }
func (d *ramDev) Flush() error { return nil }

func (d *ramDev) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || off+int64(len(p)) > int64(len(d.data)) {
		return 0, errors.New("out of bounds")
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	copy(p, d.data[off:int(off)+len(p)])
	return len(p), nil
}

func (d *ramDev) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 || off+int64(len(p)) > int64(len(d.data)) {
		return 0, errors.New("out of bounds")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	copy(d.data[off:int(off)+len(p)], p)
	return len(p), nil
}
