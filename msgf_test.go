package log

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// TestMsgfMatchesFmt verifies that Msgf produces byte-identical message output
// to the fmt-based path across the fast-path verbs AND the fallback verbs, so
// the appendFormatFast optimization never changes observable output.
func TestMsgfMatchesFmt(t *testing.T) {
	tests := map[string]struct {
		format string
		args   []any
	}{
		"plain":            {"hello world", nil},
		"str":              {"hello %s", []any{"world"}},
		"two str":          {"%s and %s", []any{"a", "b"}},
		"int":              {"n=%d", []any{42}},
		"neg int":          {"n=%d", []any{-7}},
		"int64":            {"n=%d", []any{int64(9223372036854775807)}},
		"uint8":            {"n=%d", []any{uint8(255)}},
		"mixed s d":        {"%s=%d", []any{"x", 7}},
		"v string":         {"%v", []any{"plain"}},
		"v int":            {"%v", []any{123}},
		"percent literal":  {"100%% done", nil},
		"trailing percent": {"50%%", nil},
		// fallback cases (verbs/types appendFormatFast declines):
		"bool t":     {"ok=%t", []any{true}},
		"float":      {"%f", []any{3.14}},
		"v struct":   {"%v", []any{struct{ A int }{1}}},
		"quoted":     {"%q", []any{"hi"}},
		"width":      {"%5d", []any{3}},
		"hex":        {"%x", []any{255}},
		"too few":    {"%s %s", []any{"only"}},
		"too many":   {"%s", []any{"a", "b"}},
		"v float":    {"%v", []any{1.5}},
		"escape out": {"say %s", []any{"\"hi\"\tthere"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			l := Logger{TimeFormat: TimeFormatUnix, Level: DebugLevel, Writer: IOWriter{&buf}}
			l.Info().Msgf(tt.format, tt.args...)
			got := stripTime(strings.TrimRight(buf.String(), "\n"))

			var buf2 bytes.Buffer
			l2 := Logger{TimeFormat: TimeFormatUnix, Level: DebugLevel, Writer: IOWriter{&buf2}}
			// Reference path: pre-format with fmt, then send via Msg.
			l2.Info().Msg(fmt.Sprintf(tt.format, tt.args...))
			want := stripTime(strings.TrimRight(buf2.String(), "\n"))

			if got != want {
				t.Errorf("Msgf(%q,%v)\n  got:  %q\n  want: %q", tt.format, tt.args, got, want)
			}
		})
	}
}
