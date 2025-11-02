package nbd

import (
	"context"
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

	S3Bucket    string
	S3Region    string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
}

func Run(cfg Config) error {
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	var st store.Store
	if cfg.S3Bucket != "" {
		s3Store, err := store.NewS3Store(context.Background(), store.S3Config{
			Bucket:          cfg.S3Bucket,
			Region:          cfg.S3Region,
			Endpoint:        cfg.S3Endpoint,
			AccessKeyID:     cfg.S3AccessKey,
			SecretAccessKey: cfg.S3SecretKey,
		})
		if err != nil {
			return err
		}
		st = s3Store
		log.Printf("nbd: listening on %s (defaultSize=%d, chunkSize=%d, storage=s3, bucket=%s)",
			cfg.Addr, cfg.DefaultSize, cfg.ChunkSize, cfg.S3Bucket)
	} else {
		st = store.NewFSStore(filepath.Join(cfg.DataDir, "exports"))
		log.Printf("nbd: listening on %s (defaultSize=%d, chunkSize=%d, storage=filesystem)",
			cfg.Addr, cfg.DefaultSize, cfg.ChunkSize)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("nbd: accept error: %v", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			newDevice := func(name string, size uint64) core.Device {
				return core.NewMemDevice(name, int64(size), cfg.ChunkSize, st)
			}

			if err := ServeConn(c, cfg, newDevice); err != nil {
				log.Printf("nbd: connection %s error: %v", c.RemoteAddr(), err)
			}
		}(conn)
	}
}
