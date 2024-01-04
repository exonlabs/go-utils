package xcfg

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var (
	ErrError        = errors.New("")
	ErrLoadFailed   = fmt.Errorf("%wloading failed", ErrError)
	ErrSaveFailed   = fmt.Errorf("%wsaving failed", ErrError)
	ErrFileNotExist = fmt.Errorf("%wfile does not exist", ErrError)
	ErrEncodeFailed = fmt.Errorf("%wencoding failed", ErrError)
	ErrDecodeFailed = fmt.Errorf("%wdecoding failed", ErrError)
)

type FileConfig interface {
	Load() error
	Save() error
	Purge() error
	Dump() ([]byte, error)
	Buffer() Dict
	Keys() []string
	KeysN(int) []string
	KeyExist(string) bool
	Get(string, any) any
	Set(string, any)
	Delete(string)
	GetSecure(string, any) (any, error)
	SetSecure(string, any) error
}

// base configuration file handler
type BaseFileConfig struct {
	Dict
	filePath string

	// secure data encoding and decoding callback
	Encode func([]byte) ([]byte, error)
	Decode func([]byte) ([]byte, error)
}

func NewBaseFileConfig(filePath string, defaults Dict) *BaseFileConfig {
	if defaults == nil {
		defaults = make(Dict)
	}
	return &BaseFileConfig{
		Dict:     NewDict(defaults),
		filePath: filePath,
	}
}

func (fc *BaseFileConfig) Purge() error {
	fc.Dict = make(Dict)
	if _, err := os.Stat(fc.filePath); !os.IsNotExist(err) {
		if err := os.Remove(fc.filePath); err != nil {
			return fmt.Errorf("%w%s", ErrError, err.Error())
		}
	}
	return nil
}

// return the contents of the config file
func (fc *BaseFileConfig) Dump() ([]byte, error) {
	if _, err := os.Stat(fc.filePath); os.IsNotExist(err) {
		return nil, ErrFileNotExist
	}
	data, err := os.ReadFile(fc.filePath)
	if err != nil {
		return nil, fmt.Errorf("%w%s", ErrError, err.Error())
	}
	return data, nil
}

// return handler to internal Dict object
func (fc *BaseFileConfig) Buffer() Dict {
	return fc.Dict
}

// get secure value from config by key
func (fc *BaseFileConfig) GetSecure(key string, defval any) (any, error) {
	if !fc.KeyExist(key) {
		return defval, nil
	}

	var b []byte
	var err error

	if val, ok := fc.Get(key, nil).(string); ok {
		if b, err = hex.DecodeString(val); err != nil {
			return nil, fmt.Errorf("%w, %s", ErrDecodeFailed, err.Error())
		}
	} else {
		return nil, fmt.Errorf("%w, invalid value type", ErrDecodeFailed)
	}
	return fc.decode(b)
}

// set secure value in config by key, creates key if not exist
func (fc *BaseFileConfig) SetSecure(key string, newval any) error {
	b, err := fc.encode(newval)
	if err != nil {
		return err
	}
	fc.Set(key, hex.EncodeToString(b))
	return nil
}

// data encoding with byte mangling
func (fc *BaseFileConfig) encode(data any) ([]byte, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", ErrEncodeFailed, err.Error())
	}

	// if external function defined
	if fc.Encode != nil {
		b, err = fc.Encode(b)
		if err != nil {
			return nil, fmt.Errorf("%w, %s", ErrEncodeFailed, err.Error())
		}
	} else {
		b = _mangle([]byte(base64.StdEncoding.EncodeToString(b)))
	}
	return b, nil
}

// encoding with byte mangling
func (fc *BaseFileConfig) decode(data []byte) (any, error) {
	var b []byte
	var err error

	// if external function defined
	if fc.Decode != nil {
		b, err = fc.Decode(data)
	} else {
		b, err = base64.StdEncoding.DecodeString(string(_mangle(data)))
	}
	if err != nil {
		return nil, fmt.Errorf("%w, %s", ErrDecodeFailed, err.Error())
	}

	var result any
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, fmt.Errorf("%w, %s", ErrDecodeFailed, err.Error())
	}
	return result, nil
}

// mangle bytes data to avoid easy detectable patterns.
// inverting every 2 bytes order
func _mangle(b []byte) []byte {
	res := make([]byte, 0, len(b))
	l := len(b) - 1
	for i := 0; i < l; i += 2 {
		if (i + 1) < l {
			res = append(res, b[i+1])
		}
		res = append(res, b[i])
	}
	res = append(res, b[l])
	return res
}
