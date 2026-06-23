// On platforms where the runtime implements time.now directly in assembly
// (windows, linux/amd64) or as a special case (plan9), time.now is a single
// call that yields wall+monotonic together, so there is nothing to save by
// splitting it. Keep linking time.now directly.

//go:build windows || (linux && amd64) || plan9

package log

import _ "unsafe" // for go:linkname

//go:noescape
//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64)
