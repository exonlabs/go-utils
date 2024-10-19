// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package numx_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/exonlabs/go-utils/pkg/abc/numx"
	"github.com/exonlabs/go-utils/pkg/abc/slicex"
)

// convert byte buffer to numeric then back to bytes
func num_convert(b []byte, signed bool, ltendian bool) (any, []byte) {
	var n any
	var v []byte
	l := len(b)
	if ltendian {
		if signed {
			n = numx.I64(slicex.ReverseCopy(b))
			v = slicex.ReverseCopy(numx.Q8(n.(int64)))[:l]
		} else {
			n = numx.U64(slicex.ReverseCopy(b))
			v = slicex.ReverseCopy(numx.B8(n.(uint64)))[:l]
		}
	} else {
		if signed {
			n = numx.I64(b)
			v = numx.Q8(n.(int64))[8-l:]
		} else {
			n = numx.U64(b)
			v = numx.B8(n.(uint64))[8-l:]
		}
	}
	return n, v
}

func TestSmallValues_Unsigned(t *testing.T) {
	for k := 1; k <= 8; k++ {
		b_in := append(bytes.Repeat([]byte{0x00}, k-1), 0x1a)
		num, b_out := num_convert(b_in, false, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestSmallValues_Signed(t *testing.T) {
	for k := 1; k <= 8; k++ {
		b_in := append(bytes.Repeat([]byte{0x00}, k-1), 0x1a)
		num, b_out := num_convert(b_in, true, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestLimitValues_Unsigned(t *testing.T) {
	for k := 1; k <= 8; k++ {
		b_in := append(bytes.Repeat([]byte{0xff}, k-1), 0x1a)
		num, b_out := num_convert(b_in, false, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestLimitValues_Signed(t *testing.T) {
	for k := 1; k <= 8; k++ {
		b_in := append(bytes.Repeat([]byte{0xff}, k-1), 0x1a)
		num, b_out := num_convert(b_in, true, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestMSB00_Unsigned(t *testing.T) {
	for k := 1; k <= 7; k++ {
		b_in := append([]byte{0x00}, bytes.Repeat([]byte{0xff}, k)...)
		num, b_out := num_convert(b_in, false, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestMSB00_Signed(t *testing.T) {
	for k := 1; k <= 7; k++ {
		b_in := append([]byte{0x00}, bytes.Repeat([]byte{0xff}, k)...)
		num, b_out := num_convert(b_in, true, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestMSBFF_Unsigned(t *testing.T) {
	for k := 1; k <= 7; k++ {
		b_in := append([]byte{0xff}, bytes.Repeat([]byte{0x10}, k)...)
		num, b_out := num_convert(b_in, false, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestMSBFF_Signed(t *testing.T) {
	for k := 1; k <= 7; k++ {
		b_in := append([]byte{0xff}, bytes.Repeat([]byte{0x10}, k)...)
		num, b_out := num_convert(b_in, true, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestLittleEndianValues_Unsigned(t *testing.T) {
	for k := 1; k <= 7; k++ {
		b_in := append(bytes.Repeat([]byte{0x10}, k), 0xff)
		num, b_out := num_convert(b_in, false, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}

func TestLittleEndianValues_Signed(t *testing.T) {
	for k := 1; k <= 7; k++ {
		b_in := append(bytes.Repeat([]byte{0x10}, k), 0xff)
		num, b_out := num_convert(b_in, true, false)
		t.Logf("input: %v ---> %v %v\n",
			hex.EncodeToString(b_in), num, hex.EncodeToString(b_out))
		if bytes.Equal(b_in, b_out) {
			t.Logf("VALID")
		} else {
			t.Errorf("FAILED")
		}
	}
}
