package crypto

import (
	"encoding/base64"
	"github.com/zeebo/blake3"
)

type Blake3Hasher struct {
	hasher *blake3.Hasher
}

func NewBlake3() *Blake3Hasher {
	return &Blake3Hasher{
		hasher: blake3.New(),
	}
}

func (b *Blake3Hasher) HashBytes(data []byte) []byte {
	b.hasher.Reset()
	b.hasher.Write(data)
	return b.hasher.Sum(nil)
}

func (b *Blake3Hasher) HashToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(b.HashBytes(data))
}

func (b *Blake3Hasher) HashString(data string) []byte {
	return b.HashBytes([]byte(data))
}
