package xcfg

import (
	"encoding/json"
)

// json file config
type JsonConfig struct {
	*fileConfig
}

// create new json config file handler
func NewJsonConfig(filepath string, defaults map[string]any) *JsonConfig {
	cfg := &JsonConfig{
		fileConfig: newFileConfig(filepath, defaults),
	}
	cfg.fileConfig.loader = cfg.load
	cfg.fileConfig.dumper = cfg.dump
	return cfg
}

// parse raw bytes json config and update local buffer
func (cfg *JsonConfig) load(b []byte) error {
	if b != nil {
		// parse
		var data map[string]any
		if err := json.Unmarshal(b, &data); err != nil {
			return err
		}
		// update local buffer
		cfg.Update(data)
	}
	return nil
}

// serialize config buffer to raw bytes json
func (cfg *JsonConfig) dump() ([]byte, error) {
	// generate raw json data bytes
	b, err := json.MarshalIndent(cfg.Buffer, "", "  ")
	if err != nil {
		return nil, err
	}
	// add newline ending
	b = append(b, 0x0A)
	return b, nil
}
