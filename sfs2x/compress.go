package sfs2x

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// defaultCompressThreshold is the payload-size threshold used by the
// official client: payloads larger than this are zlib-compressed before
// framing (header bit 0x20).
const defaultCompressThreshold = 1024

// compressPayload zlib-compresses a frame payload (header bit 0x20).
func compressPayload(payload []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(payload); err != nil {
		return nil, fmt.Errorf("zlib compress: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("zlib compress: %w", err)
	}
	return buf.Bytes(), nil
}

// decompressPayload inflates a zlib-compressed frame payload. Output is
// capped at maxDecompressedSize to resist decompression bombs.
func decompressPayload(payload []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("zlib decompress: %w", err)
	}
	defer zr.Close()

	var out bytes.Buffer
	// Limit output to a sane multiple of the compressed size plus a floor —
	// a 1024-byte bomb must not inflate into gigabytes of memory.
	limit := int64(maxDecompressedSize)
	n, err := io.Copy(&out, io.LimitReader(zr, limit+1))
	if err != nil {
		return nil, fmt.Errorf("zlib decompress: %w", err)
	}
	if n > limit {
		return nil, fmt.Errorf("zlib decompress: output exceeds %d bytes", maxDecompressedSize)
	}
	return out.Bytes(), nil
}

// maxDecompressedSize caps how much a compressed frame may expand to.
const maxDecompressedSize = 64 << 20 // 64 MiB
