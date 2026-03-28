package main

import (
	"fmt"
	"sync"
)

// Ответ для тимлида: Финальное значение не равно 1000, потому что операция counter++
// не является атомарной, и из-за отсутствия синхронизации возникает состояние гонки (data race).

func main() {
	var counter int
	var wg sync.WaitGroup
	var mu sync.Mutex // Добавляем мьютекс для защиты счетчика

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			mu.Lock() // Блокируем доступ для других горутин
			counter++
			mu.Unlock() // Открываем доступ
		}()
	}
	wg.Wait()
	fmt.Println(counter)
}
