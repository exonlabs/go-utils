package types

import (
	"slices"
	"strings"
)

const (
	// nested keys seperator
	sepNDict = "."
)

// Nested Dict type with nested keys support
type NDict map[string]any

// create new NDict from initial map data
func CreateNDict(buff map[string]any) NDict {
	if buff == nil {
		return NDict{}
	}
	for key, value := range buff {
		switch val := value.(type) {
		case map[string]any:
			buff[key] = CreateNDict(val)
		case Dict:
			buff[key] = CreateNDict(val)
		case NDict:
			buff[key] = CreateNDict(val)
		case []map[string]any:
			b := []NDict{}
			for _, v := range val {
				b = append(b, CreateNDict(v))
			}
			buff[key] = b
		case []Dict:
			b := []NDict{}
			for _, v := range val {
				b = append(b, CreateNDict(v))
			}
			buff[key] = b
		case []NDict:
			b := []NDict{}
			for _, v := range val {
				b = append(b, CreateNDict(v))
			}
			buff[key] = b
		case []any:
			b := []any{}
			for _, v := range val {
				switch v.(type) {
				case map[string]any, Dict, NDict:
					b = append(b, CreateNDict(v.(map[string]any)))
				default:
					b = append(b, v)
				}
			}
			buff[key] = b
		}
	}
	return NDict(buff)
}

// recursive convert NDict into standard map data
func StripNDict(buff map[string]any) map[string]any {
	if buff == nil {
		return map[string]any{}
	}
	for key, value := range buff {
		switch val := value.(type) {
		case map[string]any:
			buff[key] = StripNDict(val)
		case Dict:
			buff[key] = StripNDict(val)
		case NDict:
			buff[key] = StripNDict(val)
		case []map[string]any:
			b := []map[string]any{}
			for _, v := range val {
				b = append(b, StripNDict(v))
			}
			buff[key] = b
		case []Dict:
			b := []map[string]any{}
			for _, v := range val {
				b = append(b, StripNDict(v))
			}
			buff[key] = b
		case []NDict:
			b := []map[string]any{}
			for _, v := range val {
				b = append(b, StripNDict(v))
			}
			buff[key] = b
		case []any:
			b := []any{}
			for _, sv := range val {
				switch v := sv.(type) {
				case map[string]any:
					b = append(b, StripNDict(v))
				case Dict:
					b = append(b, StripNDict(v))
				case NDict:
					b = append(b, StripNDict(v))
				default:
					b = append(b, v)
				}
			}
			buff[key] = b
		}
	}
	return buff
}

// return list up to N level nested _keys
func (d NDict) _keys(lvl int) []string {
	keys := []string{}
	for k := range d {
		if lvl != 1 {
			if v, ok := d[k].(NDict); ok {
				for _, sk := range v._keys(lvl - 1) {
					keys = append(keys, k+sepNDict+sk)
				}
				continue
			}
		}
		keys = append(keys, k)
	}
	return keys
}

// return sorted list of all nested level keys
func (d NDict) Keys() []string {
	keys := d._keys(-1)
	if len(keys) > 0 {
		slices.Sort(keys)
	}
	return keys
}

// return sorted recursive list up to N level nested keys
func (d NDict) KeysN(lvl int) []string {
	keys := d._keys(lvl)
	if len(keys) > 0 {
		slices.Sort(keys)
	}
	return keys
}

// check if key exist in dict
func (d NDict) KeyExist(key string) bool {
	k0, kn, next := strings.Cut(key, sepNDict)
	if val, ok := d[k0]; ok {
		// not nested key
		if !next {
			return true
		}
		// value is of type Dict
		if v, ok := val.(NDict); ok {
			return v.KeyExist(kn)
		}
	}
	return false
}

// get value from dict by key or return default value
func (d NDict) Get(key string, defval any) any {
	k0, kn, next := strings.Cut(key, sepNDict)
	if val, ok := d[k0]; ok {
		// not nested key
		if !next {
			return val
		}
		// value is of type Dict
		if v, ok := val.(NDict); ok {
			return v.Get(kn, defval)
		}
	}
	return defval
}

func (d NDict) GetBool(key string, defval bool) bool {
	if v, ok := d.Get(key, defval).(bool); ok {
		return v
	}
	return defval
}
func (d NDict) GetString(key string, defval string) string {
	if v, ok := d.Get(key, defval).(string); ok {
		return v
	}
	return defval
}
func (d NDict) GetInt(key string, defval int) int {
	if v, ok := d.Get(key, defval).(int); ok {
		return v
	}
	return defval
}
func (d NDict) GetInt8(key string, defval int8) int8 {
	if v, ok := d.Get(key, defval).(int8); ok {
		return v
	}
	return defval
}
func (d NDict) GetInt16(key string, defval int16) int16 {
	if v, ok := d.Get(key, defval).(int16); ok {
		return v
	}
	return defval
}
func (d NDict) GetInt32(key string, defval int32) int32 {
	if v, ok := d.Get(key, defval).(int32); ok {
		return v
	}
	return defval
}
func (d NDict) GetInt64(key string, defval int64) int64 {
	if v, ok := d.Get(key, defval).(int64); ok {
		return v
	}
	return defval
}
func (d NDict) GetUint(key string, defval uint) uint {
	if v, ok := d.Get(key, defval).(uint); ok {
		return v
	}
	return defval
}
func (d NDict) GetUint8(key string, defval uint8) uint8 {
	if v, ok := d.Get(key, defval).(uint8); ok {
		return v
	}
	return defval
}
func (d NDict) GetUint16(key string, defval uint16) uint16 {
	if v, ok := d.Get(key, defval).(uint16); ok {
		return v
	}
	return defval
}
func (d NDict) GetUint32(key string, defval uint32) uint32 {
	if v, ok := d.Get(key, defval).(uint32); ok {
		return v
	}
	return defval
}
func (d NDict) GetUint64(key string, defval uint64) uint64 {
	if v, ok := d.Get(key, defval).(uint64); ok {
		return v
	}
	return defval
}
func (d NDict) GetFloat32(key string, defval float32) float32 {
	if v, ok := d.Get(key, defval).(float32); ok {
		return v
	}
	return defval
}
func (d NDict) GetFloat64(key string, defval float64) float64 {
	if v, ok := d.Get(key, defval).(float64); ok {
		return v
	}
	return defval
}

// set value in dict by key
func (d NDict) Set(key string, newval any) {
	k0, kn, next := strings.Cut(key, sepNDict)
	// not nested key
	if !next {
		d[k0] = newval
	} else {
		// 1st level key not exist or not of type Dict
		if _, ok := d[k0].(NDict); !ok {
			d[k0] = NDict{}
		}
		d[k0].(NDict).Set(kn, newval)
	}
}

// delete value from dict by key
func (d NDict) Del(key string) {
	k0, kn, next := strings.Cut(key, sepNDict)
	if val, ok := d[k0]; ok {
		// not nested key
		if !next {
			delete(d, k0)
			return
		}
		// value is of type Dict
		if v, ok := val.(NDict); ok {
			v.Del(kn)
		}
	}
}

// update dict from updt dict
func (d NDict) Update(updt map[string]any) {
	buff := CreateNDict(updt)
	for _, k := range buff.Keys() {
		d.Set(k, buff.Get(k, nil))
	}
}
