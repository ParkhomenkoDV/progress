# Progress

Пакет `progress` предоставляет прогресс-бар с периодическим обновлением для Go-приложений.

## Install

```bash
go get github.com/ParkhomenkoDV/progress
```

## Usage

### Fields

| Поле         | Тип             | Описание                                                |
|--------------|-----------------|---------------------------------------------------------|
| `Interval`   | `time.Duration` | Частота обновления (например, `500*time.Millisecond`).  |
| `Description`| `string`        | Текст, выводимый перед прогресс-баром.                  |
| `Length`     | `uint8`         | Длина шкалы в символах (0 – не отображать).             |
| `Total`      | `uint64`        | Общее количество единиц работы (0 – неизвестно).        |
| `ShowETA`    | `bool`          | Показывать оценочное время до завершения.               |
| `ShowSpeed`  | `bool`          | Показывать скорость обработки (шт/с).                   |
| `Leave`      | `bool`          | Оставить прогресс полсе завершения.                     |

```go
package main

import (
    "context"
    "sync/atomic"
    "time"
    
    "github.com/ParkhomenkoDV/progress"
)

func main() {
    var items, errors uint64 // атомарные счетчики
    bar := progress.New(
        500*time.Millisecond, // интервал обновления
        "Loading",            // описание
        20,                   // длина шкалы
        100,                  // всего элементов
        true,                 // показывать ETA
        true,                 // показывать скорость
        true,                 // оставить прогресс после завершения
    )

    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        // Симуляция работы
        for i := 0; i < 100; i++ {
            atomic.AddUint64(&items, 1)
            time.Sleep(100 * time.Millisecond)
        }
        cancel()
    }()

    go bar.Show(ctx, &items, &errors)
}
```

## Result

```
Loading 100% |-------   | 69/100 ❌ 0 ⏰ 0h3m15s ⚡️ 100.0 it/s 
```

## License

MIT