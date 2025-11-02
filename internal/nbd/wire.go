package nbd

import (
	"encoding/binary"
	"io"
)

var be = binary.BigEndian

func readN(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func readU16(r io.Reader) (uint16, error) {
	var v uint16
	if err := binary.Read(r, be, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func readU32(r io.Reader) (uint32, error) {
	var v uint32
	if err := binary.Read(r, be, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func readU64(r io.Reader) (uint64, error) {
	var v uint64
	if err := binary.Read(r, be, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func writeU16(w io.Writer, v uint16) error { return binary.Write(w, be, v) }
func writeU32(w io.Writer, v uint32) error { return binary.Write(w, be, v) }
func writeU64(w io.Writer, v uint64) error { return binary.Write(w, be, v) }
