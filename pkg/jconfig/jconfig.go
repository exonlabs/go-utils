// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package jconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/ciphering"
)

// JConfig represents a configuration manager that handles loading,
// saving, and backing up configuration data.
type JConfig struct {
	Buffer   dictx.Dict        // Holds the current configuration in memory
	cfgPath  string            // Path to the main configuration file
	bakPath  string            // Path to the backup configuration file (optional)
	fhandler FileHandler       // file handler for accessing files
	cipher   ciphering.Handler // Cipher handler for encryption and decryption (optional)
}

// New creates a new JConfig instance with the provided file path and default values.
// Returns an error if the file path is empty.
func New(path string, defaults dictx.Dict) (*JConfig, error) {
	path = filepath.Clean(path)
	if path == "" {
		return nil, errors.New("file path cannot be empty")
	}
	if defaults == nil {
		defaults = dictx.Dict{}
	}
	return &JConfig{
		Buffer:   defaults,
		cfgPath:  path,
		fhandler: NewStdFileHandler(),
	}, nil
}

// SetFileHandler sets a new file handler.
func (c *JConfig) SetFileHandler(handler FileHandler) {
	if handler != nil {
		c.fhandler = handler
	}
}

// InitBackup sets the backup file path for the configuration.
// Returns an error if the provided path is empty.
func (c *JConfig) InitBackup(path string) error {
	path = filepath.Clean(path)
	if path == "" {
		return errors.New("file path cannot be empty")
	}
	c.bakPath = path
	return nil
}

// EnableBackup enables automatic backup by creating a backup file
// at the same path as the config file, with a `.backup` suffix.
func (c *JConfig) EnableBackup() {
	c.bakPath = c.cfgPath + ".backup"
}

// IsExist checks whether the main configuration file exists.
func (c *JConfig) IsExist() bool {
	return c.fhandler.IsExist(c.cfgPath)
}

// IsBackupExist checks whether the backup file exists.
func (c *JConfig) IsBackupExist() bool {
	if c.bakPath != "" {
		return c.fhandler.IsExist(c.bakPath)
	}
	return false
}

// load merges the provided byte slice into the current buffer
// after unmarshalling it from JSON.
func (c *JConfig) load(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	var buffer map[string]any
	if err := json.Unmarshal(b, &buffer); err != nil {
		return err
	}
	// Merge the new data into the current buffer
	dictx.Merge(c.Buffer, buffer)
	return nil
}

// Load reads the configuration from the main file and loads it into memory.
// If the main config fails to load, attempts to load from a backup file.
// Also saves the loaded data back to the backup if successful.
func (c *JConfig) Load() error {
	var b []byte
	var err error

	// Attempt to load the primary configuration file
	if c.IsExist() {
		b, err = c.fhandler.Read(c.cfgPath)
		if err == nil {
			if err = c.load(b); err == nil {
				if c.bakPath != "" {
					c.fhandler.Write(c.bakPath, b, 0o664)
				}
				return nil
			}
		}
	}

	// Attempt to load the backup file if the primary failed
	if c.IsBackupExist() {
		b, err = c.fhandler.Read(c.bakPath)
		if err == nil {
			if err = c.load(b); err == nil {
				return c.fhandler.Write(c.cfgPath, b, 0o664)
			}
		}
	}

	return err
}

