# Solution: NBD S3 Storage Server

## Overview

This is an implementation of a Network Block Device (NBD) server that durably persists data to S3-compatible storage. The server supports lazy loading, page-based caching, and efficient write batching to minimize S3 operations.

## Core Requirements

| Requirement | Implementation | Code Location |
|------------|----------------|---------------|
| NBD protocol support (read/write/flush) | Fully implemented | `internal/nbd/transmit.go` |
| Fixed newstyle handshake | NBD_OPT_GO and NBD_OPT_ABORT | `internal/nbd/handshake.go` |
| Arbitrary export names | Each export isolated in S3 | `internal/store/s3store.go:64` |
| Durable S3 persistence | All data stored in S3 | `internal/store/s3store.go` |
| Proper flush semantics | Write-back cache with flush | `internal/core/filedev.go:122-152` |

## Architecture

### Component Overview

```
┌─────────────────────────────────────────┐
│         NBD Client (nbd-client)         │
│              /dev/nbdX                  │
└─────────────────┬───────────────────────┘
                  │ TCP (NBD Protocol)
┌─────────────────▼───────────────────────┐
│           NBD Server (nbds3d)           │
│                                         │
│  ┌────────────────────────────────┐    │
│  │  Handshake (handshake.go)      │    │
│  │  - Fixed newstyle negotiation  │    │
│  │  - Export name handling        │    │
│  └────────────────────────────────┘    │
│                                         │
│  ┌────────────────────────────────┐    │
│  │  Transmit (transmit.go)        │    │
│  │  - Read/Write/Flush commands   │    │
│  └────────────┬───────────────────┘    │
│               │                         │
│  ┌────────────▼───────────────────┐    │
│  │  FileDevice (filedev.go)       │    │
│  │  - Page cache (in-memory)      │    │
│  │  - Lazy loading                │    │
│  │  - Dirty page tracking         │    │
│  └────────────┬───────────────────┘    │
│               │                         │
│  ┌────────────▼───────────────────┐    │
│  │  Store Interface               │    │
│  │  - ReadPage / WritePage        │    │
│  └────────┬───────────┬───────────┘    │
│           │           │                 │
│    ┌──────▼─────┐ ┌──▼──────────┐     │
│    │  FSStore   │ │  S3Store    │     │
│    │ (local fs) │ │ (AWS SDK)   │     │
│    └────────────┘ └──────┬──────┘     │
└───────────────────────────┼────────────┘
                            │
                   ┌────────▼────────┐
                   │  S3 / MinIO     │
                   │  (Object Store) │
                   └─────────────────┘
```

### Key Design Decisions

**1. Page-Based Storage**
- Device divided into fixed-size pages (default 4MB)
- Each page stored as separate S3 object: `exports/<export>/page-XXXXXXXX.bin`
- Enables lazy loading and efficient caching

**2. Write-Back Cache**
- Writes update in-memory pages and mark them dirty
- Flush operation writes all dirty pages to S3
- Reduces S3 API calls and costs

**3. Zero-Fill on Missing Pages**
- Non-existent pages return zeros (sparse storage)
- Only written pages consume S3 storage
- S3Store: `internal/store/s3store.go:77-79`

**4. Store Interface Abstraction**
- Clean separation between NBD logic and storage backend
- Easy to test with FSStore
- Production uses S3Store
- Interface defined in `internal/store/types.go`


### Additional Features

#### 1. Tiered Caching (Lazy Load + Write-Back)
- In-memory page cache within `FileDevice` supports **lazy loading** from S3 on demand, minimizing unnecessary reads.  
- Implements **write-back caching** that batches in-memory writes before flushing to S3, reducing PUT operations.  
- **Implementation:** `internal/core/filedev.go:34–120`

#### 2. Pluggable Storage Backends
- Unified `Store` interface enabling multiple persistence layers.  
- Two concrete implementations: **`FSStore`** (local filesystem) and **`S3Store`** (S3-compatible storage).  
- Simplifies testing and allows easy backend swaps.  
- **Interface:** `internal/store/types.go:15–19`

#### 3. Concurrent Connection Support
- Each NBD connection runs in its own **goroutine** with an independent `FileDevice` instance.  
- **Thread-safe** operation ensured via mutex protection.  
- **Implementation:** `internal/nbd/server.go:53–68`

#### 4. Operational Logging
- Logs **flush** and **write** operations for visibility and debugging.  
- Contextual **error logging** improves traceability during S3 persistence or cache sync events.

## Code Structure

```
cmd/nbds3d/
  main.go              # Entry point, CLI flags, config

internal/
  nbd/
    server.go          # TCP server, connection handling
    handshake.go       # NBD protocol negotiation
    transmit.go        # Read/write/flush command handling
    protocol.go        # NBD protocol constants

  core/
    core.go            # Device interface
    filedev.go         # Page cache implementation
    memdev.go          # In-memory device (legacy)

  store/
    types.go           # Store interface definition
    fsstore.go         # Filesystem storage backend
    s3store.go         # S3 storage backend
```

## Configuration

The server supports flexible configuration via command-line flags:

**Storage Selection:**
- If `--s3-bucket` is set → uses S3Store
- Otherwise → uses FSStore (filesystem)

**S3 Configuration:**
- `--s3-bucket`: S3 bucket name
- `--s3-region`: AWS region (default: us-east-1)
- `--s3-endpoint`: Custom endpoint for MinIO/other S3-compatible services
- `--s3-access-key`: Access key ID
- `--s3-secret-key`: Secret access key

**Performance Tuning:**
- `--chunk-size`: Page size in bytes (default: 4MB)
- Larger pages = fewer S3 objects, but more memory per page
- Smaller pages = finer granularity, more S3 API calls

## Testing

Successfully tested with:
- MinIO (S3-compatible local storage)
- ext4 filesystem operations
- Data persistence across server restarts
- Multiple simultaneous connections
- Lazy loading verification

## Performance Characteristics

**S3 API Calls:**
- Read: 1 GetObject per page (cached after first read)
- Write: 1 PutObject per dirty page per flush
- Lazy loading minimizes unnecessary reads

**Memory Usage:**
- Proportional to number of accessed pages
- Each page: `chunk-size` bytes (default 4MB)
- Dirty pages retained until flush
