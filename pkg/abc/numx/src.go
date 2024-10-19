// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package numx

const maxUint64 = 1<<64 - 1

// return minimum number of a and b
func _min(a, b int) int {
	if b < a {
		return b
	}
	return a
}

// Convert big-endian bytes into uint64 number.
func U64(b []byte) uint64 {
	size := _min(len(b), 8)
	if size == 0 {
		return 0
	}
	val := uint64(0)
	for i := 0; i < size; i++ {
		val += uint64(b[i]) << (8 * (size - i - 1))
	}
	return val
}

// Convert big-endian bytes into uint32 number.
func U32(b []byte) uint32 {
	return uint32(U64(b[:_min(len(b), 4)]))
}

// Convert big-endian bytes into uint16 number.
func U16(b []byte) uint16 {
	return uint16(U64(b[:_min(len(b), 2)]))
}

// Convert big-endian bytes into uint8 number.
func U8(b []byte) uint8 {
	return uint8(U64(b[:_min(len(b), 1)]))
}

// Convert uint64 number n into big-endian bytes format.
func B8(n uint64) []byte {
	// create byte b and fill in reverse order for big-endian
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[7-i] = byte(n >> (8 * i))
	}
	return b
}
func B4(val uint32) []byte {
	return B8(uint64(val))[4:]
}
func B2(val uint16) []byte {
	return B8(uint64(val))[6:]
}
func B1(val uint8) []byte {
	return B8(uint64(val))[7:]
}

// convert big-endian bytes buffer into int
func I64(buffer []byte) int64 {
	size := _min(len(buffer), 8)
	if size == 0 {
		return 0
	}
	// parse bytes as uint64
	v := U64(buffer)
	// check negative number by MSB, if 1 then -ve number,
	// get 2's complement and add -ve sign
	if buffer[0]>>7 == 0x01 {
		if size == 8 {
			return -int64(maxUint64 - v + 1)
		}
		return -int64(uint64(0x01)<<(8*size) - v)
	}
	return int64(v)
}
func I32(buffer []byte) int32 {
	return int32(I64(buffer[:_min(len(buffer), 4)]))
}
func I16(buffer []byte) int16 {
	return int16(I64(buffer[:_min(len(buffer), 2)]))
}
func I8(buffer []byte) int8 {
	return int8(I64(buffer[:_min(len(buffer), 1)]))
}

// convert int into big-endian bytes buffer
func Q8(val int64) []byte {
	var v uint64
	if val >= 0 {
		// +ve numbers
		v = uint64(val)
	} else {
		// -ve numbers, get 2's complement
		v = maxUint64 - uint64(-val) + 1
	}
	// create byte buffer and fill in reverse order for big-endian
	buffer := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buffer[7-i] = byte(v >> (8 * i))
	}
	return buffer
}
func Q4(val int32) []byte {
	return Q8(int64(val))[4:]
}
func Q2(val int16) []byte {
	return Q8(int64(val))[6:]
}
func Q1(val int8) []byte {
	return Q8(int64(val))[7:]
}
