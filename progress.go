package progress

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Bar struct {
	Interval    time.Duration // Частота обновления
	Description string        // Описание
	Length      uint8         // Длина окна (0 - не отображать)
	Total       uint64        // Общее количество единиц работы (0 – неизвестно)
	ShowSpeed   bool          // Показывать скорость обработки (шт/сек)
	ShowETA     bool          // Показывать оценочное время до завершения
}

func New(
	interval time.Duration,
	description string,
	length uint8,
	total uint64,
	showSpeed, showETA bool,
) *Bar {
	return &Bar{
		Interval:    interval.Abs(),
		Description: description,
		Length:      length,
		Total:       total,
		ShowSpeed:   showSpeed,
		ShowETA:     showETA,
	}
}

// Show отображает прогресс в реальном времени.
// Параметры:
//   - items   – указатель на атомарный счётчик обработанных элементов
//   - errors  – указатель на атомарный счётчик ошибок (может быть nil)
func (b *Bar) Show(ctx context.Context, items, errors *uint64) {
	ticker := time.NewTicker(b.Interval)
	defer ticker.Stop()

	// Буферизованный вывод снижает число системных вызовов.
	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	var (
		prevItems uint64
		prevTime  = time.Now()
	)

	// Выводим прогресс
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.print(bw, items, errors, prevItems, prevTime)
			// обновляем предыдущие значения после вывода
			prevItems = atomic.LoadUint64(items)
			prevTime = time.Now()
		}
	}
}

// print формирует и выводит строку прогресса.
func (b *Bar) print(bw *bufio.Writer, items, errors *uint64, prevItems uint64, prevTime time.Time) {
	now := time.Now()
	itms := atomic.LoadUint64(items)

	var line string = fmt.Sprintf("\r%s ", b.Description)

	if b.Total > 0 {
		percent := float64(itms) / float64(b.Total)
		load := b.getLoad(percent)
		line += fmt.Sprintf("%s %d / %d (%.1f%%)", load, itms, b.Total, percent*100)
	} else {
		line += fmt.Sprintf("%d", itms)
	}

	// Счётчики успехов и ошибок
	if errors != nil {
		line += fmt.Sprintf(" | ❌ %d", atomic.LoadUint64(errors))
	}

	// Скорость (items/sec)
	if b.ShowSpeed {
		elapsed := now.Sub(prevTime).Seconds()
		if elapsed > 0 && itms > prevItems {
			speed := float64(itms-prevItems) / elapsed
			line += fmt.Sprintf(" | %.1f шт/с", speed)
		}
	}

	// ETA (оценочное время до завершения)
	if b.ShowETA && b.Total > 0 && itms > 0 && itms < b.Total {
		elapsed := now.Sub(prevTime).Seconds()
		if elapsed > 0 && itms > prevItems {
			rate := float64(itms-prevItems) / elapsed
			if rate > 0 {
				remaining := float64(b.Total-itms) / rate
				line += fmt.Sprintf(" | ETA: %s", formatDuration(time.Duration(remaining*float64(time.Second))))
			}
		}
	}

	// Очищаем текущую строку перед выводом, чтобы избежать артефактов.
	line = "\033[2K" + line

	fmt.Fprint(bw, line) // Запись в буферизованный writer.
	bw.Flush()           // Немедленный вывод
}

func (b *Bar) getLoad(percent float64) string {
	if b.Length == 0 {
		return ""
	}
	done := int(percent * float64(b.Length))
	extra := int(b.Length) - done
	return "|" + strings.Repeat("-", done) + strings.Repeat(" ", extra) + "|"
}

// formatDuration форматирует длительность в удобочитаемый вид.
func formatDuration(dur time.Duration) string {
	totalSec := int64(dur.Round(time.Second) / time.Second)
	neg := totalSec < 0
	if neg {
		totalSec = -totalSec
	}

	var (
		h = totalSec / 3600        // часы
		m = (totalSec % 3600) / 60 // минуты
		s = totalSec % 60          // секунды
	)

	// Буфер достаточного размера: знак + до 10 цифр часов + 'h' + 2 цифры минут + 'm' + 2 цифры секунд + 's' = 18
	buf := make([]byte, 0, 20)
	if neg {
		buf = append(buf, '-')
	}
	buf = strconv.AppendInt(buf, h, 10)
	buf = append(buf, 'h')
	buf = strconv.AppendInt(buf, m, 10)
	buf = append(buf, 'm')
	buf = strconv.AppendInt(buf, s, 10)
	buf = append(buf, 's')
	return string(buf)
}
