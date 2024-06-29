package types

import (
	"bytes"
	"encoding/gob"
	"reflect"
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
func NewNDict(buff map[string]any) NDict {
	if buff == nil {
		return NDict{}
	}
	// delete empty keys
	delete(buff, "")
	// recursive conversion
	for key, value := range buff {
		switch val := value.(type) {
		case map[string]any:
			buff[key] = NewNDict(val)
		case Dict:
			buff[key] = NewNDict(val)
		case NDict:
			buff[key] = NewNDict(val)
		case []map[string]any:
			b := []NDict{}
			for _, v := range val {
				b = append(b, NewNDict(v))
			}
			buff[key] = b
		case []Dict:
			b := []NDict{}
			for _, v := range val {
				b = append(b, NewNDict(v))
			}
			buff[key] = b
		case []NDict:
			b := []NDict{}
			for _, v := range val {
				b = append(b, NewNDict(v))
			}
			buff[key] = b
		case []any:
			b := []any{}
			for _, v := range val {
				switch d := v.(type) {
				case map[string]any:
					b = append(b, NewNDict(d))
				case Dict:
					b = append(b, NewNDict(d))
				case NDict:
					b = append(b, NewNDict(d))
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

// create deep clone of NDict
func CloneNDict(src map[string]any) (NDict, error) {
	gob.Register(Dict{})
	gob.Register([]Dict{})
	gob.Register(NDict{})
	gob.Register([]NDict{})
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(src); err != nil {
		return nil, err
	}
	var dst Dict
	if err := gob.NewDecoder(&b).Decode(&dst); err != nil {
		return nil, err
	}
	return NewNDict(dst), nil
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
func (d NDict) IsExist(key string) bool {
	k0, kn, next := strings.Cut(key, sepNDict)
	if val, ok := d[k0]; ok {
		// not nested key
		if !next {
			return true
		}
		// value is of type Dict
		if v, ok := val.(NDict); ok {
			return v.IsExist(kn)
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

func (d NDict) GetDict(key string, defval Dict) Dict {
	switch val := d.Get(key, defval).(type) {
	case Dict:
		return val
	case NDict:
		return NewDict(val)
	case map[string]any:
		return NewDict(val)
	}
	return defval
}
func (d NDict) GetNDict(key string, defval NDict) NDict {
	switch val := d.Get(key, defval).(type) {
	case NDict:
		return val
	case Dict:
		return NewNDict(val)
	case map[string]any:
		return NewNDict(val)
	}
	return defval
}
func (d NDict) GetBool(key string, defval bool) bool {
	switch val := d.Get(key, defval).(type) {
	case bool:
		return val
	}
	return defval
}
func (d NDict) GetByte(key string, defval byte) byte {
	switch val := d.Get(key, defval).(type) {
	case byte:
		return val
	}
	return defval
}
func (d NDict) GetBytes(key string, defval []byte) []byte {
	switch val := d.Get(key, defval).(type) {
	case []byte:
		return val
	case string:
		return []byte(val)
	case rune:
		return []byte(string(val))
	}
	return defval
}
func (d NDict) GetString(key string, defval string) string {
	switch val := d.Get(key, defval).(type) {
	case string:
		return val
	case rune:
		return string(val)
	case []byte:
		return string(val)
	}
	return defval
}
func (d NDict) GetRune(key string, defval rune) rune {
	switch val := d.Get(key, defval).(type) {
	case rune:
		return val
	}
	return defval
}
func (d NDict) GetFloat64(key string, defval float64) float64 {
	switch val := d.Get(key, defval).(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case int16:
		return float64(val)
	case int8:
		return float64(val)
	case int:
		return float64(val)
	case uint64:
		return float64(val)
	case uint32:
		return float64(val)
	case uint16:
		return float64(val)
	case uint8:
		return float64(val)
	case uint:
		return float64(val)
	}
	return defval
}
func (d NDict) GetFloat32(key string, defval float32) float32 {
	return float32(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetInt64(key string, defval int64) int64 {
	return int64(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetInt32(key string, defval int32) int32 {
	return int32(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetInt16(key string, defval int16) int16 {
	return int16(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetInt8(key string, defval int8) int8 {
	return int8(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetInt(key string, defval int) int {
	return int(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetUint64(key string, defval uint64) uint64 {
	return uint64(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetUint32(key string, defval uint32) uint32 {
	return uint32(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetUint16(key string, defval uint16) uint16 {
	return uint16(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetUint8(key string, defval uint8) uint8 {
	return uint8(d.GetFloat64(key, float64(defval)))
}
func (d NDict) GetUint(key string, defval uint) uint {
	return uint(d.GetFloat64(key, float64(defval)))
}

func (d NDict) GetSlice(key string, defval []any) []any {
	val := d.Get(key, defval)
	if reflect.TypeOf(val).Kind() == reflect.Slice {
		return val.([]any)
	}
	return defval
}
func (d NDict) GetDictSlice(key string, defval []Dict) []Dict {
	res := []Dict{}
	switch val := d.Get(key, defval).(type) {
	case []Dict:
		return val
	case []NDict:
		for _, v := range val {
			res = append(res, NewDict(v))
		}
	case []map[string]any:
		for _, v := range val {
			res = append(res, NewDict(v))
		}
	case []any:
	R:
		for _, v := range val {
			switch d := v.(type) {
			case map[string]any:
				res = append(res, NewDict(d))
			case Dict:
				res = append(res, NewDict(d))
			case NDict:
				res = append(res, NewDict(d))
			default:
				res = nil
				break R
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetNDictSlice(key string, defval []NDict) []NDict {
	res := []NDict{}
	switch val := d.Get(key, defval).(type) {
	case []NDict:
		return val
	case []Dict:
		for _, v := range val {
			res = append(res, NewNDict(v))
		}
	case []map[string]any:
		for _, v := range val {
			res = append(res, NewNDict(v))
		}
	case []any:
	R:
		for _, v := range val {
			switch d := v.(type) {
			case map[string]any:
				res = append(res, NewNDict(d))
			case Dict:
				res = append(res, NewNDict(d))
			case NDict:
				res = append(res, NewNDict(d))
			default:
				res = nil
				break R
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetBoolSlice(key string, defval []bool) []bool {
	res := []bool{}
	switch val := d.Get(key, defval).(type) {
	case []bool:
		return val
	case []any:
	R:
		for _, v := range val {
			switch d := v.(type) {
			case bool:
				res = append(res, bool(d))
			default:
				res = nil
				break R
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetStringSlice(key string, defval []string) []string {
	res := []string{}
	switch val := d.Get(key, defval).(type) {
	case []string:
		return val
	case []rune:
		for _, v := range val {
			res = append(res, string(v))
		}
	case []any:
	R:
		for _, v := range val {
			switch d := v.(type) {
			case string:
				res = append(res, string(d))
			case rune:
				res = append(res, string(d))
			default:
				res = nil
				break R
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetRuneSlice(key string, defval []rune) []rune {
	res := []rune{}
	switch val := d.Get(key, defval).(type) {
	case []rune:
		return val
	case []any:
	R:
		for _, v := range val {
			switch d := v.(type) {
			case rune:
				res = append(res, rune(d))
			default:
				res = nil
				break R
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetFloat64Slice(key string, defval []float64) []float64 {
	res := []float64{}
	switch val := d.Get(key, defval).(type) {
	case []float64:
		return val
	case []float32:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []int64:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []int32:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []int16:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []int8:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []int:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []uint64:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []uint32:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []uint16:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []uint8:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []uint:
		for _, v := range val {
			res = append(res, float64(v))
		}
	case []any:
	R:
		for _, v := range val {
			switch n := v.(type) {
			case float64:
				res = append(res, float64(n))
			case float32:
				res = append(res, float64(n))
			case int64:
				res = append(res, float64(n))
			case int32:
				res = append(res, float64(n))
			case int16:
				res = append(res, float64(n))
			case int8:
				res = append(res, float64(n))
			case int:
				res = append(res, float64(n))
			case uint64:
				res = append(res, float64(n))
			case uint32:
				res = append(res, float64(n))
			case uint16:
				res = append(res, float64(n))
			case uint8:
				res = append(res, float64(n))
			case uint:
				res = append(res, float64(n))
			default:
				res = nil
				break R
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetFloat32Slice(key string, defval []float32) []float32 {
	res := []float32{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, float32(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetInt64Slice(key string, defval []int64) []int64 {
	res := []int64{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int64(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetInt32Slice(key string, defval []int32) []int32 {
	res := []int32{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int32(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetInt16Slice(key string, defval []int16) []int16 {
	res := []int16{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int16(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetInt8Slice(key string, defval []int8) []int8 {
	res := []int8{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int8(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetIntSlice(key string, defval []int) []int {
	res := []int{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetUint64Slice(key string, defval []uint64) []uint64 {
	res := []uint64{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint64(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetUint32Slice(key string, defval []uint32) []uint32 {
	res := []uint32{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint32(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetUint16Slice(key string, defval []uint16) []uint16 {
	res := []uint16{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint16(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetUint8Slice(key string, defval []uint8) []uint8 {
	res := []uint8{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint8(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d NDict) GetUintSlice(key string, defval []uint) []uint {
	res := []uint{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}

// set value in dict by key
func (d NDict) Set(key string, newval any) {
	if len(key) != 0 {
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
}

// delete value from dict by key
func (d NDict) Del(key string) {
	if len(key) != 0 {
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
}

// update dict from updt dict
func (d NDict) Update(updt map[string]any) {
	buff := NewNDict(updt)
	for _, key := range buff.Keys() {
		if len(key) != 0 {
			d.Set(key, buff.Get(key, nil))
		}
	}
}

// delete all values from dict
func (d NDict) Reset() {
	for k := range d {
		delete(d, k)
	}
}
