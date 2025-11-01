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

	s3Bucket := flag.String("s3-bucket", "", "S3 bucket name (enables S3 storage when set)")
	s3Region := flag.String("s3-region", "us-east-1", "S3 region")
	s3Endpoint := flag.String("s3-endpoint", "", "S3 endpoint URL (for MinIO or other S3-compatible services)")
	s3AccessKey := flag.String("s3-access-key", "", "S3 access key ID")
	s3SecretKey := flag.String("s3-secret-key", "", "S3 secret access key")

	flag.Parse()

	cfg := nbd.Config{
		Addr:        *addr,
		DefaultSize: *defaultSize,
		ChunkSize:   *chunkSize,
		DataDir:     *dataDir,
		S3Bucket:    *s3Bucket,
		S3Region:    *s3Region,
		S3Endpoint:  *s3Endpoint,
		S3AccessKey: *s3AccessKey,
		S3SecretKey: *s3SecretKey,
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", cfg.DataDir, err)
	}

	if err := nbd.Run(cfg); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
