package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"slices"

	"github.com/exonlabs/go-utils/pkg/conv/xnum"
)

// overall status counters
var numSuccess, numTotal = 0, 0

func success_msg() string {
	return fmt.Sprintf("\033[0;32m SUCCESS \033[0m")
}

func fail_msg() string {
	return fmt.Sprintf("\033[0;31m FAIL \033[0m")
}

// Reverse the elements of the slice and return a copy
func revcopy[S ~[]T, T any](s S) S {
	b := make([]T, len(s))
	copy(b, s)
	slices.Reverse(b)
	return b
}

func validate(buffer []byte, signed bool, ltendian bool) bool {
	var n any
	var v []byte
	l := len(buffer)
	if ltendian {
		if signed {
			n = xnum.I64(revcopy(buffer))
			v = revcopy(xnum.Q8(n.(int64)))[:l]
		} else {
			n = xnum.U64(revcopy(buffer))
			v = revcopy(xnum.B8(n.(uint64)))[:l]
		}
	} else {
		if signed {
			n = xnum.I64(buffer)
			v = xnum.Q8(n.(int64))[8-l:]
		} else {
			n = xnum.U64(buffer)
			v = xnum.B8(n.(uint64))[8-l:]
		}
	}
	fmt.Printf("%v %v %v  ---> ",
		hex.EncodeToString(buffer), n, hex.EncodeToString(v))
	return bytes.Equal(v, buffer)
}

func assert_eq(buffer []byte, signed bool, ltendian bool) {
	numTotal += 1
	res, msg := validate(buffer, signed, ltendian), ""
	if res {
		numSuccess += 1
		msg = success_msg()
	} else {
		msg = fail_msg()
	}
	fmt.Printf("%v\n", msg)
}

func assert_neq(buffer []byte, signed bool, ltendian bool) {
	numTotal += 1
	res, msg := validate(buffer, signed, ltendian), ""
	if !res {
		numSuccess += 1
		msg = success_msg()
	} else {
		msg = fail_msg()
	}
	fmt.Printf("%v\n", msg)
}

func main() {
	fmt.Printf("\n\n*** Test Small Values ***\n")
	fmt.Printf("\n-- Unsigned values --\n")
	for k := 1; k <= 8; k++ {
		b := append(bytes.Repeat([]byte{0x00}, k-1), 0x1a)
		assert_eq(b, false, false)
	}
	fmt.Printf("\n-- Signed values --\n")
	for k := 1; k <= 8; k++ {
		b := append(bytes.Repeat([]byte{0x00}, k-1), 0x1a)
		assert_eq(b, true, false)
	}

	fmt.Printf("\n\n*** Test Limit Values ***\n")
	fmt.Printf("\n-- Unsigned values --\n")
	for k := 1; k <= 8; k++ {
		b := bytes.Repeat([]byte{0xff}, k)
		assert_eq(b, false, false)
	}
	fmt.Printf("\n-- Signed values --\n")
	for k := 1; k <= 8; k++ {
		b := bytes.Repeat([]byte{0xff}, k)
		assert_eq(b, true, false)
	}

	fmt.Printf("\n\n*** Test MSB Zero byte Values ***\n")
	fmt.Printf("\n-- Unsigned values --\n")
	for k := 1; k <= 7; k++ {
		b := append([]byte{0x00}, bytes.Repeat([]byte{0xff}, k)...)
		assert_eq(b, false, false)
	}
	fmt.Printf("\n-- Signed values --\n")
	for k := 1; k <= 7; k++ {
		b := append([]byte{0x00}, bytes.Repeat([]byte{0xff}, k)...)
		assert_eq(b, true, false)
	}

	fmt.Printf("\n\n*** Test MSB 0xFF byte Values ***\n")
	fmt.Printf("\n-- Unsigned values --\n")
	for k := 1; k <= 7; k++ {
		b := append([]byte{0xff}, bytes.Repeat([]byte{0x10}, k)...)
		assert_eq(b, false, false)
	}
	fmt.Printf("\n-- Signed values --\n")
	for k := 1; k <= 7; k++ {
		b := append([]byte{0xff}, bytes.Repeat([]byte{0x10}, k)...)
		assert_eq(b, true, false)
	}

	fmt.Printf("\n\n*** Test little endian Values ***\n")
	fmt.Printf("\n-- Unsigned values --\n")
	for k := 1; k <= 7; k++ {
		b := append(bytes.Repeat([]byte{0x10}, k), 0xff)
		assert_eq(b, false, true)
	}
	fmt.Printf("\n-- Signed values --\n")
	for k := 1; k <= 7; k++ {
		b := append(bytes.Repeat([]byte{0x10}, k), 0xff)
		assert_eq(b, true, true)
	}

	fmt.Printf("\n\n*** Result ***\n")
	fmt.Printf("\nSuccess: %v / %v\n", numSuccess, numTotal)

	fmt.Printf("\n\n")
}
