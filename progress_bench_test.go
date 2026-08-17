package progress

import (
	"testing"
	"time"
)

func BenchmarkAdd(b *testing.B) {
	bar := &Bar{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bar.Add(1)
		}
	})
}

func BenchmarkAddError(b *testing.B) {
	bar := &Bar{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bar.AddError(1)
		}
	})
}

func BenchmarkGetLoad(b *testing.B) {
	bar := &Bar{Length: 50}
	percent := 0.75
	for i := 0; i < b.N; i++ {
		bar.getLoad(percent)
	}
}

// BenchmarkFormatDuration замеряет производительность форматирования.
func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{
		0,
		5 * time.Second,
		2*time.Minute + 30*time.Second,
		1*time.Hour + 20*time.Minute + 15*time.Second,
		100 * time.Hour,
		-2*time.Hour - 30*time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d := range durations {
			formatDuration(d)
		}
	}
}
