//go:build unit

package service

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBedrockEventStreamDecoderHeaderBeyondFrame(t *testing.T) {
	// Small valid-CRC frame: reproduce bounds failure without a large allocation.
	raw := make([]byte, 16)
	binary.BigEndian.PutUint32(raw[0:4], 16)
	binary.BigEndian.PutUint32(raw[4:8], 100)
	binary.BigEndian.PutUint32(raw[8:12], crc32.ChecksumIEEE(raw[:8]))
	binary.BigEndian.PutUint32(raw[12:16], crc32.ChecksumIEEE(raw[:12]))
	require.NotPanics(t, func() { _, err := newBedrockEventStreamDecoder(bytes.NewReader(raw)).Decode(); require.Error(t, err) })
}
