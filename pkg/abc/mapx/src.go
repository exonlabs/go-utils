// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package mapx

// Find the key of the first match value v in map m.
// returns the key if value is found and bool status indication.
func Find[M ~map[T]E, T, E comparable](m M, v E) (T, bool) {
	var r T
	for key, val := range m {
		if val == v {
			return key, true
		}
	}
	return r, false
}
