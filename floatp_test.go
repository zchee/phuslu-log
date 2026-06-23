package log

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"testing"
)

// TestFloat64pMatchesStrconv verifies Float64p emits exactly the fixed-precision
// rendering strconv produces, including the NaN/Inf quoted forms.
func TestFloat64pMatchesStrconv(t *testing.T) {
	type tc struct {
		f    float64
		prec int
		want string
	}
	cases := map[string]tc{
		"pi prec2":    {3.14159, 2, "3.14"},
		"pi prec6":    {3.14159, 6, "3.141590"},
		"neg prec1":   {-2.5, 1, "-2.5"},
		"zero prec0":  {0, 0, "0"},
		"round up":    {0.005, 2, "0.01"},
		"large prec0": {1234567.89, 0, "1234568"},
		"nan":         {math.NaN(), 2, `"NaN"`},
		"inf+":        {math.Inf(1), 2, `"+Inf"`},
		"inf-":        {math.Inf(-1), 2, `"-Inf"`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			l := Logger{TimeFormat: TimeFormatUnix, Level: DebugLevel, Writer: IOWriter{&buf}}
			l.Info().Float64p("k", c.f, c.prec).Msg("")
			got := stripTime(strings.TrimRight(buf.String(), "\n"))
			want := `{"time":T,"level":"info","k":` + c.want + "}"
			if got != want {
				t.Errorf("Float64p(%v,%d)\n  got:  %q\n  want: %q", c.f, c.prec, got, want)
			}
			// also cross-check the bare rendering against strconv for finite values
			if !math.IsNaN(c.f) && !math.IsInf(c.f, 0) {
				ref := string(strconv.AppendFloat(nil, c.f, 'f', c.prec, 64))
				if c.want != ref {
					t.Errorf("want literal %q disagrees with strconv %q", c.want, ref)
				}
			}
		})
	}
}

var fpSink []byte

func BenchmarkFloat64Shortest(b *testing.B) {
	buf := make([]byte, 0, 32)
	b.ReportAllocs()
	for b.Loop() {
		buf = appendFloat(buf[:0], 3.14159, 64)
	}
	fpSink = buf
}

func BenchmarkFloat64pPrec2(b *testing.B) {
	buf := make([]byte, 0, 32)
	b.ReportAllocs()
	for b.Loop() {
		buf = appendFloatPrec(buf[:0], 3.14159, 2, 64)
	}
	fpSink = buf
}
