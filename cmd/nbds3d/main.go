package main

import (
	"flag"
	"log"

	"nbds3d/internal/nbd"
)

func main() {
	addr := flag.String("addr", ":10809", "listen address")
	defaultSize := flag.Uint64("default-size", 1073741824, "default export size in bytes (1GiB)")
	chunkSize := flag.Uint64("chunk-size", 4194304, "chunk/page size in bytes (4MiB)")
	flag.Parse()

	cfg := nbd.Config{
		Addr:        *addr,
		DefaultSize: *defaultSize,
		ChunkSize:   *chunkSize,
	}

	if err := nbd.Run(cfg); err != nil {
		log.Fatalf("nbds3d: %v", err)
	}
}
