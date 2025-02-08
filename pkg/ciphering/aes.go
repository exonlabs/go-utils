// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package ciphering

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"io"
)

// Handler defines the contract for encryption and decryption methods.
type Handler interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

// aesHandler is a struct for AES encryption/decryption.
// It uses an AEAD cipher for authenticated encryption.
type aesHandler struct {
	aead cipher.AEAD // Authenticated encryption with associated data
	aad  []byte      // Additional data used for encryption
}

// Encrypt encrypts the input data using AES-GCM.
// It generates a nonce, encrypts the data, and prepends the nonce to the output.
func (h *aesHandler) Encrypt(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, errors.New("input data cannot be empty")
	}
	nonce := make([]byte, h.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := h.aead.Seal(nil, nonce, b, h.aad)
	return append(nonce, ciphertext...), nil
}

// Decrypt decrypts the input data using AES-GCM.
// It extracts the nonce from the input and decrypts the data.
func (h *aesHandler) Decrypt(b []byte) ([]byte, error) {
	nonceSize := h.aead.NonceSize()
	if len(b) <= nonceSize {
		return nil, errors.New("input data is too short")
	}
	nonce, ciphertext := b[:nonceSize], b[nonceSize:]
	plaintext, err := h.aead.Open(nil, nonce, ciphertext, h.aad)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// AES128 provides AES encryption with a 128-bit key.
type AES128 struct {
	*aesHandler
}

// NewAES128 creates a new AES-128 handler using the provided secret.
// The secret is hashed with SHA-256 to derive a 128-bit key and AAD.
func NewAES128(secret string) (*AES128, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[len(key)/2:]) // AES-128, 128-bit key
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AES128{
		aesHandler: &aesHandler{
			aead: aead,
			aad:  key[:len(key)/2], // 128-bit key
		},
	}, nil
}

// AES256 provides AES encryption with a 256-bit key.
type AES256 struct {
	*aesHandler
}

// NewAES256 creates a new AES-256 handler using the provided secret.
// The secret is hashed with SHA-512 to derive a 256-bit key and AAD.
func NewAES256(secret string) (*AES256, error) {
	key := sha512.Sum512([]byte(secret))
	block, err := aes.NewCipher(key[len(key)/2:]) // AES-256, 256-bit key
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AES256{
		aesHandler: &aesHandler{
			aead: aead,
			aad:  key[:len(key)/2], // 256-bit key
		},
	}, nil
}
