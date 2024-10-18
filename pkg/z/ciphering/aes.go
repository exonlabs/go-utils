package xcipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"io"
)

type Cipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type aesCipher struct {
	aead cipher.AEAD
	aad  []byte
}

// encryption using AES-GCM
func (ciph *aesCipher) Encrypt(in []byte) ([]byte, error) {
	if len(in) == 0 {
		return nil, errors.New("malformed data for encryption")
	}
	nonce := make([]byte, ciph.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	out := ciph.aead.Seal(nil, nonce, in, ciph.aad)
	return append(nonce, out...), nil
}

// decryption using AES-GCM
func (ciph *aesCipher) Decrypt(in []byte) ([]byte, error) {
	n := ciph.aead.NonceSize()
	if len(in) <= n {
		return nil, errors.New("malformed data for decryption")
	}
	out, err := ciph.aead.Open(nil, in[:n], in[n:], ciph.aad)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type AES128 struct {
	*aesCipher
}

func NewAES128(secret string) (*AES128, error) {
	h := sha256.Sum256([]byte(secret))
	l := len(h) / 2
	for i := 0; i < l; i += 2 { // mangle first 16 bytes
		h[i], h[i+1] = h[i+1], h[i]
	}
	block, err := aes.NewCipher(h[l:]) // AES-128, 128-bit key
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AES128{
		aesCipher: &aesCipher{
			aead: aead,
			aad:  h[:l], // 128-bit key
		},
	}, nil
}

type AES256 struct {
	*aesCipher
}

func NewAES256(secret string) (*AES256, error) {
	h := sha512.Sum512([]byte(secret))
	l := len(h) / 2
	for i := 0; i < l; i += 2 { // mangle first 32 bytes
		h[i], h[i+1] = h[i+1], h[i]
	}
	block, err := aes.NewCipher(h[l:]) // AES-256, 256-bit key
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AES256{
		aesCipher: &aesCipher{
			aead: aead,
			aad:  h[:l], // 256-bit key
		},
	}, nil
}
