// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package jconfig_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/jconfig"
)

func ExampleJConfig_SetSecure() {
	cfg, _ := jconfig.New("config.json", dictx.Dict{})
	cfg.InitAES128("thisis128bitkey!!")

	val := map[string]any{"username": "admin", "password": "secret"}
	cfg.SetSecure("credentials", val)

	// Retrieve the secure value
	retrieved, _ := cfg.GetSecure("credentials", nil)
	fmt.Println(retrieved)

	// Output: map[password:secret username:admin]
}

func ExampleJConfig_GetSecure() {
	cfg, _ := jconfig.New("config.json", dictx.Dict{})
	cfg.InitAES128("thisis128bitkey!!")

	// Retrieve a non-existing key
	retrieved, _ := cfg.GetSecure("non_existing_key", "default")
	fmt.Println(retrieved)

	// Output: default
}
