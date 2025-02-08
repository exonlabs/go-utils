// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package jconfig_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/jconfig"
)

// TestNewConfig tests creating a new Config instance with valid parameters
func TestNewConfig(t *testing.T) {
	defaults := dictx.Dict{"foo": "bar"}
	_, err := jconfig.New("config.json", defaults)
	require.NoError(t, err)
}

// TestInitAES128 tests the initialization of AES-128 encryption
func TestInitAES128(t *testing.T) {
	defaults := dictx.Dict{}
	cfg, err := jconfig.New("config.json", defaults)
	require.NoError(t, err)

	err = cfg.InitAES128("thisis128bitkey!!")
	require.NoError(t, err)
}

// TestInitAES256 tests the initialization of AES-256 encryption
func TestInitAES256(t *testing.T) {
	defaults := dictx.Dict{}
	cfg, err := jconfig.New("config.json", defaults)
	require.NoError(t, err)

	err = cfg.InitAES256("thisisaverylongkeythatisfortestingaes256")
	require.NoError(t, err)
}

// TestSetGetSecure tests encryption and decryption of secure values
func TestSetGetSecure(t *testing.T) {
	cfg, err := jconfig.New("config.json", dictx.Dict{})
	require.NoError(t, err)

	err = cfg.InitAES128("thisis128bitkey!!")
	require.NoError(t, err)

	val := dictx.Dict{"username": "admin", "password": "secret"}
	err = cfg.SetSecure("credentials", val)
	require.NoError(t, err)

	retrieved, err := cfg.GetSecure("credentials", nil)
	require.NoError(t, err)
	assert.Equal(t, val, retrieved)
}

// TestGetSecureDefaultValue tests GetSecure with a non-existing key
func TestGetSecureDefaultValue(t *testing.T) {
	cfg, err := jconfig.New("config.json", dictx.Dict{})
	require.NoError(t, err)

	err = cfg.InitAES128("thisis128bitkey!!")
	require.NoError(t, err)

	// Non-existing key should return default value
	defaultValue := "default"
	retrieved, err := cfg.GetSecure("non_existing_key", defaultValue)
	require.NoError(t, err)
	assert.Equal(t, defaultValue, retrieved)
}

// TestInvalidSecureValueFormat tests handling of invalid value formats in GetSecure
func TestInvalidSecureValueFormat(t *testing.T) {
	cfg, err := jconfig.New("config.json", dictx.Dict{})
	require.NoError(t, err)

	// Set an invalid non-string value
	cfg.Set("invalid_key", 12345)

	_, err = cfg.GetSecure("invalid_key", nil)
	assert.Error(t, err)
}
