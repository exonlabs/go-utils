// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package numx_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/numx"
)

func ExampleU64() {
	b := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	fmt.Println(numx.U64(b))
	// Output: 72623859790382856
}

func ExampleB8() {
	n := uint64(72623859790382856)
	fmt.Printf("%x\n", numx.B8(n))
	// Output: 0102030405060708
}

func ExampleU32() {
	b := []byte{0x01, 0x02, 0x03, 0x04}
	fmt.Println(numx.U32(b))
	// Output: 16909060
}

func ExampleB4() {
	n := uint32(16909060)
	fmt.Printf("%x\n", numx.B4(n))
	// Output: 01020304
}

func ExampleI64() {
	b := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	fmt.Println(numx.I64(b)) // Signed negative
	// Output: -1
}

func ExampleQ8() {
	n := int64(-1)
	fmt.Printf("%x\n", numx.Q8(n))
	// Output: ffffffffffffffff
}
