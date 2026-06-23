package log

import (
	"math"
	"strconv"
	"testing"
)

func TestAppendIntMatchesStrconv(t *testing.T) {
	intCases := []int64{
		0, 1, 9, 10, 99, 100, -1, -1234567,
		math.MinInt64, math.MaxInt64,
	}
	for _, v := range intCases {
		want := strconv.FormatInt(v, 10)
		got := string(appendInt(nil, v))
		if got != want {
			t.Errorf("appendInt(%d): got %q, want %q", v, got, want)
		}
	}

	uintCases := []uint64{
		0, 1, 9, 10, 99, 100, 12345678901234, math.MaxUint64,
	}
	for _, u := range uintCases {
		want := strconv.FormatUint(u, 10)
		got := string(appendUint(nil, u))
		if got != want {
			t.Errorf("appendUint(%d): got %q, want %q", u, got, want)
		}
	}
}
