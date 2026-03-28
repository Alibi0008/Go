package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Stage 1: Исходный код серверов (его по заданию менять нельзя)
func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
				out <- fmt.Sprintf("[%s] metric: %d", name, rand.Intn(100))
			}
		}
	}()
	return out
}

// ЗАДАЧА 1: Реализация функции FanIn
func FanIn(ctx context.Context, channels ...<-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup

	// Функция-помощник: читает из одной маленькой трубы и перекидывает в большую (out)
	multiplex := func(c <-chan string) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done(): // Если время вышло - уходим
				return
			case val, ok := <-c:
				if !ok {
					return // Если канал закрылся - уходим
				}
				// Пытаемся отправить данные в общий канал
				select {
				case out <- val:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	// Запускаем горутину-помощника для КАЖДОГО входящего канала (Alpha, Beta, Gamma)
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	// Отдельная горутина ждет, когда все помощники закончат работу, и закрывает общую трубу
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// ЗАДАЧА 2: Написание функции main
func main() {
	// Создаем контекст с таймаутом ровно на 2 секунды
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Запускаем 3 сервера
	ch1 := startServer(ctx, "Alpha")
	ch2 := startServer(ctx, "Beta")
	ch3 := startServer(ctx, "Gamma")

	// Скармливаем каналы серверов в нашу "воронку" FanIn
	ch4 := FanIn(ctx, ch1, ch2, ch3)

	// Читаем объединенные данные, пока канал ch4 не закроется
	for val := range ch4 {
		fmt.Println(val)
	}
}
