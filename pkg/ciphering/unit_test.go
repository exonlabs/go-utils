// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package ciphering_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exonlabs/go-utils/pkg/ciphering"
)

func TestAES128_EncryptDecrypt(t *testing.T) {
	secret := "mysecret"
	aes128, err := ciphering.NewAES128(secret)
	require.NoError(t, err)

	plaintext := []byte("Test data")

	// Test encryption
	ciphertext, err := aes128.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext, "Ciphertext should not match plaintext")

	// Test decryption
	decrypted, err := aes128.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "Decrypted data should match original plaintext")
}

func TestAES128_EmptyInput(t *testing.T) {
	secret := "mysecret"
	aes128, err := ciphering.NewAES128(secret)
	require.NoError(t, err)

	// Test encrypting empty input
	_, err = aes128.Encrypt(nil)
	assert.EqualError(t, err, "input data cannot be empty")

	// Test decrypting empty input
	_, err = aes128.Decrypt(nil)
	assert.EqualError(t, err, "input data is too short")
}

func TestAES256_EncryptDecrypt(t *testing.T) {
	secret := "myverysecuresecret"
	aes256, err := ciphering.NewAES256(secret)
	require.NoError(t, err)

	plaintext := []byte("Another test data")

	// Test encryption
	ciphertext, err := aes256.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext, "Ciphertext should not match plaintext")

	// Test decryption
	decrypted, err := aes256.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "Decrypted data should match original plaintext")
}

func TestAES256_EmptyInput(t *testing.T) {
	secret := "myverysecuresecret"
	aes256, err := ciphering.NewAES256(secret)
	require.NoError(t, err)

	// Test encrypting empty input
	_, err = aes256.Encrypt(nil)
	assert.EqualError(t, err, "input data cannot be empty")

	// Test decrypting empty input
	_, err = aes256.Decrypt(nil)
	assert.EqualError(t, err, "input data is too short")
}
