package types

import (
	"bytes"
	"encoding/gob"
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
				switch v.(type) {
				case map[string]any, Dict, NDict:
					b = append(b, NewDict(v.(map[string]any)))
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
	if v, ok := d.Get(key, defval).(Dict); ok {
		return v
	}
	return defval
}
func (d Dict) GetNDict(key string, defval NDict) NDict {
	if v, ok := d.Get(key, defval).(NDict); ok {
		return v
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
func (d Dict) GetByte(key string, defval byte) byte {
	if v, ok := d[key].(byte); ok {
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

func (d Dict) GetDictSlice(key string, defval []Dict) []Dict {
	if v, ok := d.Get(key, defval).([]Dict); ok {
		return v
	}
	return defval
}
func (d Dict) GetNDictSlice(key string, defval []NDict) []NDict {
	if v, ok := d.Get(key, defval).([]NDict); ok {
		return v
	}
	return defval
}
func (d Dict) GetBoolSlice(key string, defval []bool) []bool {
	if v, ok := d[key].([]bool); ok {
		return v
	}
	return defval
}
func (d Dict) GetStringSlice(key string, defval []string) []string {
	if v, ok := d[key].([]string); ok {
		return v
	}
	return defval
}
func (d Dict) GetByteSlice(key string, defval []byte) []byte {
	if v, ok := d[key].([]byte); ok {
		return v
	}
	return defval
}
func (d Dict) GetIntSlice(key string, defval []int) []int {
	if v, ok := d[key].([]int); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt8Slice(key string, defval []int8) []int8 {
	if v, ok := d[key].([]int8); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt16Slice(key string, defval []int16) []int16 {
	if v, ok := d[key].([]int16); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt32Slice(key string, defval []int32) []int32 {
	if v, ok := d[key].([]int32); ok {
		return v
	}
	return defval
}
func (d Dict) GetInt64Slice(key string, defval []int64) []int64 {
	if v, ok := d[key].([]int64); ok {
		return v
	}
	return defval
}
func (d Dict) GetUintSlice(key string, defval []uint) []uint {
	if v, ok := d[key].([]uint); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint8Slice(key string, defval []uint8) []uint8 {
	if v, ok := d[key].([]uint8); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint16Slice(key string, defval []uint16) []uint16 {
	if v, ok := d[key].([]uint16); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint32Slice(key string, defval []uint32) []uint32 {
	if v, ok := d[key].([]uint32); ok {
		return v
	}
	return defval
}
func (d Dict) GetUint64Slice(key string, defval []uint64) []uint64 {
	if v, ok := d[key].([]uint64); ok {
		return v
	}
	return defval
}
func (d Dict) GetFloat32Slice(key string, defval []float32) []float32 {
	if v, ok := d[key].([]float32); ok {
		return v
	}
	return defval
}
func (d Dict) GetFloat64Slice(key string, defval []float64) []float64 {
	if v, ok := d[key].([]float64); ok {
		return v
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
