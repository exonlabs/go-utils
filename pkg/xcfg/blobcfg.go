package xcfg

import (
	"errors"
	"fmt"
	"os"

	"github.com/exonlabs/go-utils/pkg/types"
)

// binary configuration file handler
type BlobConfig struct {
	*BaseFileConfig
}

func NewBlobConfig(filePath string, defaults types.Dict) *BlobConfig {
	return &BlobConfig{
		BaseFileConfig: NewBaseFileConfig(filePath, defaults),
	}
}

func (bc *BlobConfig) String() string {
	return fmt.Sprintf("<BlobConfig: %s %v>", bc.filePath, bc.Dict)
}

// load config buffer from file and merge with defaults
func (bc *BlobConfig) Load() error {
	// read raw bytes data
	b, err := bc.Dump()
	if err != nil {
		if errors.Is(err, ErrFileNotExist) {
			return nil
		}
		return err
	}

	// parse
	var data any
	data, err = bc.decode(b)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrLoadFailed, err.Error())
	}

	// update existing config with data buffer
	if d, ok := data.(map[string]any); ok {
		bc.Update(d)
	} else {
		return fmt.Errorf("%w, invalid file content", ErrLoadFailed)
	}
	return nil
}

// save the current config buffer to file
func (bc *BlobConfig) Save() error {
	// generate raw bytes blob data
	b, err := bc.encode(bc.Dict)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrSaveFailed, err.Error())
	}
	// add newline ending
	b = append(b, 0x0A)

	// write bytes to file
	err = os.WriteFile(bc.filePath, b, 0o666)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrSaveFailed, err.Error())
	}
	return nil
}
