package types

import (
	"slices"
)

// Simple Dict type
type Dict map[string]any

// create new Dict type from initial map data
func CreateDict(d Dict) Dict {
	if d == nil {
		return make(Dict)
	}
	return d
}

// return sorted list of all keys
func (d Dict) Keys() []string {
	keys := []string{}
	for k := range d {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// check if key exist in dict
func (d Dict) KeyExist(key string) bool {
	_, ok := d[key]
	return ok
}

// get value from dict by key or return default value
func (d Dict) Get(key string, defval any) any {
	if val, ok := d[key]; ok {
		return val
	}
	return defval
}
func (d Dict) GetBool(key string, defval bool) bool {
	if v, ok := d[key].(bool); ok {
		return v
	}
	return defval
}
func (d Dict) GetString(key string, defval string) string {
	if v, ok := d[key].(string); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt(key string, defval int) int {
	if v, ok := d[key].(int); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt8(key string, defval int8) int8 {
	if v, ok := d[key].(int8); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt16(key string, defval int16) int16 {
	if v, ok := d[key].(int16); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt32(key string, defval int32) int32 {
	if v, ok := d[key].(int32); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt64(key string, defval int64) int64 {
	if v, ok := d[key].(int64); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint(key string, defval uint) uint {
	if v, ok := d[key].(uint); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint8(key string, defval uint8) uint8 {
	if v, ok := d[key].(uint8); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint16(key string, defval uint16) uint16 {
	if v, ok := d[key].(uint16); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint32(key string, defval uint32) uint32 {
	if v, ok := d[key].(uint32); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint64(key string, defval uint64) uint64 {
	if v, ok := d[key].(uint64); ok {
		return v
	}
	return defval
}
func (d Dict) GetFloat32(key string, defval float32) float32 {
	if v, ok := d[key].(float32); ok {
		return v
	}
	return defval
}
func (d Dict) GetFloat64(key string, defval float64) float64 {
	if v, ok := d[key].(float64); ok {
		return v
	}
	return defval
}

// set value in dict by key
func (d Dict) Set(key string, newval any) {
	d[key] = newval
}

// delete value from dict by key
func (d Dict) Delete(key string) {
	if _, ok := d[key]; ok {
		delete(d, key)
	}
}

// update dict with updt dict
func (d Dict) Update(updt Dict) {
	for _, key := range updt.Keys() {
		d[key] = updt[key]
	}
}
