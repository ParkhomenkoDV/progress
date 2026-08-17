package progress

import (
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	interval := 500 * time.Millisecond
	desc := "test"
	length := uint8(20)
	total := uint64(100)
	bar := New(interval, desc, length, total, true, true, false)

	if bar.Interval != interval {
		t.Errorf("Interval = %v, want %v", bar.Interval, interval)
	}
	if bar.Description != desc {
		t.Errorf("Description = %q, want %q", bar.Description, desc)
	}
	if bar.Length != length {
		t.Errorf("Length = %d, want %d", bar.Length, length)
	}
	if bar.Total != total {
		t.Errorf("Total = %d, want %d", bar.Total, total)
	}
	if bar.done != 0 {
		t.Errorf("done = %d, want 0", bar.done)
	}
	if bar.errs != 0 {
		t.Errorf("errs = %d, want 0", bar.errs)
	}
	if !bar.ShowETA || !bar.ShowSpeed || bar.Leave {
		t.Errorf("ShowETA=%v, ShowSpeed=%v, Leave=%v, want true, true, false", bar.ShowETA, bar.ShowSpeed, bar.Leave)
	}
}

func TestAddAndAddError(t *testing.T) {
	bar := New(time.Second, "", 0, 0, false, false, false)
	bar.Add(5)
	if atomic.LoadUint64(&bar.done) != 5 {
		t.Errorf("done = %d, want 5", bar.done)
	}
	bar.AddError(3)
	if atomic.LoadUint64(&bar.errs) != 3 {
		t.Errorf("errs = %d, want 3", bar.errs)
	}
}

func TestPrint(t *testing.T) {
	// Сохраняем оригинальный stdout
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
		r.Close()
		w.Close()
	}()

	bar := &Bar{
		Description: "Test",
		Total:       100,
		ShowETA:     true,
		ShowSpeed:   true,
		Leave:       true,
	}
	// Устанавливаем done=50, err=2
	atomic.StoreUint64(&bar.done, 50)
	atomic.StoreUint64(&bar.errs, 2)

	prevDone := uint64(40)
	prevTime := time.Now().Add(-2 * time.Second) // симулируем 2 секунды назад

	bar.print(prevDone, prevTime)

	// Закрываем писатель и читаем вывод
	w.Close()
	out, _ := io.ReadAll(r)
	output := string(out)

	// Проверяем наличие ключевых частей
	checks := []string{
		"Test",
		"50%", // потому что 50/100
		"50/100",
		"❌ 2",
		"⏰",  // ETA должно быть, т.к. скорость >0
		"⚡️", // скорость
	}
	for _, chk := range checks {
		if !strings.Contains(output, chk) {
			t.Errorf("print output missing %q, got %q", chk, output)
		}
	}
}

func TestGetLoad(t *testing.T) {
	bar := &Bar{Length: 10}
	tests := []struct {
		percent float64
		want    string
	}{
		{0.0, "|          |"},
		{0.5, "|-----     |"},
		{1.0, "|----------|"},
		{0.33, "|---       |"},
	}
	for _, tt := range tests {
		got := bar.getLoad(tt.percent)
		if got != tt.want {
			t.Errorf("getLoad(%v) = %q, want %q", tt.percent, got, tt.want)
		}
	}
	// Length = 0
	bar.Length = 0
	if s := bar.getLoad(0.5); s != "" {
		t.Errorf("getLoad with Length=0 returned %q, want empty", s)
	}
}

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
