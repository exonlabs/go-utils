// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xdict

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strings"
)

// Key seperator char
const Seperator = "."

// Dict type representation
type Dict = map[string]any

// Clone creates a deep copy of a Dict
func Clone(d Dict) (Dict, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(d); err != nil {
		return nil, err
	}
	var v Dict
	if err := gob.NewDecoder(&b).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// KeysN returns a sorted list of keys up to N levels nested keys.
// zero len keys are omitted.
func KeysN(d Dict, n int) []string {
	keys := []string{}
	for k := range d {
		if len(k) > 0 {
			if n != 1 {
				if v, ok := d[k].(Dict); ok {
					for _, sk := range KeysN(v, n-1) {
						keys = append(keys, k+Seperator+sk)
					}
					continue
				}
			}
			keys = append(keys, k)
		}
	}
	if len(keys) > 0 {
		sort.Strings(keys)
	}
	return keys
}

// Keys returns a sorted list of all nested level keys.
// zero len keys are omitted.
func Keys(d Dict) []string {
	return KeysN(d, -1)
}

// IsExist checks if a key exist in dict
func IsExist(d Dict, key string) bool {
	if len(key) == 0 {
		return false
	}
	k0, kn, next := strings.Cut(key, Seperator)
	if val, ok := d[k0]; ok {
		if next { // we have nested key
			if v, ok := val.(Dict); ok {
				return IsExist(v, kn)
			}
		} else {
			return true
		}
	}
	return false
}

// Get returns value from dict by key.
// if key is not found then default_value is returned.
func Get(d Dict, key string, default_value any) any {
	if len(key) == 0 {
		return default_value
	}
	k0, kn, next := strings.Cut(key, Seperator)
	if val, ok := d[k0]; ok {
		if next { // we have nested key
			if v, ok := val.(Dict); ok {
				return Get(v, kn, default_value)
			}
		} else {
			return val
		}
	}
	return default_value
}

// Fetch returns value from dict by key with type casting conversion.
// if key is not found then default_value is returned.
func Fetch[T any | ~[]any](d Dict, key string, default_value T) T {
	val := Get(d, key, nil)
	if v, ok := val.(T); ok {
		return v
	}
	return default_value
}

// Set addes new_value in dict by key.
// if a key already exists its value is overwritten.
func Set(d Dict, key string, new_value any) {
	if len(key) == 0 {
		return
	}
	k0, kn, next := strings.Cut(key, Seperator)
	if next { // we have nested key
		// 1st level key not exist or not of type Dict
		if _, ok := d[k0].(Dict); !ok {
			d[k0] = Dict{}
		}
		Set(d[k0].(Dict), kn, new_value)
	} else {
		d[k0] = new_value
	}
}

// Merge updates a src dict recursively by an updt dict.
func Merge(src, updt Dict) {
	for _, key := range Keys(updt) {
		Set(src, key, Get(updt, key, nil))
	}
}

// Delete removes a key from dict if exists.
func Delete(d Dict, key string) {
	if len(key) == 0 {
		return
	}
	k0, kn, next := strings.Cut(key, Seperator)
	if val, ok := d[k0]; ok {
		if next { // we have nested key
			if v, ok := val.(Dict); ok {
				Delete(v, kn)
			}
		} else {
			delete(d, k0)
		}
	}
}
