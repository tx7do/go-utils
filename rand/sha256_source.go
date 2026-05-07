package rand

import (
	"crypto/sha256"
	"encoding/binary"
)

type Sha256Source struct {
	seed  []byte
	nonce uint64
}

func NewSha256Source(seed []byte) *Sha256Source {
	return &Sha256Source{seed: seed, nonce: 0}
}

// Uint64 生成一个基于 SHA-256 哈希的随机 uint64。每次调用都会增加 nonce，确保生成不同的随机数。
func (s *Sha256Source) Uint64() uint64 {
	s.nonce++
	h := sha256.New()
	h.Write(s.seed)

	nonceBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBuf, s.nonce)
	h.Write(nonceBuf)

	hashResult := h.Sum(nil)
	return binary.LittleEndian.Uint64(hashResult[:8])
}
