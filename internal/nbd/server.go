package nbd

import (
	"log"
	"net"
)

type Config struct {
	Addr        string
	DefaultSize uint64
	ChunkSize   uint64
}

func Run(cfg Config) error {
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("nbd: listening on %s (defaultSize=%d, chunkSize=%d)", cfg.Addr, cfg.DefaultSize, cfg.ChunkSize)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("nbd: temporary accept error: %v", err)
				continue
			}
			return err
		}

		go func(c net.Conn) {
			defer c.Close()
			if err := ServeConn(c, cfg); err != nil {
				log.Printf("nbd: connection %s error: %v", c.RemoteAddr(), err)
			}
		}(conn)
	}
}
