package nbd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"nbds3d/internal/core"
	"net"
)

const (
	NBD_INFO_EXPORT = 0
)

func ServeConn(c net.Conn, cfg Config, newDev func(name string, size uint64) core.Device) error {
	br := bufio.NewReader(c)
	bw := bufio.NewWriter(c)
	defer c.Close()
	defer bw.Flush()

	if err := writeU64(bw, NBDMAGIC); err != nil {
		return err
	}
	if err := writeU64(bw, IHAVEOPT); err != nil {
		return err
	}
	if err := writeU16(bw, uint16(NBD_FLAG_FIXED_NEWSTYLE|NBD_FLAG_NO_ZEROES)); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	if _, err := readU32(br); err != nil {
		return fmt.Errorf("client flags: %w", err)
	}

	var exportName string
	exportSize := cfg.DefaultSize

	for {
		magic, err := readU64(br)
		if err != nil {
			return err
		}
		if magic != IHAVEOPT {
			return fmt.Errorf("bad option magic: 0x%x", magic)
		}

		opt, err := readU32(br)
		if err != nil {
			return err
		}
		olen, err := readU32(br)
		if err != nil {
			return err
		}
		data, err := readN(br, int(olen))
		if err != nil {
			return err
		}

		switch opt {
		case NBD_OPT_ABORT:
			if err := writeReply(bw, opt, NBD_REP_ACK, nil); err != nil {
				return err
			}
			_ = bw.Flush()
			return nil

		case NBD_OPT_GO:
			if len(data) < 6 {
				_ = writeReply(bw, opt, NBD_REP_ERR_INVALID, []byte("short GO"))
				continue
			}
			rd := bytes.NewReader(data)
			nameLen, _ := readU32(rd)
			if int(nameLen)+6 > len(data) {
				_ = writeReply(bw, opt, NBD_REP_ERR_INVALID, []byte("bad nameLen"))
				continue
			}
			name := make([]byte, nameLen)
			if _, err := io.ReadFull(rd, name); err != nil {
				_ = writeReply(bw, opt, NBD_REP_ERR_INVALID, []byte("name read"))
				continue
			}
			exportName = string(name)

			infoCount, _ := readU16(rd)
			if _, err := io.CopyN(io.Discard, rd, int64(infoCount*2)); err != nil && err != io.EOF {
				_ = writeReply(bw, opt, NBD_REP_ERR_INVALID, []byte("info read"))
				continue
			}

			txFlags := uint16(NBD_FLAG_HAS_FLAGS | NBD_FLAG_SEND_FLUSH)
			if err := writeReply(bw, opt, NBD_REP_INFO, infoExportPayload(exportSize, txFlags)); err != nil {
				return err
			}
			if err := writeReply(bw, opt, NBD_REP_ACK, nil); err != nil {
				return err
			}
			if err := bw.Flush(); err != nil {
				return err
			}

			dev := newDev(exportName, exportSize)
			defer dev.Close()
			return transmit(br, bw, dev)

		default:
			if err := writeReply(bw, opt, NBD_REP_ERR_UNSUP, nil); err != nil {
				return err
			}
			if err := bw.Flush(); err != nil {
				return err
			}
		}
	}
}

func infoExportPayload(size uint64, txFlags uint16) []byte {
	var b bytes.Buffer
	_ = writeU16(&b, NBD_INFO_EXPORT)
	_ = writeU64(&b, size)
	_ = writeU16(&b, txFlags)
	return b.Bytes()
}

func writeReply(w *bufio.Writer, opt uint32, repType uint32, payload []byte) error {
	if err := writeU64(w, NBD_REP_MAGIC); err != nil {
		return err
	}
	if err := writeU32(w, opt); err != nil {
		return err
	}
	if err := writeU32(w, repType); err != nil {
		return err
	}
	if err := writeU32(w, uint32(len(payload))); err != nil {
		return err
	}
	if len(payload) > 0 {
		if _, err := w.Write(payload); err != nil {
			return err
		}
	}
	return nil
}
