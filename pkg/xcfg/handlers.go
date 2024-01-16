package xcfg

import (
	"encoding/json"
	"errors"
	"os"
)

type Handler interface {
	Type() string
	Load(*FileConfig) error
	Save(*FileConfig) error
}

// //////////////////////////////////////////////////

// JSON file handler
type JsonHandler struct{}

func (h *JsonHandler) Type() string {
	return "Json"
}

// load data from file and update current config buffer
func (h *JsonHandler) Load(cfg *FileConfig) error {
	// read raw byte contents
	b, err := cfg.Dump()
	if err != nil {
		if errors.Is(err, ErrFileNotExist) {
			return nil
		}
		return err
	}
	// parse
	var data map[string]any
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	} else {
		// update existing config with loaded data
		cfg.Update(data)
	}
	return nil
}

// save current config buffer to file
func (h *JsonHandler) Save(cfg *FileConfig) error {
	// generate raw json data bytes
	b, err := json.MarshalIndent(cfg.Buffer, "", "  ")
	if err != nil {
		return err
	}
	// add newline ending
	b = append(b, 0x0A)
	// write bytes to file
	err = os.WriteFile(cfg.filePath, b, 0o666)
	if err != nil {
		return err
	}
	return nil
}

// //////////////////////////////////////////////////

// binary file handler
type BlobHandler struct{}

func (h *BlobHandler) Type() string {
	return "Blob"
}

// load data from file and update current config buffer
func (h *BlobHandler) Load(cfg *FileConfig) error {
	// read raw bytes data
	b, err := cfg.Dump()
	if err != nil {
		if errors.Is(err, ErrFileNotExist) {
			return nil
		}
		return err
	}
	// parse
	if d, err := cfg.dataEncoder.Decode(b); err != nil {
		return err
	} else if data, ok := d.(map[string]any); ok {
		// update existing config with loaded data
		cfg.Update(data)
	} else {
		return errors.New("invalid file content")
	}
	return nil
}

// save current config buffer to file
func (h *BlobHandler) Save(cfg *FileConfig) error {
	// generate raw bytes blob data
	b, err := cfg.dataEncoder.Encode(cfg.Buffer)
	if err != nil {
		return err
	}
	// write bytes to file
	err = os.WriteFile(cfg.filePath, b, 0o666)
	if err != nil {
		return err
	}
	return nil
}
