package tomtypes

import "time"

type TestType struct {
	Time      time.Time
	Duration  time.Duration
	FixedUint uint64 `binary:"fixed64"`
	Byte      byte
	Bytes     []byte
	ByteArr   [4]byte
	ZeroArr   [0]byte
	IntPtr    *int
	Slice     []struct{ A, B int }

	testName string
}
