package types

import (
	"slices"
	"strings"
)

const (
	// default nested keys seperator
	defaultKeySep = string(".")
)

// Dict type with nested keys support
type Dict map[string]any

// initialize new Dict with map type conversion
func NewDict(d Dict) Dict {
	if d == nil {
		return make(Dict)
	}
	for key, val := range d {
		switch v := val.(type) {
		case Dict:
			d[key] = NewDict(v)
		case map[string]any:
			d[key] = NewDict(v)
		}
	}
	return d
}

// return sorted list of all nested level keys
func (d Dict) Keys() []string {
	return d.KeysN(0)
}

// return sorted recursive list up to N level nested keys
func (d Dict) KeysN(lvl int) []string {
	keys := []string{}
	for k := range d {
		if v, ok := d[k].(Dict); ok && lvl != 1 {
			for _, sk := range v.KeysN(lvl - 1) {
				keys = append(keys, k+defaultKeySep+sk)
			}
		} else {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}

// check if key exist in dict
func (d Dict) KeyExist(key string) bool {
	k := strings.SplitN(key, defaultKeySep, 2)
	if val, ok := d[k[0]]; ok {
		// if not nested key
		if len(k) < 2 {
			return true
		}
		// if value is of type Dict
		if v, ok := val.(Dict); ok {
			return v.KeyExist(k[1])
		}
	}
	return false
}

// get value from dict by key or return default value
func (d Dict) Get(key string, defval any) any {
	k := strings.SplitN(key, defaultKeySep, 2)
	if val, ok := d[k[0]]; ok {
		// if not nested key
		if len(k) < 2 {
			return val
		}
		// if value is of type Dict
		if v, ok := val.(Dict); ok {
			return v.Get(k[1], defval)
		}
	}
	return defval
}
func (d Dict) GetBool(key string, defval bool) bool {
	if v, ok := d.Get(key, defval).(bool); ok {
		return v
	}
	return defval
}
func (d Dict) GetString(key string, defval string) string {
	if v, ok := d.Get(key, defval).(string); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt(key string, defval int) int {
	if v, ok := d.Get(key, defval).(int); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt8(key string, defval int8) int8 {
	if v, ok := d.Get(key, defval).(int8); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt16(key string, defval int16) int16 {
	if v, ok := d.Get(key, defval).(int16); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt32(key string, defval int32) int32 {
	if v, ok := d.Get(key, defval).(int32); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt64(key string, defval int64) int64 {
	if v, ok := d.Get(key, defval).(int64); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint(key string, defval uint) uint {
	if v, ok := d.Get(key, defval).(uint); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint8(key string, defval uint8) uint8 {
	if v, ok := d.Get(key, defval).(uint8); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint16(key string, defval uint16) uint16 {
	if v, ok := d.Get(key, defval).(uint16); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint32(key string, defval uint32) uint32 {
	if v, ok := d.Get(key, defval).(uint32); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint64(key string, defval uint64) uint64 {
	if v, ok := d.Get(key, defval).(uint64); ok {
		return v
	}
	return defval
}
func (d Dict) GetFloat32(key string, defval float32) float32 {
	if v, ok := d.Get(key, defval).(float32); ok {
		return v
	}
	return defval
}
func (d Dict) GetFloat64(key string, defval float64) float64 {
	if v, ok := d.Get(key, defval).(float64); ok {
		return v
	}
	return defval
}

// set value in dict by key
func (d Dict) Set(key string, newval any) {
	k := strings.SplitN(key, defaultKeySep, 2)
	// if not nested key
	if len(k) < 2 {
		d[k[0]] = newval
		return
	}
	// if 1st level key not exist or not of type Dict
	if _, ok := d[k[0]].(Dict); !ok {
		d[k[0]] = make(Dict)
	}
	val := d[k[0]].(Dict)
	val.Set(k[1], newval)
}

// delete value from dict by key
func (d Dict) Delete(key string) {
	k := strings.SplitN(key, defaultKeySep, 2)
	if val, ok := d[k[0]]; ok {
		// if not nested key
		if len(k) < 2 {
			delete(d, k[0])
			return
		}
		// if value is of type Dict
		if v, ok := val.(Dict); ok {
			v.Delete(k[1])
			return
		}
	}
	return
}

// update dict with updt dict
func (d Dict) Update(updt Dict) {
	buffer := NewDict(updt)
	for _, k := range buffer.Keys() {
		d.Set(k, buffer.Get(k, nil))
	}
}