// Save serializes the current buffer to a formatted JSON byte slice,
// then writes the configuration buffer to both the main file
// and the backup file (if a backup path is set).
func (c *JConfig) Save() error {
	b, err := json.MarshalIndent(c.Buffer, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err = c.fhandler.Write(c.cfgPath, b, 0o664); err != nil {
		return err
	}
	if c.bakPath != "" {
		return c.fhandler.Write(c.bakPath, b, 0o664)
	}
	return nil
}

// Keys returns a list of all keys in the configuration buffer.
func (c *JConfig) Keys() []string {
	return dictx.KeysN(c.Buffer, -1)
}

// Get retrieves a value from the configuration buffer by key.
// If the key is not found, the default_value is returned.
func (c *JConfig) Get(key string, defaultValue any) any {
	return dictx.Get(c.Buffer, key, defaultValue)
}

// Set adds a new value in the configuration buffer by key.
// If the key already exists, its value is overwritten.
func (c *JConfig) Set(key string, newValue any) {
	dictx.Set(c.Buffer, key, newValue)
}

// Merge updates a configuration buffer recursively with an update dictionary.
// It merges keys and values, allowing nested dictionaries to be updated as well.
func (c *JConfig) Merge(updt dictx.Dict) {
	dictx.Merge(c.Buffer, updt)
}

// Delete removes a key from the configuration buffer if it exists.
// It supports nested keys using the separator.
func (c *JConfig) Delete(key string) {
	dictx.Delete(c.Buffer, key)
}

// Purge clears the configuration buffer and deletes the main and
// backup files (if they exist).
func (c *JConfig) Purge() error {
	c.Buffer = dictx.Dict{}
	if c.IsBackupExist() {
		c.fhandler.Remove(c.bakPath)
	}
	if c.IsExist() {
		return c.fhandler.Remove(c.cfgPath)
	}
	return nil
}

///////////////////////////////////////////////////////

// InitCipher initializes ciphering handler
func (c *JConfig) InitCipher(cipher ciphering.Handler) error {
	if cipher != nil {
		c.cipher = cipher
		return nil
	}
	return errors.New("empty cipher handler")
}

// InitAES128 initializes AES-128 encryption for the configuration
// using the provided secret key.
// Returns an error if the secret is invalid or encryption setup fails.
func (c *JConfig) InitAES128(secret string) error {
	cipher, err := ciphering.NewAES128(secret)
	if err != nil {
		return err
	}
	c.cipher = cipher
	return nil
}

// InitAES256 initializes AES-256 encryption for the configuration
// using the provided secret key.
// Returns an error if the secret is invalid or encryption setup fails.
func (c *JConfig) InitAES256(secret string) error {
	cipher, err := ciphering.NewAES256(secret)
	if err != nil {
		return err
	}
	c.cipher = cipher
	return nil
}

// GetSecure retrieves and decrypts a secure value by key from the configuration.
// If the key does not exist or decryption fails, it returns the defaultValue.
// Returns an error if encryption is not configured or the value format is invalid.
func (c *JConfig) GetSecure(key string, defaultValue any) (any, error) {
	if c.cipher == nil {
		return nil, fmt.Errorf("ciphering is not configured")
	}
	// Retrieve the encrypted value from the buffer
	data := dictx.Get(c.Buffer, key, nil)
	if data == nil {
		return defaultValue, nil
	}
	// Ensure the value is a base64 encoded string
	if encryptedStr, ok := data.(string); ok && len(encryptedStr) > 0 {
		encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedStr)
		if err != nil {
			return nil, err
		}
		decryptedBytes, err := c.cipher.Decrypt(encryptedBytes)
		if err != nil {
			return nil, err
		}
		var val any
		err = json.Unmarshal(decryptedBytes, &val)
		if err != nil {
			return nil, err
		}
		return val, nil
	}

	// Invalid value format or empty string
	return nil, fmt.Errorf("invalid value format")
}

// SetSecure encrypts and stores a secure value by key in the configuration.
// The key is created if it doesn't exist.
// Returns an error if encryption is not configured.
func (c *JConfig) SetSecure(key string, val any) error {
	if c.cipher == nil {
		return fmt.Errorf("ciphering is not configured")
	}
	valBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	encryptedBytes, err := c.cipher.Encrypt(valBytes)
	if err != nil {
		return err
	}
	encryptedStr := base64.StdEncoding.EncodeToString(encryptedBytes)
	dictx.Set(c.Buffer, key, encryptedStr)
	return nil
}
