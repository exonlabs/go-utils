// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package dictx

import (
	"strings"
)

// Key separator character used for nested keys
const Separator = "."

// Dict type representation as a map with string keys and any values
type Dict = map[string]any

// Clone creates a deep copy of a Dict.
// It returns a new dictionary that is a copy of the original,
// preserving the structure and values.
func Clone(d Dict) (Dict, error) {
	newDict := make(Dict, len(d))
	for k, v := range d {
		if nestedDict, ok := v.(Dict); ok {
			clonedNestedDict, err := Clone(nestedDict)
			if err != nil {
				return nil, err
			}
			newDict[k] = clonedNestedDict
		} else {
			newDict[k] = v
		}
	}
	return newDict, nil
}

// KeysN returns a list of keys up to N levels nested.
// If n is 1, only top-level keys are returned.
// If n is greater than 1, it retrieves nested keys accordingly.
// Zero-length keys are omitted from the results.
func KeysN(d Dict, n int) []string {
	if len(d) == 0 {
		return nil
	}
	keys := make([]string, 0, len(d))
	for k, v := range d {
		if len(k) > 0 {
			if n != 1 {
				if nestedDict, ok := v.(Dict); ok {
					for _, sk := range KeysN(nestedDict, n-1) {
						keys = append(keys, k+Separator+sk)
					}
					continue
				}
			}
			keys = append(keys, k)
		}
	}
	return keys
}

// Keys returns a list of all keys in the dictionary,
// regardless of nesting levels. It omits zero-length keys.
func Keys(d Dict) []string {
	return KeysN(d, -1)
}

// IsExist checks if a key exists in the dictionary.
// It supports nested keys using the separator.
// Returns true if the key exists, false otherwise.
func IsExist(d Dict, key string) bool {
	if len(key) == 0 {
		return false
	}
	keys := strings.Split(key, Separator)
	current := d
	for _, k := range keys {
		val, ok := current[k]
		if !ok {
			return false
		}
		if nestedDict, ok := val.(Dict); ok {
			current = nestedDict
		} else if k != keys[len(keys)-1] {
			// We reached a non-Dict value before the last key
			return false
		}
	}
	return true
}

// Get retrieves a value from the dictionary by key.
// If the key is not found, the default_value is returned.
func Get(d Dict, key string, defaultValue any) any {
	if len(key) == 0 {
		return defaultValue
	}
	keys := strings.Split(key, Separator)
	current := d
	for i, k := range keys {
		val, ok := current[k]
		if !ok {
			return defaultValue
		}
		if i == len(keys)-1 {
			return val
		} else if nestedDict, ok := val.(Dict); ok {
			current = nestedDict
		} else {
			return defaultValue
		}
	}
	return defaultValue
}

// Fetch retrieves a value from the dictionary by key with type casting conversion.
// If the key is not found, the default_value is returned.
func Fetch[T any](d Dict, key string, defaultValue T) T {
	val := Get(d, key, nil)
	if v, ok := val.(T); ok {
		return v
	}
	return defaultValue
}

// Set adds a new value in the dictionary by key.
// If the key already exists, its value is overwritten.
func Set(d Dict, key string, newValue any) {
	if len(key) == 0 {
		return
	}
	keys := strings.Split(key, Separator)
	current := d
	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = newValue
			return
		}
		// If not a Dict, create new nested Dict
		if nestedDict, ok := current[k].(Dict); ok {
			current = nestedDict
		} else {
			newDict := Dict{}
			current[k] = newDict
			current = newDict
		}
	}
}

// Merge updates a source dictionary recursively with an update dictionary.
// It merges keys and values, allowing nested dictionaries to be updated as well.
func Merge(src, updt Dict) {
	for k, v := range updt {
		if vDict, ok := v.(Dict); ok {
			if srcDict, ok := src[k].(Dict); ok {
				Merge(srcDict, vDict)
			} else {
				src[k] = vDict // Copy the nested Dict
			}
		} else {
			src[k] = v // Non-Dict value is overwritten
		}
	}
}

// Delete removes a key from the dictionary if it exists.
// It supports nested keys using the separator.
func Delete(d Dict, key string) {
	if len(key) == 0 {
		return
	}
	keys := strings.Split(key, Separator)
	current := d
	for i, k := range keys {
		if i == len(keys)-1 {
			delete(current, k)
			return
		}
		if nestedDict, ok := current[k].(Dict); ok {
			current = nestedDict
		} else {
			return
		}
	}
}
