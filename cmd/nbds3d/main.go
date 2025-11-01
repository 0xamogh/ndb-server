package main

import (
	"flag"
	"log"
	"os"

	"nbds3d/internal/nbd"
)

func main() {
	addr := flag.String("addr", ":10809", "listen address (host:port)")
	defaultSize := flag.Uint64("default-size", 1073741824, "default export size in bytes (e.g. 1073741824 = 1GiB)")
	chunkSize := flag.Uint64("chunk-size", 4194304, "page/chunk size in bytes (e.g. 4194304 = 4MiB)")
	dataDir := flag.String("data-dir", "./data", "directory to store exports/pages")

	flag.Parse()

	cfg := nbd.Config{
		Addr:        *addr,
		DefaultSize: *defaultSize,
		ChunkSize:   *chunkSize,
		DataDir:     *dataDir,
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", cfg.DataDir, err)
	}

	if err := nbd.Run(cfg); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
