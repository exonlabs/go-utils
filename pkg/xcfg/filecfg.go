package xcfg

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/exonlabs/go-utils/pkg/types"
)

type Buffer = types.NDict

// configuration file handler
type FileConfig struct {
	Buffer

	// config filepath on disk
	filePath string
	// data format handler
	fileHandler Handler
	// flag for binary file format
	isblob bool

	// secure data encoding and decoding handler
	dataEncoder Encoder
}

// create new json config file handler
func NewJsonConfig(filePath string, defaults map[string]any) *FileConfig {
	return &FileConfig{
		Buffer:      types.CreateNDict(defaults),
		filePath:    filePath,
		fileHandler: &JsonHandler{},
		dataEncoder: &defaultEncoder{},
	}
}

// create new blob config file handler
func NewBlobConfig(filePath string, defaults map[string]any) *FileConfig {
	return &FileConfig{
		Buffer:      types.CreateNDict(defaults),
		filePath:    filePath,
		fileHandler: &BlobHandler{},
		isblob:      true,
		dataEncoder: &defaultEncoder{},
	}
}

// implement the Stringer interface
func (cfg *FileConfig) String() string {
	return fmt.Sprintf("<%sConfig: %s %v>",
		cfg.fileHandler.Type(), cfg.filePath, cfg.Buffer)
}

// set new encoder
func (cfg *FileConfig) SetEncoder(enc Encoder) {
	if enc != nil {
		cfg.dataEncoder = enc
	}
}

// load data from file and update current config buffer
func (cfg *FileConfig) Load() error {
	if err := cfg.fileHandler.Load(cfg); err != nil {
		return fmt.Errorf("%w, %s", ErrLoadFailed, err.Error())
	}
	return nil
}

// save current config buffer to file
func (cfg *FileConfig) Save() error {
	if err := cfg.fileHandler.Save(cfg); err != nil {
		return fmt.Errorf("%w, %s", ErrSaveFailed, err.Error())
	}
	return nil
}

// delete config file from disk and purge reset config buffer
func (cfg *FileConfig) Purge() error {
	if _, err := os.Stat(cfg.filePath); !os.IsNotExist(err) {
		if err := os.Remove(cfg.filePath); err != nil {
			return fmt.Errorf("%w%s", ErrError, err.Error())
		}
	}
	cfg.Buffer = Buffer(nil)
	return nil
}

// return the raw byte contents of config file
func (cfg *FileConfig) Dump() ([]byte, error) {
	if _, err := os.Stat(cfg.filePath); os.IsNotExist(err) {
		return nil, ErrFileNotExist
	}
	data, err := os.ReadFile(cfg.filePath)
	if err != nil {
		return nil, fmt.Errorf("%w%s", ErrError, err.Error())
	}
	return data, nil
}

// get secure value from config by key
func (cfg *FileConfig) GetSecure(key string, defval any) (any, error) {
	if cfg.isblob {
		return cfg.Get(key, defval), nil
	}
	data := cfg.Get(key, nil)
	if data == nil {
		// key not exist
		return defval, nil
	}
	err := errors.New("invalid data format")
	if d, ok := data.(string); ok {
		b, err := hex.DecodeString(d)
		if err == nil {
			val, err := cfg.dataEncoder.Decode(b)
			if err == nil {
				return val, nil
			}
		}
	}
	return nil, fmt.Errorf("%w, %s", ErrDecodeFailed, err.Error())
}

// set secure value in config by key, creates key if not exist
func (cfg *FileConfig) SetSecure(key string, newval any) error {
	if cfg.isblob {
		cfg.Set(key, newval)
		return nil
	}
	if b, err := cfg.dataEncoder.Encode(newval); err != nil {
		return fmt.Errorf("%w, %s", ErrEncodeFailed, err.Error())
	} else {
		cfg.Set(key, hex.EncodeToString(b))
	}
	return nil
}
