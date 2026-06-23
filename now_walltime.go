// On platforms where the runtime implements time.now indirectly as
// walltime()+nanotime() (darwin, the BSDs, non-amd64 linux, solaris, aix,
// wasip1, js), time.now pays for a monotonic-clock read that a logger never
// uses. Call runtime.walltime directly and skip the nanotime() call, returning
// mono == 0 since every caller discards it.
//
// runtime.walltime carries an explicit compatibility guarantee
// (go.dev/issue/67401), the same promise time.now relies on.

//go:build !windows && !(linux && amd64) && !plan9

package log

import _ "unsafe" // for go:linkname

//go:noescape
//go:linkname walltime runtime.walltime
func walltime() (sec int64, nsec int32)

func now() (sec int64, nsec int32, mono int64) {
	sec, nsec = walltime()
	return sec, nsec, 0
}
