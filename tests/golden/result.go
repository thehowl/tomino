package tomtypes

import "unsafe"

// URLMessage is the tomino message for the type
// net/url.URL
type URLMessage struct {
	Scheme string `json:"Scheme"`
	Opaque string `json:"Opaque"`
	User   *struct {
	} `json:"User"`
	Host        string `json:"Host"`
	Path        string `json:"Path"`
	RawPath     string `json:"RawPath"`
	OmitHost    bool   `json:"OmitHost"`
	ForceQuery  bool   `json:"ForceQuery"`
	RawQuery    string `json:"RawQuery"`
	Fragment    string `json:"Fragment"`
	RawFragment string `json:"RawFragment"`
}

func (msg URLMessage) MarshalBinary() ([]byte, error) {
	return msg.AppendBinary(nil)
}

func (msg URLMessage) AppendBinary(b []byte) ([]byte, error) {

	// field number 1
	b = append(b, 10)
	if len(msg.Scheme) > 0 {
		switch {
		case len(msg.Scheme) <= maxVarint1:
			b = append(b, byte(len(msg.Scheme)))
			b = append(b, msg.Scheme...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.Scheme)|0x80), byte(len(msg.Scheme)>>7))
			b = append(b, msg.Scheme...)
		default:
			b = growBytes(b, 10+len(msg.Scheme))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.Scheme)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.Scheme...)
		}
	}
	// field number 2
	b = append(b, 18)
	if len(msg.Opaque) > 0 {
		switch {
		case len(msg.Opaque) <= maxVarint1:
			b = append(b, byte(len(msg.Opaque)))
			b = append(b, msg.Opaque...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.Opaque)|0x80), byte(len(msg.Opaque)>>7))
			b = append(b, msg.Opaque...)
		default:
			b = growBytes(b, 10+len(msg.Opaque))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.Opaque)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.Opaque...)
		}
	}
	// field number 3
	b = append(b, 26)
	// msg.User - Not implemented
	// field number 4
	b = append(b, 34)
	if len(msg.Host) > 0 {
		switch {
		case len(msg.Host) <= maxVarint1:
			b = append(b, byte(len(msg.Host)))
			b = append(b, msg.Host...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.Host)|0x80), byte(len(msg.Host)>>7))
			b = append(b, msg.Host...)
		default:
			b = growBytes(b, 10+len(msg.Host))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.Host)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.Host...)
		}
	}
	// field number 5
	b = append(b, 42)
	if len(msg.Path) > 0 {
		switch {
		case len(msg.Path) <= maxVarint1:
			b = append(b, byte(len(msg.Path)))
			b = append(b, msg.Path...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.Path)|0x80), byte(len(msg.Path)>>7))
			b = append(b, msg.Path...)
		default:
			b = growBytes(b, 10+len(msg.Path))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.Path)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.Path...)
		}
	}
	// field number 6
	b = append(b, 50)
	if len(msg.RawPath) > 0 {
		switch {
		case len(msg.RawPath) <= maxVarint1:
			b = append(b, byte(len(msg.RawPath)))
			b = append(b, msg.RawPath...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.RawPath)|0x80), byte(len(msg.RawPath)>>7))
			b = append(b, msg.RawPath...)
		default:
			b = growBytes(b, 10+len(msg.RawPath))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.RawPath)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.RawPath...)
		}
	}
	// field number 7
	b = append(b, 56)

	if msg.OmitHost {
		b = append(b, 1)
	}

	// field number 8
	b = append(b, 64)

	if msg.ForceQuery {
		b = append(b, 1)
	}

	// field number 9
	b = append(b, 74)
	if len(msg.RawQuery) > 0 {
		switch {
		case len(msg.RawQuery) <= maxVarint1:
			b = append(b, byte(len(msg.RawQuery)))
			b = append(b, msg.RawQuery...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.RawQuery)|0x80), byte(len(msg.RawQuery)>>7))
			b = append(b, msg.RawQuery...)
		default:
			b = growBytes(b, 10+len(msg.RawQuery))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.RawQuery)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.RawQuery...)
		}
	}
	// field number 10
	b = append(b, 82)
	if len(msg.Fragment) > 0 {
		switch {
		case len(msg.Fragment) <= maxVarint1:
			b = append(b, byte(len(msg.Fragment)))
			b = append(b, msg.Fragment...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.Fragment)|0x80), byte(len(msg.Fragment)>>7))
			b = append(b, msg.Fragment...)
		default:
			b = growBytes(b, 10+len(msg.Fragment))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.Fragment)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.Fragment...)
		}
	}
	// field number 11
	b = append(b, 90)
	if len(msg.RawFragment) > 0 {
		switch {
		case len(msg.RawFragment) <= maxVarint1:
			b = append(b, byte(len(msg.RawFragment)))
			b = append(b, msg.RawFragment...)
		case len(b) <= maxVarint2:
			b = append(b, byte(len(msg.RawFragment)|0x80), byte(len(msg.RawFragment)>>7))
			b = append(b, msg.RawFragment...)
		default:
			b = growBytes(b, 10+len(msg.RawFragment))
			uvlen := putUvarint(b[len(b):len(b)+10], uint64(len(msg.RawFragment)))
			b = b[:len(b)+uvlen]
			b = append(b, msg.RawFragment...)
		}
	}
	return b, nil
}

// ---
// encoding helpers

const (
	// These are common when encoding lengths, and have fast paths instead of
	// calling putUvarint.
	maxVarint1 = (1 << 7) - 1
	maxVarint2 = (1 << 7) - 1
)

// Non-generic version of slices.Grow.
func growBytes(s []byte, n int) []byte {
	if n -= cap(s) - len(s); n > 0 {
		s = append(s[:cap(s)], make([]byte, n)...)[:len(s)]
	}
	return s
}

// putUvarint encodes a uint64 into buf and returns the number of bytes written.
// If the buffer is too small, PutUvarint will panic.
// Copied from package binary.
func putUvarint(buf []byte, x uint64) int {
	i := 0
	for x >= 0x80 {
		buf[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	buf[i] = byte(x)
	return i + 1
}

// putVarint encodes an int64 into buf and returns the number of bytes written.
// If the buffer is too small, PutVarint will panic.
func putVarint(buf []byte, x int64) int {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return putUvarint(buf, ux)
}

func putUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

func putUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

// avoid unused import errors.
var _ = unsafe.Pointer((*int)(nil))
