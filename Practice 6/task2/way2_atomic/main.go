package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Ответ для тимлида: Финальное значение не равно 1000, потому что операция counter++
// не является атомарной, и из-за отсутствия синхронизации возникает состояние гонки (data race).

func main() {
	var counter atomic.Int64 // Используем специальный потокобезопасный тип
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			counter.Add(1) // Атомарно увеличиваем на 1
		}()
	}
	wg.Wait()
	fmt.Println(counter.Load()) // Атомарно читаем результат
}
