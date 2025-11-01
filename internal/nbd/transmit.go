package nbd

import (
	"bufio"
	"errors"
	"io"

	"nbds3d/internal/core"
)

func writeSimpleReply(w *bufio.Writer, errCode uint32, cookie uint64, payload []byte) error {
	if err := writeU32(w, NBD_SIMPLE_REPLY_MAGIC); err != nil {
		return err
	}
	if err := writeU32(w, errCode); err != nil {
		return err
	}
	if err := writeU64(w, cookie); err != nil {
		return err
	}
	if errCode == 0 && len(payload) > 0 {
		if _, err := w.Write(payload); err != nil {
			return err
		}
	}
	return w.Flush()
}

func transmit(br *bufio.Reader, bw *bufio.Writer, exportName string, exportSize uint64, cfg Config) error {
	dev, err := core.OpenOrCreate(exportName, exportSize)
	if err != nil {
		return err
	}
	defer dev.Close()

	for {
		magic, err := readU32(br)
		if err != nil {
			return err
		}
		if magic != NBD_REQUEST_MAGIC {
			return errors.New("bad request magic")
		}

		cmdFlags, err := readU16(br)
		if err != nil {
			return err
		}
		_ = cmdFlags
		typ, err := readU16(br)
		if err != nil {
			return err
		}
		cookie, err := readU64(br)
		if err != nil {
			return err
		}
		off, err := readU64(br)
		if err != nil {
			return err
		}
		lengthU32, err := readU32(br)
		if err != nil {
			return err
		}
		length := int(lengthU32)

		switch typ {
		case NBD_CMD_READ:
			if int64(off)+int64(length) > dev.Size() {
				if err := writeSimpleReply(bw, NBD_EINVAL, cookie, nil); err != nil {
					return err
				}
				continue
			}
			buf := make([]byte, length)
			if _, err := io.ReadFull(bytesReaderZero{}, buf[:0]); err != nil { /* no-op to avoid import */
			}
			if _, err := dev.ReadAt(buf, int64(off)); err != nil {
				if err := writeSimpleReply(bw, NBD_EIO, cookie, nil); err != nil {
					return err
				}
				continue
			}
			if err := writeSimpleReply(bw, 0, cookie, buf); err != nil {
				return err
			}

		case NBD_CMD_WRITE:
			buf := make([]byte, length)
			if _, err := io.ReadFull(br, buf); err != nil {
				return err
			}
			if int64(off)+int64(length) > dev.Size() {
				// discard already-read data; respond error
				if err := writeSimpleReply(bw, NBD_ENOSPC, cookie, nil); err != nil {
					return err
				}
				continue
			}
			if _, err := dev.WriteAt(buf, int64(off)); err != nil {
				if err := writeSimpleReply(bw, NBD_EIO, cookie, nil); err != nil {
					return err
				}
				continue
			}
			if err := writeSimpleReply(bw, 0, cookie, nil); err != nil {
				return err
			}

		case NBD_CMD_FLUSH:
			if err := dev.Flush(); err != nil {
				if err := writeSimpleReply(bw, NBD_EIO, cookie, nil); err != nil {
					return err
				}
				continue
			}
			if err := writeSimpleReply(bw, 0, cookie, nil); err != nil {
				return err
			}

		case NBD_CMD_DISC:
			return nil

		default:
			if err := writeSimpleReply(bw, NBD_EINVAL, cookie, nil); err != nil {
				return err
			}
		}
	}
}

// tiny hack to avoid unused import for io during build; no actual read.
type bytesReaderZero struct{}

func (bytesReaderZero) Read(p []byte) (int, error) { return 0, io.EOF }
