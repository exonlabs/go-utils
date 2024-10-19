// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package mapx

// Find searches for the first key in the map `m` that matches the value `v`.
// It returns the key and a boolean indicating whether the value was found.
//
// - M is a map type where the key is of type T and the value is of type E.
// - T represents the key type, E represents the value type.
func Find[M ~map[T]E, T, E comparable](m M, v E) (T, bool) {
	var r T
	for key, val := range m {
		if val == v {
			return key, true
		}
	}
	return r, false
}
