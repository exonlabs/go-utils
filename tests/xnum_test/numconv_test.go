package xnum_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/exonlabs/go-utils/pkg/conv/xnum"
	"github.com/exonlabs/go-utils/tests"
)

// convert byte buffer to numeric then back to bytes
func num_convert(b []byte, signed bool, ltendian bool) (any, []byte) {
	var n any
	var v []byte
	l := len(b)
	if ltendian {
		if signed {
			n = xnum.I64(tests.RevCopy(b))
			v = tests.RevCopy(xnum.Q8(n.(int64)))[:l]
		} else {
			n = xnum.U64(tests.RevCopy(b))
			v = tests.RevCopy(xnum.B8(n.(uint64)))[:l]
		}
	} else {
		if signed {
			n = xnum.I64(b)
			v = xnum.Q8(n.(int64))[8-l:]
		} else {
			n = xnum.U64(b)
			v = xnum.B8(n.(uint64))[8-l:]
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
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
			t.Logf(tests.ValidMsg())
		} else {
			t.Errorf(tests.FailMsg())
		}
	}
}
