// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package ciphering_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/ciphering"
)

func ExampleAES128() {
	// Create a new AES-128 handler with a secret key
	aes128, err := ciphering.NewAES128("mysecret")
	if err != nil {
		fmt.Println(err)
	}

	// Data to be encrypted
	plaintext := []byte("This is a secret message")

	// Encrypt the data
	ciphertext, err := aes128.Encrypt(plaintext)
	if err != nil {
		fmt.Println(err)
	}

	// Decrypt the data
	decrypted, err := aes128.Decrypt(ciphertext)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Decrypted: %s\n", decrypted)

	// Output:
	// Decrypted: This is a secret message
}

func ExampleAES256() {
	// Create a new AES-256 handler with a secret key
	aes256, err := ciphering.NewAES256("anothersecret")
	if err != nil {
		fmt.Println(err)
	}

	// Data to be encrypted
	plaintext := []byte("Another secret message")

	// Encrypt the data
	ciphertext, err := aes256.Encrypt(plaintext)
	if err != nil {
		fmt.Println(err)
	}

	// Decrypt the data
	decrypted, err := aes256.Decrypt(ciphertext)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Decrypted: %s\n", decrypted)

	// Output:
	// Decrypted: Another secret message
}
