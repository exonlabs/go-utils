// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package numx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exonlabs/go-utils/pkg/abc/numx"
)

func TestU64(t *testing.T) {
	assert.Equal(t, uint64(0),
		numx.U64([]byte{}), "Empty slice should return 0")
	assert.Equal(t, uint64(0x0102030405060708),
		numx.U64([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}))
	assert.Equal(t, uint64(0x08070605),
		numx.U64([]byte{0x08, 0x07, 0x06, 0x05}))
}

func TestU32(t *testing.T) {
	assert.Equal(t, uint32(0),
		numx.U32([]byte{}), "Empty slice should return 0")
	assert.Equal(t, uint32(0x01020304),
		numx.U32([]byte{0x01, 0x02, 0x03, 0x04}))
	assert.Equal(t, uint32(0x030201),
		numx.U32([]byte{0x03, 0x02, 0x01}))
}

func TestU16(t *testing.T) {
	assert.Equal(t, uint16(0),
		numx.U16([]byte{}), "Empty slice should return 0")
	assert.Equal(t, uint16(0x0102),
		numx.U16([]byte{0x01, 0x02}))
	assert.Equal(t, uint16(0x0201),
		numx.U16([]byte{0x02, 0x01}))
}

func TestU8(t *testing.T) {
	assert.Equal(t, uint8(0),
		numx.U8([]byte{}), "Empty slice should return 0")
	assert.Equal(t, uint8(0x01),
		numx.U8([]byte{0x01}))
	assert.Equal(t, uint8(0xFF),
		numx.U8([]byte{0xFF}))
}

func TestB8(t *testing.T) {
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		numx.B8(0))
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		numx.B8(0x0102030405060708))
}

func TestB4(t *testing.T) {
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00},
		numx.B4(0))
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04},
		numx.B4(0x01020304))
}

func TestB2(t *testing.T) {
	assert.Equal(t, []byte{0x00, 0x00},
		numx.B2(0))
	assert.Equal(t, []byte{0x01, 0x02},
		numx.B2(0x0102))
}

func TestB1(t *testing.T) {
	assert.Equal(t, []byte{0x00},
		numx.B1(0))
	assert.Equal(t, []byte{0xFF},
		numx.B1(0xFF))
}

func TestI64(t *testing.T) {
	assert.Equal(t, int64(0),
		numx.I64([]byte{}), "Empty slice should return 0")
	assert.Equal(t, int64(-1),
		numx.I64([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}),
		"Signed negative")
	assert.Equal(t, int64(0x0102030405060708),
		numx.I64([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}),
		"Positive")
}

func TestI32(t *testing.T) {
	assert.Equal(t, int32(0), numx.I32([]byte{}), "Empty slice should return 0")
	assert.Equal(t, int32(-1),
		numx.I32([]byte{0xFF, 0xFF, 0xFF, 0xFF}), "Signed negative")
	assert.Equal(t, int32(0x01020304),
		numx.I32([]byte{0x01, 0x02, 0x03, 0x04}), "Positive")
}

func TestI16(t *testing.T) {
	assert.Equal(t, int16(0),
		numx.I16([]byte{}), "Empty slice should return 0")
	assert.Equal(t, int16(-1),
		numx.I16([]byte{0xFF, 0xFF}), "Signed negative")
	assert.Equal(t, int16(0x0102),
		numx.I16([]byte{0x01, 0x02}), "Positive")
}

func TestI8(t *testing.T) {
	assert.Equal(t, int8(0),
		numx.I8([]byte{}), "Empty slice should return 0")
	assert.Equal(t, int8(-1),
		numx.I8([]byte{0xFF}), "Signed negative")
	assert.Equal(t, int8(0x01),
		numx.I8([]byte{0x01}), "Positive")
}

func TestQ8(t *testing.T) {
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		numx.Q8(0))
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		numx.Q8(-1), "Signed negative")
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		numx.Q8(0x0102030405060708))
}

func TestQ4(t *testing.T) {
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00},
		numx.Q4(0))
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF},
		numx.Q4(-1), "Signed negative")
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04},
		numx.Q4(0x01020304))
}

func TestQ2(t *testing.T) {
	assert.Equal(t, []byte{0x00, 0x00},
		numx.Q2(0))
	assert.Equal(t, []byte{0xFF, 0xFF},
		numx.Q2(-1), "Signed negative")
	assert.Equal(t, []byte{0x01, 0x02},
		numx.Q2(0x0102))
}

func TestQ1(t *testing.T) {
	assert.Equal(t, []byte{0x00},
		numx.Q1(0))
	assert.Equal(t, []byte{0xFF},
		numx.Q1(-1), "Signed negative")
	assert.Equal(t, []byte{0x01},
		numx.Q1(0x01))
}
