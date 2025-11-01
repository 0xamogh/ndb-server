package store

import "context"

type PageAddress struct {
	Export string // export name
	Index  uint64 // page index (offset/pageSize)
	Size   uint64 // page size in bytes
}

type FSStore struct {
	rootDir string
}

type Store interface {
	ReadPage(ctx context.Context, addr PageAddress) ([]byte, error)
	WritePage(ctx context.Context, addr PageAddress, data []byte) error
	FlushExport(ctx context.Context, export string) error
}
