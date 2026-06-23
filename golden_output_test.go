package log

import (
	"bufio"
	"bytes"
	"flag"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "regenerate testdata/golden.txt from current output")

// stripTime replaces the bare-decimal Unix time value emitted by the
// TimeFormatUnix fast path with a fixed sentinel "T", so the remaining
// field-serialization bytes are deterministic and stable across timestamp
// implementation changes (e.g. the runtime.walltime optimization).
func stripTime(s string) string {
	const p = `{"time":`
	if !strings.HasPrefix(s, p) {
		return s
	}
	i := len(p)
	for i < len(s) && s[i] != ',' {
		i++
	}
	return p + "T" + s[i:]
}

// goldenObj is the struct serialized by the any/struct golden case.
type goldenObj struct {
	Name string
	Age  int
	OK   bool
}

// goldenCases returns the corpus in a stable, sorted-by-name order. Each case's
// output is captured live so control bytes are reproduced exactly; the expected
// bytes live in testdata/golden.txt (regenerate with -update-golden), which
// avoids fragile source-literal transcription of escape sequences.
func goldenCases() []struct {
	name string
	run  func(l *Logger)
} {
	long := strings.Repeat("a", 200)
	cases := []struct {
		name string
		run  func(l *Logger)
	}{
		{"str/empty", func(l *Logger) { l.Info().Str("k", "").Msg("") }},
		{"str/ascii", func(l *Logger) { l.Info().Str("k", "short_no_escape").Msg("") }},
		{"str/quote", func(l *Logger) { l.Info().Str("k", "a\"b").Msg("") }},
		{"str/backslash", func(l *Logger) { l.Info().Str("k", "a\\b").Msg("") }},
		{"str/newline", func(l *Logger) { l.Info().Str("k", "a\nb").Msg("") }},
		{"str/cr", func(l *Logger) { l.Info().Str("k", "a\rb").Msg("") }},
		{"str/tab", func(l *Logger) { l.Info().Str("k", "a\tb").Msg("") }},
		{"str/formfeed", func(l *Logger) { l.Info().Str("k", "a\fb").Msg("") }},
		{"str/backspace", func(l *Logger) { l.Info().Str("k", "a\bb").Msg("") }},
		{"str/lt", func(l *Logger) { l.Info().Str("k", "a<b").Msg("") }},
		{"str/apos", func(l *Logger) { l.Info().Str("k", "a'b").Msg("") }},
		{"str/nul", func(l *Logger) { l.Info().Str("k", "a\x00b").Msg("") }},
		{"str/allescapes", func(l *Logger) { l.Info().Str("k", "\"\\\n\r\t\f\b<'\x00").Msg("") }},
		{"str/utf8", func(l *Logger) { l.Info().Str("k", "日本語").Msg("") }},
		{"str/emoji", func(l *Logger) { l.Info().Str("k", "a😀b").Msg("") }},
		{"str/long", func(l *Logger) { l.Info().Str("k", long).Msg("") }},
		{"int/zero", func(l *Logger) { l.Info().Int("k", 0).Msg("") }},
		{"int/one", func(l *Logger) { l.Info().Int("k", 1).Msg("") }},
		{"int/nine", func(l *Logger) { l.Info().Int("k", 9).Msg("") }},
		{"int/ten", func(l *Logger) { l.Info().Int("k", 10).Msg("") }},
		{"int/neg", func(l *Logger) { l.Info().Int("k", -1).Msg("") }},
		{"int/negbig", func(l *Logger) { l.Info().Int("k", -1234567).Msg("") }},
		{"int/minint64", func(l *Logger) { l.Info().Int("k", math.MinInt64).Msg("") }},
		{"int/maxint64", func(l *Logger) { l.Info().Int("k", math.MaxInt64).Msg("") }},
		{"int/99", func(l *Logger) { l.Info().Int("k", 99).Msg("") }},
		{"int/100", func(l *Logger) { l.Info().Int("k", 100).Msg("") }},
		{"uint64/zero", func(l *Logger) { l.Info().Uint64("k", 0).Msg("") }},
		{"uint64/max", func(l *Logger) { l.Info().Uint64("k", math.MaxUint64).Msg("") }},
		{"uint64/mid", func(l *Logger) { l.Info().Uint64("k", 12345678901234).Msg("") }},
		{"float/zero", func(l *Logger) { l.Info().Float64("k", 0).Msg("") }},
		{"float/one", func(l *Logger) { l.Info().Float64("k", 1.0).Msg("") }},
		{"float/pi", func(l *Logger) { l.Info().Float64("k", 3.14159).Msg("") }},
		{"float/neg", func(l *Logger) { l.Info().Float64("k", -2.5).Msg("") }},
		{"float/nan", func(l *Logger) { l.Info().Float64("k", math.NaN()).Msg("") }},
		{"float/inf+", func(l *Logger) { l.Info().Float64("k", math.Inf(1)).Msg("") }},
		{"float/inf-", func(l *Logger) { l.Info().Float64("k", math.Inf(-1)).Msg("") }},
		{"float/1e-7", func(l *Logger) { l.Info().Float64("k", 1e-7).Msg("") }},
		{"float/1e21", func(l *Logger) { l.Info().Float64("k", 1e21).Msg("") }},
		{"float/small", func(l *Logger) { l.Info().Float64("k", 0.0001).Msg("") }},
		{"bool/true", func(l *Logger) { l.Info().Bool("k", true).Msg("") }},
		{"bool/false", func(l *Logger) { l.Info().Bool("k", false).Msg("") }},
		{"bytes/nil", func(l *Logger) { l.Info().Bytes("k", nil).Msg("") }},
		{"bytes/empty", func(l *Logger) { l.Info().Bytes("k", []byte{}).Msg("") }},
		{"bytes/ascii", func(l *Logger) { l.Info().Bytes("k", []byte("hello")).Msg("") }},
		{"bytes/escape", func(l *Logger) { l.Info().Bytes("k", []byte("hi\nbye\t\"x\"")).Msg("") }},
		{"any/nil", func(l *Logger) { l.Info().Any("k", nil).Msg("") }},
		{"any/struct", func(l *Logger) { l.Info().Any("k", goldenObj{"alice", 30, true}).Msg("") }},
		{"any/map", func(l *Logger) { l.Info().Any("k", map[string]int{"a": 1}).Msg("") }},
		{"any/slice", func(l *Logger) { l.Info().Any("k", []int{1, 2, 3}).Msg("") }},
		{"any/string", func(l *Logger) { l.Info().Any("k", "plain").Msg("") }},
		{"any/int", func(l *Logger) { l.Info().Any("k", 42).Msg("") }},
		{"msg/empty", func(l *Logger) { l.Info().Str("k", "v").Msg("") }},
		{"msg/plain", func(l *Logger) { l.Info().Str("foo", "bar").Msg("hello world") }},
		{"msg/quote", func(l *Logger) { l.Info().Msg("say \"hi\"") }},
		{"msg/newline", func(l *Logger) { l.Info().Msg("line1\nline2") }},
		{"msgf/str", func(l *Logger) { l.Info().Str("foo", "bar").Msgf("hello %s", "world") }},
		{"msgf/int", func(l *Logger) { l.Info().Msgf("n=%d", 42) }},
		{"msgf/mixed", func(l *Logger) { l.Info().Msgf("%s=%d ok=%t", "x", 7, true) }},
		{"mixed/entry", func(l *Logger) { l.Info().Str("svc", "api").Int("code", 200).Float64("dur", 12.5).Msg("request done") }},
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].name < cases[j].name })
	return cases
}

