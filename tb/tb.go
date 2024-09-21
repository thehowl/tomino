// Package tb defines elementary functions to encode and decode data in the
// tomino binary encoding.
package tb

// MaxFieldNumber indicates the maximum field number which may be used.
// Beyond this number, it won't be encoded correctly.
const MaxFieldNumber uint64 = (1 << 29) - 1

// Uvarint / PutUvarint copied from source code of package binary.

const maxVarintLen64 = 10

// Uvarint decodes a uint64 from buf and returns that value and the
// number of bytes read (> 0). If an error occurred, the value is 0
// and the number of bytes n is <= 0 meaning:
//
//	n == 0: buf too small
//	n  < 0: value larger than 64 bits (overflow)
//	        and -n is the number of bytes read
func Uvarint(buf []byte) (uint64, int) {
	var x uint64
	var s uint
	for i, b := range buf {
		if i == maxVarintLen64 {
			// Catch byte reads past maxVarintLen64.
			return 0, -(i + 1) // overflow
		}
		if b < 0x80 {
			if i == maxVarintLen64-1 && b > 1 {
				return 0, -(i + 1) // overflow
			}
			return x | uint64(b)<<s, i + 1
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0
}

// PutUvarint encodes a uint64 into buf and returns the number of bytes written.
// If the buffer is too small, PutUvarint will panic.
func PutUvarint(buf []byte, x uint64) int {
	i := 0
	for x >= 0x80 {
		buf[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	buf[i] = byte(x)
	return i + 1
}
