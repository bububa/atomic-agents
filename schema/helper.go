package schema

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// --- Helper functions for binary marshaling/unmarshaling slices ---

// writeStringSlice writes a slice of strings to the buffer.
// Format: int32 (count) + for each string: (int32 (length) + bytes)
func writeStringSlice(buf *bytes.Buffer, sl []string) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(sl))); err != nil {
		return fmt.Errorf("failed to write slice length: %w", err)
	}
	for _, s := range sl {
		sBytes := []byte(s)
		if err := binary.Write(buf, binary.LittleEndian, int32(len(sBytes))); err != nil {
			return fmt.Errorf("failed to write string length: %w", err)
		}
		if _, err := buf.Write(sBytes); err != nil {
			return fmt.Errorf("failed to write string bytes: %w", err)
		}
	}
	return nil
}

// readStringSlice reads a slice of strings from the reader.
func readStringSlice(reader *bytes.Reader) ([]string, error) {
	var count int32
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("failed to read slice length: %w", err)
	}

	sl := make([]string, count)
	for i := range int(count) {
		var strLen int32
		if err := binary.Read(reader, binary.LittleEndian, &strLen); err != nil {
			return nil, fmt.Errorf("failed to read string length for element %d: %w", i, err)
		}
		strBytes := make([]byte, strLen)
		if _, err := io.ReadFull(reader, strBytes); err != nil {
			return nil, fmt.Errorf("failed to read string bytes for element %d: %w", i, err)
		}
		sl[i] = string(strBytes)
	}
	return sl, nil
}

// writeBytesSliceSlice writes a slice of byte slices (representing file contents) to the buffer.
// Format: int32 (outer slice count) + for each inner []byte: (int32 (length) + bytes)
func writeBytesSliceSlice(buf *bytes.Buffer, sl [][]byte) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(sl))); err != nil {
		return fmt.Errorf("failed to write outer slice length: %w", err)
	}
	for _, bSlice := range sl {
		if err := binary.Write(buf, binary.LittleEndian, int32(len(bSlice))); err != nil {
			return fmt.Errorf("failed to write inner slice length: %w", err)
		}
		if _, err := buf.Write(bSlice); err != nil {
			return fmt.Errorf("failed to write inner slice bytes: %w", err)
		}
	}
	return nil
}

// readBytesSliceSlice reads a slice of byte slices from the reader.
func readBytesSliceSlice(reader *bytes.Reader) ([][]byte, error) {
	var count int32
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("failed to read outer slice length: %w", err)
	}

	sl := make([][]byte, count)
	for i := range int(count) {
		var byteLen int32
		if err := binary.Read(reader, binary.LittleEndian, &byteLen); err != nil {
			return nil, fmt.Errorf("failed to read inner slice length for element %d: %w", i, err)
		}
		bSlice := make([]byte, byteLen)
		if _, err := io.ReadFull(reader, bSlice); err != nil {
			return nil, fmt.Errorf("failed to read inner slice bytes for element %d: %w", i, err)
		}
		sl[i] = bSlice
	}
	return sl, nil
}
