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
