package xcfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/exonlabs/go-utils/pkg/types"
)

// JSON configuration file handler
type JsonConfig struct {
	*BaseFileConfig
}

func NewJsonConfig(filePath string, defaults types.Dict) *JsonConfig {
	return &JsonConfig{
		BaseFileConfig: NewBaseFileConfig(filePath, defaults),
	}
}

func (jc *JsonConfig) String() string {
	return fmt.Sprintf("<JsonConfig: %s %v>", jc.filePath, jc.Dict)
}

// load config buffer from file and merge with defaults
func (jc *JsonConfig) Load() error {
	// read raw bytes data
	b, err := jc.Dump()
	if err != nil {
		if errors.Is(err, ErrFileNotExist) {
			return nil
		}
		return err
	}

	// parse
	var data map[string]any
	if err = json.Unmarshal(b, &data); err != nil {
		return fmt.Errorf("%w, %s", ErrLoadFailed, err.Error())
	}

	// update existing config with data buffer
	jc.Update(data)
	return nil
}

// save the current config buffer to file
func (jc *JsonConfig) Save() error {
	// generate raw bytes json data
	b, err := json.MarshalIndent(jc.Dict, "", "  ")
	if err != nil {
		return fmt.Errorf("%w, %s", ErrSaveFailed, err.Error())
	}
	// add newline ending
	b = append(b, 0x0A)

	// write bytes to file
	err = os.WriteFile(jc.filePath, b, 0o666)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrSaveFailed, err.Error())
	}
	return nil
}
