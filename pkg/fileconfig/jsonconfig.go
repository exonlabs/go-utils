package fileconfig

import (
	"encoding/json"
	"fmt"
	"os"
)

// JsonFileConfig represents a JSON configuration file handler
type JsonFileConfig struct {
	*fileConfig
}

// NewJsonConfig creates a new instance of JsonFileConfig
func NewJsonConfig(filePath string, data map[string]any) (FileConfig, error) {
	if data == nil {
		data = make(map[string]any)
	}
	var jc JsonFileConfig
	jc.fileConfig = newFileConfig(filePath)
	if err := jc.Load(data); err != nil {
		return nil, err
	}
	return &jc, nil
}

// Load loads config buffer from file and merges with defaults
func (jc *JsonFileConfig) Load(data map[string]any) error {
	var value map[string]any
	if _, err := os.Stat(jc.FilePath); err == nil {
		file, err := os.ReadFile(jc.FilePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(file, &value); err != nil {
			return fmt.Errorf("invalid data contents of file: %s", jc.FilePath)
		}
	}
	jc.fileConfig.Buffer = data

	return nil
}

// Save saves the current config buffer to the file
func (jc *JsonFileConfig) Save() error {
	if jc.fileConfig.Buffer == nil {
		return fmt.Errorf("invalid buffer data type")
	}

	file, err := json.MarshalIndent(jc.fileConfig.Buffer, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(jc.fileConfig.FilePath, file, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// Dump returns the contents of the config file
func (jc *JsonFileConfig) Dump() ([]byte, error) {
	_, err := os.Stat(jc.fileConfig.FilePath)
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(jc.fileConfig.FilePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Buffer returns the current config buffer
func (jc *JsonFileConfig) Buffer() map[string]any {
	return jc.fileConfig.Buffer
}
