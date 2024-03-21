package xcfg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/exonlabs/go-utils/pkg/crypto/xcipher"
	"github.com/exonlabs/go-utils/pkg/types"
)

type Buffer = types.NDict

// configuration file handler
type fileConfig struct {
	Buffer

	// config filepath on disk
	filepath string
	bakpath  string

	// load and dump callback func pointers
	loader func([]byte) error
	dumper func() ([]byte, error)

	// cipher object
	cipher xcipher.Cipher
}

// create new json config file handler
func newFileConfig(filepath string, defaults map[string]any) *fileConfig {
	return &fileConfig{
		Buffer:   types.CreateNDict(defaults),
		filepath: filepath,
	}
}

// enable config file backup support
func (cfg *fileConfig) EnableBackup(bakpath string) {
	if bakpath != "" {
		cfg.bakpath = bakpath
	} else {
		cfg.bakpath = cfg.filepath + ".backup"
	}
}

// reset local buffer
func (cfg *fileConfig) Reset() {
	cfg.Buffer = Buffer(nil)
}

// check config file exist on disk
func (cfg *fileConfig) IsExist() bool {
	_, err := os.Stat(cfg.filepath)
	return !os.IsNotExist(err)
}

// check backup config file exist on disk, if backup support enabled.
func (cfg *fileConfig) IsBackupExist() bool {
	if cfg.bakpath != "" {
		_, err := os.Stat(cfg.bakpath)
		return !os.IsNotExist(err)
	}
	return false
}

// read raw bytes content of config file, if error: then check and
// read the backup file if backup support enabled.
func (cfg *fileConfig) Load() error {
	var b []byte
	var err error
	if cfg.IsExist() {
		b, err = os.ReadFile(cfg.filepath)
		if err == nil {
			err = cfg.loader(b)
			if err == nil {
				cfg.saveBackup(b)
				return nil
			}
		}
	}
	if cfg.IsBackupExist() {
		b, err = os.ReadFile(cfg.bakpath)
		if err == nil {
			err = cfg.loader(b)
			if err == nil {
				return cfg.Save()
			}
		}
	}
	return err
}

// write raw bytes content to config file, if not error: then check and
// write backup config if backup support enabled.
func (cfg *fileConfig) Save() error {
	b, err := cfg.dumper()
	if err != nil {
		return err
	}
	err = os.WriteFile(cfg.filepath, b, 0o666)
	if err != nil {
		return err
	}
	return cfg.saveBackup(b)
}
func (cfg *fileConfig) saveBackup(b []byte) error {
	if cfg.bakpath != "" {
		return os.WriteFile(cfg.bakpath, b, 0o666)
	}
	return nil
}

// delete config and backup files from disk and reset local buffer
func (cfg *fileConfig) Purge() error {
	cfg.Reset()
	if cfg.IsBackupExist() {
		os.Remove(cfg.bakpath)
	}
	if cfg.IsExist() {
		return os.Remove(cfg.filepath)
	}
	return nil
}

// //////////////////////////////////////////

func (cfg *fileConfig) SetAES128(secret string) error {
	cipher, err := xcipher.NewAES128(secret)
	if err != nil {
		return err
	}
	cfg.cipher = cipher
	return nil
}

func (cfg *fileConfig) SetAES256(secret string) error {
	cipher, err := xcipher.NewAES256(secret)
	if err != nil {
		return err
	}
	cfg.cipher = cipher
	return nil
}

// get secure value from config by key
func (cfg *fileConfig) GetSecure(key string, defval any) (any, error) {
	if cfg.cipher == nil {
		return nil, fmt.Errorf("security is not configured")
	}
	data := cfg.Get(key, nil)
	if data == nil {
		// key not exist
		return defval, nil
	}
	if d, ok := data.(string); ok {
		if len(d) == 0 {
			// empty key
			return defval, nil
		}
		b, err := base64.StdEncoding.DecodeString(d)
		if err != nil {
			return nil, err
		}
		b, err = cfg.cipher.Decrypt(b)
		if err != nil {
			return nil, err
		}
		var val any
		err = json.Unmarshal(b, &val)
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return nil, fmt.Errorf("invalid value format")
}

// set secure value in config by key, creates key if not exist
func (cfg *fileConfig) SetSecure(key string, val any) error {
	if cfg.cipher == nil {
		return fmt.Errorf("security is not configured")
	}
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	b, err = cfg.cipher.Encrypt(b)
	if err != nil {
		return err
	}
	cfg.Set(key, base64.StdEncoding.EncodeToString(b))
	return nil
}
