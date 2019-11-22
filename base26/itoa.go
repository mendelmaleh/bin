// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base26

// FormatUint returns the string representation of i in base 26.
// The result uses the lower-case letters 'a' to 'z'
func FormatUint(i uint64) string {
	_, s := formatBits(nil, i, false, false)
	return s
}

// FormatInt returns the string representation of i in base 26.
// The result uses the lower-case letters 'a' to 'z'
func FormatInt(i int64) string {
	_, s := formatBits(nil, uint64(i), i < 0, false)
	return s
}

// Itoa is equivalent to FormatInt(int64(i)).
func Itoa(i int) string {
	return FormatInt(int64(i))
}

// AppendInt appends the string form of the integer i,
// as generated by FormatInt, to dst and returns the extended buffer.
func AppendInt(dst []byte, i int64) []byte {
	dst, _ = formatBits(dst, uint64(i), i < 0, true)
	return dst
}

// AppendUint appends the string form of the unsigned integer i,
// as generated by FormatUint, to dst and returns the extended buffer.
func AppendUint(dst []byte, i uint64) []byte {
	dst, _ = formatBits(dst, i, false, true)
	return dst
}

const digits = "abcdefghijklmnopqrstuvwxyz"

// formatBits computes the string representation of u in the given base.
// If neg is set, u is treated as negative int64 value. If append_ is
// set, the string is appended to dst and the resulting byte slice is
// returned as the first result value; otherwise the string is returned
// as the second result value.
//
func formatBits(dst []byte, u uint64, neg, append_ bool) (d []byte, s string) {
	var a [64 + 1]byte // +1 for sign of 64bit value in base 2
	i := len(a)

	if neg {
		u = -u
	}

	// convert bits
	// We use uint values where we can because those will
	// fit into a single register even on a 32bit machine.
	// general case
	b := uint64(26)
	for u >= b {
		i--
		// Avoid using r = a%b in addition to q = a/b
		// since 64bit division and modulo operations
		// are calculated by runtime functions on 32bit machines.
		q := u / b
		a[i] = digits[uint(u-q*b)]
		u = q
	}

	// u < base
	i--
	a[i] = digits[uint(u)]

	// add sign, if any
	if neg {
		i--
		a[i] = '-'
	}

	if append_ {
		d = append(dst, a[i:]...)
		return
	}
	s = string(a[i:])
	return
}
