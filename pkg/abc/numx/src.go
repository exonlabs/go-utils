// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package numx

const maxUint64 = 1<<64 - 1

// minNum returns the smaller of two integers a and b.
func minNum(a, b int) int {
	if b < a {
		return b
	}
	return a
}

// U64 converts a big-endian byte slice to a uint64 number.
// It processes up to the first 8 bytes of the slice.
func U64(b []byte) uint64 {
	size := minNum(len(b), 8)
	if size == 0 {
		return 0
	}
	var val uint64
	for i := 0; i < size; i++ {
		val |= uint64(b[i]) << (8 * (size - i - 1))
	}
	return val
}

// U32 converts a big-endian byte slice to a uint32 number.
func U32(b []byte) uint32 {
	return uint32(U64(b[:minNum(len(b), 4)]))
}

// U16 converts a big-endian byte slice to a uint16 number.
func U16(b []byte) uint16 {
	return uint16(U64(b[:minNum(len(b), 2)]))
}

// U8 converts a big-endian byte slice to a uint8 number.
func U8(b []byte) uint8 {
	return uint8(U64(b[:minNum(len(b), 1)]))
}

// B8 converts a uint64 number into a big-endian byte slice of length 8.
func B8(n uint64) []byte {
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[7-i] = byte(n >> (8 * i))
	}
	return b
}

// B4 converts a uint32 number into a big-endian byte slice of length 4.
func B4(n uint32) []byte {
	return B8(uint64(n))[4:]
}

// B2 converts a uint16 number into a big-endian byte slice of length 2.
func B2(n uint16) []byte {
	return B8(uint64(n))[6:]
}

// B1 converts a uint8 number into a single-byte slice.
func B1(n uint8) []byte {
	return B8(uint64(n))[7:]
}

// I64 converts a big-endian byte slice to an int64 number.
// It processes up to the first 8 bytes and handles signed integers.
func I64(b []byte) int64 {
	size := minNum(len(b), 8)
	if size == 0 {
		return 0
	}
	v := U64(b)
	// Check if the number is negative (MSB is 1)
	if b[0]>>7 == 1 {
		if size == 8 {
			// Handle full 64-bit negative numbers
			return -int64(maxUint64 - v + 1)
		}
		// Handle smaller sizes
		return -int64(uint64(1)<<(8*size) - v)
	}
	return int64(v)
}

// I32 converts a big-endian byte slice to an int32 number.
func I32(b []byte) int32 {
	return int32(I64(b[:minNum(len(b), 4)]))
}

// I16 converts a big-endian byte slice to an int16 number.
func I16(b []byte) int16 {
	return int16(I64(b[:minNum(len(b), 2)]))
}

// I8 converts a big-endian byte slice to an int8 number.
func I8(b []byte) int8 {
	return int8(I64(b[:minNum(len(b), 1)]))
}

// Q8 converts an int64 number into a big-endian byte slice of length 8.
// It handles both positive and negative numbers using 2's complement.
func Q8(n int64) []byte {
	var v uint64
	if n >= 0 {
		v = uint64(n)
	} else {
		// Calculate 2's complement for negative numbers
		v = maxUint64 - uint64(-n) + 1
	}
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[7-i] = byte(v >> (8 * i))
	}
	return b
}

// Q4 converts an int32 number into a big-endian byte slice of length 4.
func Q4(n int32) []byte {
	return Q8(int64(n))[4:]
}

// Q2 converts an int16 number into a big-endian byte slice of length 2.
func Q2(n int16) []byte {
	return Q8(int64(n))[6:]
}

// Q1 converts an int8 number into a single-byte slice.
func Q1(n int8) []byte {
	return Q8(int64(n))[7:]
}
