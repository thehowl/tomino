package tomtypes

import (
	"net/url"
	"time"
)

type TestType struct {
	Time      time.Time
	Duration  time.Duration
	FixedUint uint64 `binary:"fixed64"`
	Byte      byte
	Bytes     []byte
	ByteArr   *[4]byte
	ZeroArr   [0]byte
	IntPtr    *int
	Slice     []struct{ A, B int }
	URL       url.URL
	TReq      TestTypeRequired
	TNotReq   TestTypeNotRequired

	testName string
}

type TestTypeNotRequired struct {
	A int
	B CustomInt
}

type CustomInt int64

type TestTypeRequired TestTypeNotRequired

// TODO: type TestTypeAliased = TestTypeNotRequired
