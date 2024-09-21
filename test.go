package tomtypes

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
	// msg.Scheme - Not implemented
	// field number 2
	b = append(b, 18)
	// msg.Opaque - Not implemented
	// field number 3
	b = append(b, 26)
	// msg.User - Not implemented
	// field number 4
	b = append(b, 34)
	// msg.Host - Not implemented
	// field number 5
	b = append(b, 42)
	// msg.Path - Not implemented
	// field number 6
	b = append(b, 50)
	// msg.RawPath - Not implemented
	// field number 7
	b = append(b, 56)

	if msg.OmitHost {
		b = append(b, 1)
	} else {
		b = append(b, 0)
	}

	// field number 8
	b = append(b, 64)

	if msg.ForceQuery {
		b = append(b, 1)
	} else {
		b = append(b, 0)
	}

	// field number 9
	b = append(b, 74)
	// msg.RawQuery - Not implemented
	// field number 10
	b = append(b, 82)
	// msg.Fragment - Not implemented
	// field number 11
	b = append(b, 90)
	// msg.RawFragment - Not implemented
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
