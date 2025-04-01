package main

import (
	"context"
	"fmt"
	"time"

	"github.com/h2p2f/extchannel"
)

func main() {
	// Create a channel with a capacity of 3 and a strategy of dropping the oldest message
	ch := asyncchannel.New[int](3, asyncchannel.DropOldest)

	// Горутина-отправитель
	go func() {
		for i := 1; i <= 5; i++ {
			err := ch.Send(context.Background(), i)
			if err != nil {
				fmt.Printf("Ошибка отправки %d: %v\n", i, err)
			} else {
				fmt.Printf("Отправлено: %d\n", i)
			}
			time.Sleep(100 * time.Millisecond)
		}
		ch.Close()
	}()

	// Goroutine-receiver (slow)
	go func() {
		for {
			val, ok := ch.Receive()
			if !ok {
				fmt.Println("Канал закрыт!")
				return
			}
			fmt.Printf("Получено: %d\n", val)
			time.Sleep(500 * time.Millisecond) // Имитируем медленную обработку
		}
	}()

	time.Sleep(3 * time.Second)
}
