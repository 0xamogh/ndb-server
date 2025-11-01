package nbd

import (
	"log"
	"nbds3d/internal/core"
	"nbds3d/internal/store"
	"net"
	"path/filepath"
)

type Config struct {
	Addr        string
	DefaultSize uint64
	ChunkSize   uint64
	DataDir     string
}

func Run(cfg Config) error {
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("nbd: listening on %s (defaultSize=%d, chunkSize=%d)", cfg.Addr, cfg.DefaultSize, cfg.ChunkSize)

	st := store.NewFSStore(filepath.Join(cfg.DataDir, "exports"))

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("nbd: accept error: %v", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			newDevice := func(name string, size uint64) core.Device {
				return core.NewFileDevice(name, int64(size), cfg.ChunkSize, st)
			}

			if err := ServeConn(c, cfg, newDevice); err != nil {
				log.Printf("nbd: connection %s error: %v", c.RemoteAddr(), err)
			}
		}(conn)
	}
}
