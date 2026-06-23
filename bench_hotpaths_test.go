package log

import (
	"io"
	"testing"
)

// hotObj is a small struct used by the Any reflection-path benchmark.
type hotObj struct {
	Name string
	Age  int
	OK   bool
}

// newHotLogger returns a logger configured like the canonical BenchmarkLogger:
// Unix time format (fast path), debug level, and a discarding writer so the
// benchmark isolates serialization cost from I/O.
func newHotLogger() Logger {
	return Logger{TimeFormat: TimeFormatUnix, Level: DebugLevel, Writer: IOWriter{io.Discard}}
}

func BenchmarkZZ_HeaderOnly(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Msg("")
	}
}

func BenchmarkZZ_StrNoEscapeMsg(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Str("k", "short_no_escape").Msg("")
	}
}

func BenchmarkZZ_StrEscapeMsg(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Str("k", "needs\"escaping\tand\nnewlines").Msg("")
	}
}

func BenchmarkZZ_IntMsg(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Int("k", 1234567).Msg("")
	}
}

func BenchmarkZZ_Uint64Msg(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Uint64("k", 12345678901234).Msg("")
	}
}

func BenchmarkZZ_FloatMsg(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Float64("k", 3.14159).Msg("")
	}
}

func BenchmarkZZ_AnyStructMsg(b *testing.B) {
	l := newHotLogger()
	o := hotObj{Name: "alice", Age: 30, OK: true}
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Any("k", o).Msg("")
	}
}

func BenchmarkZZ_Msgf(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Str("foo", "bar").Msgf("hello %s", "world")
	}
}

func BenchmarkZZ_MsgPlain(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Str("foo", "bar").Msg("hello world")
	}
}

func BenchmarkZZ_MixedEntry(b *testing.B) {
	l := newHotLogger()
	b.ReportAllocs()
	for b.Loop() {
		l.Info().Str("svc", "api").Int("code", 200).Float64("dur", 12.5).Msg("request done")
	}
}
