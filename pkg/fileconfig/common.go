package fileconfig

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	ErrValNotExists = fmt.Errorf("value not exists")
)

type FileConfig interface {
	Load(map[string]any) error
	Save() error
	Dump() ([]byte, error)
	Get(string, bool) (any, error)
	Set(string, any, bool) error
	Delete(string) error
	Purge() error
	Keys() []string
	Buffer() map[string]any
}

type fileConfig struct {
	FilePath   string
	subkeyChar string
	Buffer     map[string]any
}

// newFileConfig creates a new instance of fileConfig
func newFileConfig(filePath string) *fileConfig {
	config := &fileConfig{
		FilePath:   filePath,
		subkeyChar: ".",
		Buffer:     make(map[string]any),
	}

	return config
}

func (fc *fileConfig) mangle(blob []byte) []byte {
	res := make([]byte, 0, len(blob))
	l := len(blob) - 1
	for i := 0; i < l; i += 2 {
		if (i + 1) < l {
			res = append(res, blob[i+1])
		}
		res = append(res, blob[i])
	}
	res = append(res, blob[l])
	return res
}

func (fc *fileConfig) demangle(blob []byte) []byte {
	return fc.mangle(blob)
}

func (fc *fileConfig) encode(value any) (string, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	return hex.EncodeToString(fc.mangle([]byte(encodedData))), nil
}

func (fc *fileConfig) decode(value any) (any, error) {
	decodedData, err := hex.DecodeString(fmt.Sprintf("%v", value))
	if err != nil {
		return nil, err
	}

	decodedData = fc.demangle(decodedData)
	decodedJSON, err := base64.StdEncoding.DecodeString(string(decodedData))
	if err != nil {
		return nil, err
	}

	var result any
	err = json.Unmarshal(decodedJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Get gets a certain key from config
func (fc *fileConfig) Get(key string, decode bool) (any, error) {
	keys := strings.Split(key, fc.subkeyChar)
	val, err := fc.getNestedValue(fc.Buffer, keys)
	if err != nil {
		return nil, ErrValNotExists
	}

	if decode {
		return fc.decode(val)
	}

	return val, nil
}

// Set sets a certain key in config
func (fc *fileConfig) Set(key string, value any, encode bool) error {
	if encode {
		encVal, err := fc.encode(value)
		if err != nil {
			return err
		}
		value = encVal
	}

	if strings.Contains(key, fc.subkeyChar) {
		keys := strings.Split(key, fc.subkeyChar)
		subDict := make(map[string]any)
		subDict[keys[len(keys)-1]] = value
		for i := len(keys) - 2; i >= 0; i-- {
			tempDict := make(map[string]any)
			tempDict[keys[i]] = subDict
			subDict = tempDict
		}
		fc.mergeDicts(fc.Buffer, subDict)
	} else {
		fc.Buffer[key] = value
	}

	return nil
}

func (fc *fileConfig) Delete(key string) error {
	if strings.Contains(key, fc.subkeyChar) {
		keys := strings.Split(key, fc.subkeyChar)
		parentDict := fc.Buffer
		for i := 0; i < len(keys)-1; i++ {
			if val, ok := parentDict[keys[i]]; ok {
				if nestedDict, nestedOk := val.(map[string]any); nestedOk {
					parentDict = nestedDict
				} else {
					return fmt.Errorf("Cannot delete key from non-dictionary type")
				}
			} else {
				return fmt.Errorf("Key does not exist")
			}
		}
		delete(parentDict, keys[len(keys)-1])
	} else {
		delete(fc.Buffer, key)
	}
	return nil
}

func (fc *fileConfig) Purge() error {
	fc.Buffer = make(map[string]any)
	if _, err := os.Stat(fc.FilePath); !os.IsNotExist(err) {
		if err := os.Remove(fc.FilePath); err != nil {
			return err
		}
	}
	return nil
}

// Keys returns a slice of all keys in the buffer
func (fc *fileConfig) Keys() []string {
	if fc.Buffer == nil {
		return nil
	}

	keys := make([]string, 0, len(fc.Buffer))
	for key := range fc.Buffer {
		keys = append(keys, key)
	}

	return keys
}

func (fc *fileConfig) mergeDicts(src, updt map[string]any) {
	for k, v := range updt {
		if srcVal, ok := src[k]; ok {
			if srcDict, srcDictOk := srcVal.(map[string]any); srcDictOk {
				if updtDict, updtDictOk := v.(map[string]any); updtDictOk {
					fc.mergeDicts(srcDict, updtDict)
				} else {
					src[k] = v
				}
			} else {
				src[k] = v
			}
		} else {
			src[k] = v
		}
	}
}

// getNestedValue recursively retrieves a nested value from the config buffer
func (fc *fileConfig) getNestedValue(data map[string]any, keys []string) (any, error) {
	for _, k := range keys {
		val, ok := data[k]
		if !ok {
			return nil, fmt.Errorf("Key not found: %s", strings.Join(keys, fc.subkeyChar))
		}

		if nestedData, ok := val.(map[string]any); ok {
			data = nestedData
		} else {
			return val, nil
		}
	}
	return nil, fmt.Errorf("Invalid key: %s", strings.Join(keys, fc.subkeyChar))
}

// isValidKey checks if a key is valid (non-empty and of string type)
func isValidKey(key string) bool {
	return key != "" && reflect.TypeOf(key).Kind() == reflect.String
}
