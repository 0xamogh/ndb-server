# NBD S3 Storage Server

An NBD (Network Block Device) server with durable S3-compatible storage backend.

## Features

- NBD protocol server for network block device access
- Pluggable storage backends:
  - Local filesystem (FSStore)
  - S3-compatible storage (S3Store) - AWS S3, MinIO, etc.
- Page-based lazy loading and caching
- Durable persistence to S3

## Building

```bash
go build -o nbds3d ./cmd/nbds3d
```

## Usage

### Filesystem Storage (Default)

```bash
./nbds3d --data-dir=./data
```

### S3 Storage

```bash
./nbds3d \
  --s3-bucket=my-bucket \
  --s3-region=us-east-1 \
  --s3-access-key=YOUR_ACCESS_KEY \
  --s3-secret-key=YOUR_SECRET_KEY
```

### MinIO (S3-Compatible)

```bash
./nbds3d \
  --s3-bucket=my-bucket \
  --s3-region=us-east-1 \
  --s3-endpoint=http://localhost:9000 \
  --s3-access-key=minioadmin \
  --s3-secret-key=minioadmin \
  --chunk-size=4096
```

## Configuration Flags

- `--addr`: Listen address (default: `:10809`)
- `--default-size`: Default export size in bytes (default: `1073741824` = 1GiB)
- `--chunk-size`: Page/chunk size in bytes (default: `4194304` = 4MiB)
- `--data-dir`: Directory for filesystem storage (default: `./data`)
- `--s3-bucket`: S3 bucket name (enables S3 storage when set)
- `--s3-region`: S3 region (default: `us-east-1`)
- `--s3-endpoint`: S3 endpoint URL for MinIO or other S3-compatible services
- `--s3-access-key`: S3 access key ID
- `--s3-secret-key`: S3 secret access key

## Testing with MinIO

### 1. Start MinIO

```bash
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio-test \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  quay.io/minio/minio server /data --console-address ":9001"
```

### 2. Create Bucket

```bash
docker exec minio-test mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec minio-test mc mb local/nbd-test
```

### 3. Start NBD Server with S3 Storage

```bash
./nbds3d \
  --s3-bucket=nbd-test \
  --s3-region=us-east-1 \
  --s3-endpoint=http://localhost:9000 \
  --s3-access-key=minioadmin \
  --s3-secret-key=minioadmin \
  --chunk-size=4096
```

### 4. Connect NBD Client (on Ubuntu/Linux)

```bash
sudo modprobe nbd
sudo nbd-client <SERVER_IP> 10809 /dev/nbd1 -name test
```

### 5. Create Filesystem and Mount

```bash
sudo mkfs.ext4 /dev/nbd1
sudo mkdir -p /mnt/nbdtest
sudo mount /dev/nbd1 /mnt/nbdtest
```

### 6. Write Test Data

```bash
echo "hello from s3" | sudo tee /mnt/nbdtest/test.txt
cat /mnt/nbdtest/test.txt
sudo sync
```

### 7. Verify Data in MinIO

```bash
docker exec minio-test mc ls --recursive local/nbd-test/
```

You should see page files under `exports/test/`.

You can also browse the MinIO web console at http://localhost:9001:
- Username: `minioadmin`
- Password: `minioadmin`

### 8. Test Persistence

Disconnect and stop the server:

```bash
sudo umount /mnt/nbdtest
sudo nbd-client -d /dev/nbd1
```

Stop the NBD server (Ctrl+C), then restart it with the same S3 configuration.

Reconnect and verify your data persisted:

```bash
sudo nbd-client <SERVER_IP> 10809 /dev/nbd1 -name test
sudo mount /dev/nbd1 /mnt/nbdtest
cat /mnt/nbdtest/test.txt
ls -la /mnt/nbdtest/
```

Your files should still be there, loaded from S3.

### 9. Cleanup

```bash
sudo umount /mnt/nbdtest
sudo nbd-client -d /dev/nbd1
docker stop minio-test
docker rm minio-test
```

## Architecture

### Storage Interface

```go
type Store interface {
    ReadPage(ctx context.Context, addr PageAddress) ([]byte, error)
    WritePage(ctx context.Context, addr PageAddress, data []byte) error
    FlushExport(ctx context.Context, export string) error
}
```

### Implementations

- **FSStore**: Stores pages as files on local disk (`./data/exports/<export>/page-XXXXXXXX.bin`)
- **S3Store**: Stores pages as S3 objects (`exports/<export>/page-XXXXXXXX.bin`)

### Page Cache

The `FileDevice` implements a write-back cache:
- Pages are loaded lazily on first read
- Writes update the in-memory cache and mark pages dirty
- Flush commands write dirty pages to the storage backend
- Non-existent pages return zeros

## License

MIT
