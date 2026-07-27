package logging

import (
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		raw     string
		want    slog.Level
		wantErr bool
	}{
		{raw: "debug", want: slog.LevelDebug},
		{raw: "DEBUG", want: slog.LevelDebug},
		{raw: " info ", want: slog.LevelInfo},
		{raw: "", want: slog.LevelInfo}, // unset falls back to info
		{raw: "warn", want: slog.LevelWarn},
		{raw: "warning", want: slog.LevelWarn},
		{raw: "error", want: slog.LevelError},
		{raw: "verbose", wantErr: true}, // a typo must fail loudly at boot
	}

	for _, tc := range cases {
		got, err := ParseLevel(tc.raw)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseLevel(%q): expected an error, got level %v", tc.raw, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseLevel(%q): unexpected error: %v", tc.raw, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

func TestNormalizeCorrelationIDPreservesUsableIDs(t *testing.T) {
	for _, raw := range []string{
		"1f0a2b3c-4d5e-6f70-8192-a3b4c5d6e7f8",
		"trace_id.42:7",
		" padded-id ", // trimmed, then accepted
	} {
		got := NormalizeCorrelationID(raw)
		if got != strings.TrimSpace(raw) {
			t.Errorf("NormalizeCorrelationID(%q) = %q, want it preserved", raw, got)
		}
	}
}

func TestNormalizeCorrelationIDRejectsUnsafeIDs(t *testing.T) {
	// Anything that could inject structure into a log line, or bloat every line
	// it touches, must be replaced with a freshly minted id.
	cases := map[string]string{
		"empty":     "",
		"newline":   "abc\ndef",
		"quote":     `abc"level":"error"`,
		"space":     "abc def",
		"unicode":   "abc‮def",
		"too long":  strings.Repeat("a", maxCorrelationIDLen+1),
		"json-ish":  `{"a":1}`,
		"semicolon": "abc;def",
	}

	for name, raw := range cases {
		got := NormalizeCorrelationID(raw)
		if got == raw {
			t.Errorf("%s: NormalizeCorrelationID(%q) returned it unchanged", name, raw)
		}
		if got == "" {
			t.Errorf("%s: NormalizeCorrelationID(%q) returned an empty id", name, raw)
		}
		// The replacement must itself be safe, i.e. survive a second pass.
		if NormalizeCorrelationID(got) != got {
			t.Errorf("%s: replacement id %q is not itself acceptable", name, got)
		}
	}
}

func TestFromContextFallsBackToDefault(t *testing.T) {
	if got := FromContext(context.Background()); got != slog.Default() {
		t.Error("FromContext on a bare context should return the default logger")
	}

	want := slog.New(slog.DiscardHandler)
	ctx := WithLogger(context.Background(), want)
	if got := FromContext(ctx); got != want {
		t.Error("FromContext should return the logger stored by WithLogger")
	}
}

func TestCorrelationIDRoundTrip(t *testing.T) {
	if got := CorrelationIDFrom(context.Background()); got != "" {
		t.Errorf("CorrelationIDFrom on a bare context = %q, want empty", got)
	}

	ctx := WithCorrelationID(context.Background(), "abc-123")
	if got := CorrelationIDFrom(ctx); got != "abc-123" {
		t.Errorf("CorrelationIDFrom = %q, want %q", got, "abc-123")
	}
}
