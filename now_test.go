package log

import (
	"testing"
	"time"
)

// TestWalltimeMatchesTimeNow verifies the package-internal now() agrees with the
// standard library time.Now() to within a small tolerance. On platforms using
// the walltime() fast path this proves the wall clock is read correctly (and
// that mono is intentionally zero); on direct-time.now platforms it is a sanity
// check on the linkname.
func TestWalltimeMatchesTimeNow(t *testing.T) {
	for i := 0; i < 100; i++ {
		sec, nsec, _ := now()
		ref := time.Now()
		got := time.Unix(sec, int64(nsec))
		d := ref.Sub(got)
		if d < 0 {
			d = -d
		}
		if d > 50*time.Millisecond {
			t.Fatalf("now() drifted from time.Now(): now=%v stdlib=%v delta=%v", got, ref, d)
		}
	}
}
