package tomtypes

import "time"

type TestType struct {
	Time      time.Time
	Duration  time.Duration
	FixedUint uint64 `binary:"fixed64"`

	IntPtr *int

	testName string
}
