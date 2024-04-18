package types

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"slices"
)

// Simple Dict type
type Dict map[string]any

// create new Dict from initial map data
func NewDict(buff map[string]any) Dict {
	if buff == nil {
		return Dict{}
	}
	// delete empty keys
	delete(buff, "")
	// recursive conversion
	for key, value := range buff {
		switch val := value.(type) {
		case map[string]any:
			buff[key] = NewDict(val)
		case Dict:
			buff[key] = NewDict(val)
		case NDict:
			buff[key] = NewDict(val)
		case []map[string]any:
			b := []Dict{}
			for _, v := range val {
				b = append(b, NewDict(v))
			}
			buff[key] = b
		case []Dict:
			b := []Dict{}
			for _, v := range val {
				b = append(b, NewDict(v))
			}
			buff[key] = b
		case []NDict:
			b := []Dict{}
			for _, v := range val {
				b = append(b, NewDict(v))
			}
			buff[key] = b
		case []any:
			b := []any{}
			for _, v := range val {
				switch d := v.(type) {
				case map[string]any:
					b = append(b, NewDict(d))
				case Dict:
					b = append(b, NewDict(d))
				case NDict:
					b = append(b, NewDict(d))
				default:
					b = append(b, v)
				}
			}
			buff[key] = b
		}
	}
	return Dict(buff)
}

// recursive convert Dict into standard map data
func StripDict(buff map[string]any) map[string]any {
	if buff == nil {
		return map[string]any{}
	}
	for key, value := range buff {
		switch val := value.(type) {
		case map[string]any:
			buff[key] = StripDict(val)
		case Dict:
			buff[key] = StripDict(val)
		case NDict:
			buff[key] = StripDict(val)
		case []map[string]any:
			b := []map[string]any{}
			for _, v := range val {
				b = append(b, StripDict(v))
			}
			buff[key] = b
		case []Dict:
			b := []map[string]any{}
			for _, v := range val {
				b = append(b, StripDict(v))
			}
			buff[key] = b
		case []NDict:
			b := []map[string]any{}
			for _, v := range val {
				b = append(b, StripDict(v))
			}
			buff[key] = b
		case []any:
			b := []any{}
			for _, sv := range val {
				switch v := sv.(type) {
				case map[string]any:
					b = append(b, StripDict(v))
				case Dict:
					b = append(b, StripDict(v))
				case NDict:
					b = append(b, StripDict(v))
				default:
					b = append(b, v)
				}
			}
			buff[key] = b
		}
	}
	return buff
}

// create deep clone of Dict
func CloneDict(src map[string]any) (Dict, error) {
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
	return NewDict(dst), nil
}

// return sorted list of all keys
func (d Dict) Keys() []string {
	keys := []string{}
	for k := range d {
		keys = append(keys, k)
	}
	if len(keys) > 0 {
		slices.Sort(keys)
	}
	return keys
}

// check if key exist in dict
func (d Dict) IsExist(key string) bool {
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

func (d Dict) GetDict(key string, defval Dict) Dict {
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
func (d Dict) GetNDict(key string, defval NDict) NDict {
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
func (d Dict) GetBool(key string, defval bool) bool {
	switch val := d.Get(key, defval).(type) {
	case bool:
		return val
	}
	return defval
}
func (d Dict) GetByte(key string, defval byte) byte {
	switch val := d.Get(key, defval).(type) {
	case byte:
		return val
	}
	return defval
}
func (d Dict) GetBytes(key string, defval []byte) []byte {
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
func (d Dict) GetString(key string, defval string) string {
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
func (d Dict) GetRune(key string, defval rune) rune {
	switch val := d.Get(key, defval).(type) {
	case rune:
		return val
	}
	return defval
}
func (d Dict) GetFloat64(key string, defval float64) float64 {
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
func (d Dict) GetFloat32(key string, defval float32) float32 {
	return float32(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetInt64(key string, defval int64) int64 {
	return int64(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetInt32(key string, defval int32) int32 {
	return int32(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetInt16(key string, defval int16) int16 {
	return int16(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetInt8(key string, defval int8) int8 {
	return int8(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetInt(key string, defval int) int {
	return int(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetUint64(key string, defval uint64) uint64 {
	return uint64(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetUint32(key string, defval uint32) uint32 {
	return uint32(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetUint16(key string, defval uint16) uint16 {
	return uint16(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetUint8(key string, defval uint8) uint8 {
	return uint8(d.GetFloat64(key, float64(defval)))
}
func (d Dict) GetUint(key string, defval uint) uint {
	return uint(d.GetFloat64(key, float64(defval)))
}

func (d Dict) GetSlice(key string, defval []any) []any {
	val := d.Get(key, defval)
	if reflect.TypeOf(val).Kind() == reflect.Slice {
		return val.([]any)
	}
	return defval
}
func (d Dict) GetDictSlice(key string, defval []Dict) []Dict {
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
func (d Dict) GetNDictSlice(key string, defval []NDict) []NDict {
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
func (d Dict) GetBoolSlice(key string, defval []bool) []bool {
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
func (d Dict) GetStringSlice(key string, defval []string) []string {
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
func (d Dict) GetRuneSlice(key string, defval []rune) []rune {
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
func (d Dict) GetFloat64Slice(key string, defval []float64) []float64 {
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
func (d Dict) GetFloat32Slice(key string, defval []float32) []float32 {
	res := []float32{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, float32(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetInt64Slice(key string, defval []int64) []int64 {
	res := []int64{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int64(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetInt32Slice(key string, defval []int32) []int32 {
	res := []int32{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int32(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetInt16Slice(key string, defval []int16) []int16 {
	res := []int16{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int16(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetInt8Slice(key string, defval []int8) []int8 {
	res := []int8{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int8(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetIntSlice(key string, defval []int) []int {
	res := []int{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, int(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetUint64Slice(key string, defval []uint64) []uint64 {
	res := []uint64{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint64(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetUint32Slice(key string, defval []uint32) []uint32 {
	res := []uint32{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint32(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetUint16Slice(key string, defval []uint16) []uint16 {
	res := []uint16{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint16(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetUint8Slice(key string, defval []uint8) []uint8 {
	res := []uint8{}
	for _, v := range d.GetFloat64Slice(key, nil) {
		res = append(res, uint8(v))
	}
	if len(res) > 0 {
		return res
	}
	return defval
}
func (d Dict) GetUintSlice(key string, defval []uint) []uint {
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
func (d Dict) Set(key string, newval any) {
	if len(key) != 0 {
		d[key] = newval
	}
}

// delete value from dict by key
func (d Dict) Del(key string) {
	if len(key) != 0 {
		delete(d, key)
	}
}

// update dict from map data
func (d Dict) Update(updt map[string]any) {
	for key, val := range updt {
		if len(key) != 0 {
			d[key] = val
		}
	}
}
