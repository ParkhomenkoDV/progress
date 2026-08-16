package progress

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{"нулевая", 0, "0h0m0s"},
		{"только секунды", 5 * time.Second, "0h0m5s"},
		{"только минуты", 2 * time.Minute, "0h2m0s"},
		{"только часы", 3 * time.Hour, "3h0m0s"},
		{"часы+минуты", 2*time.Hour + 30*time.Minute, "2h30m0s"},
		{"часы+минуты+секунды", 1*time.Hour + 20*time.Minute + 15*time.Second, "1h20m15s"},
		{"округление вверх", 1*time.Hour + 20*time.Minute + 15*time.Second + 500*time.Millisecond, "1h20m16s"},
		{"округление вниз", 1*time.Hour + 20*time.Minute + 15*time.Second + 200*time.Millisecond, "1h20m15s"},
		{"большая длительность", 100 * time.Hour, "100h0m0s"},
		{"отрицательная (только часы)", -1 * time.Hour, "-1h0m0s"},
		{"отрицательная (часы+минуты)", -2*time.Hour - 30*time.Minute, "-2h30m0s"},
		{"отрицательная с секундами", -1*time.Hour - 2*time.Minute - 3*time.Second, "-1h2m3s"},
		{"отрицательная с округлением", -1*time.Hour - 2*time.Minute - 3*time.Second - 600*time.Millisecond, "-1h2m4s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.dur)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}