// captureGolden runs a case and returns its time-stripped single-line output.
// The result is strconv-quoted so control bytes survive a line-oriented file.
func captureGolden(run func(l *Logger)) string {
	var buf bytes.Buffer
	l := Logger{TimeFormat: TimeFormatUnix, Level: DebugLevel, Writer: IOWriter{&buf}}
	run(&l)
	return strings.TrimRight(stripTime(buf.String()), "\n")
}

const goldenPath = "testdata/golden.txt"

// TestGoldenOutput freezes the exact field-serialization bytes of the current
// implementation. Every optimization MUST keep this empty-diff; an intentional
// output change is made by re-running with -update-golden and is then visible
// in the testdata diff. Lines are "name\tstrconv.Quote(output)".
func TestGoldenOutput(t *testing.T) {
	cases := goldenCases()

	if *updateGolden {
		var sb strings.Builder
		for _, c := range cases {
			sb.WriteString(c.name)
			sb.WriteByte('\t')
			sb.WriteString(strconv.Quote(captureGolden(c.run)))
			sb.WriteByte('\n')
		}
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(sb.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d golden cases to %s", len(cases), goldenPath)
		return
	}

	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("open golden file (run `go test -run TestGoldenOutput -update-golden` first): %v", err)
	}
	defer f.Close()

	want := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		name, quoted, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		v, err := strconv.Unquote(quoted)
		if err != nil {
			t.Fatalf("corrupt golden line %q: %v", line, err)
		}
		want[name] = v
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	if len(want) != len(cases) {
		t.Fatalf("golden file has %d cases, test defines %d — regenerate with -update-golden", len(want), len(cases))
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := captureGolden(c.run)
			w, ok := want[c.name]
			if !ok {
				t.Fatalf("no golden entry for %q — regenerate with -update-golden", c.name)
			}
			if got != w {
				t.Errorf("golden mismatch\n  got:  %q\n  want: %q", got, w)
			}
		})
	}
}
