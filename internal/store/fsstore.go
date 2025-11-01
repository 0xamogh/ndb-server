package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func NewFSStore(root string) *FSStore {
	return &FSStore{rootDir: root}
}

func (s *FSStore) pagePath(export string, index uint64) string {
	return filepath.Join(s.rootDir, export, fmt.Sprintf("page-%08d.bin", index))
}

func (s *FSStore) ReadPage(ctx context.Context, addr PageAddress) ([]byte, error) {
	path := s.pagePath(addr.Export, addr.Index)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]byte, addr.Size), nil
		}
		return nil, err
	}
	defer file.Close()

	buf := make([]byte, addr.Size)
	_, err = io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return buf, nil
}

func (s *FSStore) WritePage(ctx context.Context, addr PageAddress, data []byte) error {
	dir := filepath.Join(s.rootDir, addr.Export)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := s.pagePath(addr.Export, addr.Index)
	tmp := path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(tmp)
		return err
	}
	file.Close()
	return os.Rename(tmp, path)
}

func (s *FSStore) FlushExport(ctx context.Context, export string) error {
	// For filesystem backend, rename is atomic — nothing extra needed.
	return nil
}
