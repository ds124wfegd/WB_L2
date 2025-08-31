package main

import (
	"fmt"
	"math/rand"
	"time"
)

func asChan(vs ...int) <-chan int {
	c := make(chan int) // создаем канал
	go func() {         // вызвав функцию asChan для a и b происходит параллельная запись одного значения в канал с, т.к
		// канал с небуфферизированный, он блокируется после записи единственного значения до его считывания
		for _, v := range vs {
			c <- v
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // пауза от 0 до 1000 мс
		}
		close(c) // закрываем канал, чтобы не было deadlock
	}()
	return c
}

func merge(a, b <-chan int) <-chan int { //  читаем с каналов a и b
	//  при этом проверяем возможность чтения или закрытие канала
	c := make(chan int) // создаем небуфферизированный канал с, в который записываем единственное \
	// значение или с канала a, или с b
	// select выбирает случайный канал с данными (если данные есть в двух каналах)
	// при закрытии канала, мы присваиваем ему nil
	// канал nil в select всегда блокируется - это исключает его из выбора
	// когда оба канала становятся nil, выходной канал закрывается
	go func() {
		for {
			select {
			case v, ok := <-a:
				if ok {
					c <- v
				} else {
					a = nil
				}
			case v, ok := <-b:
				if ok {
					c <- v
				} else {
					b = nil
				}
			}
			if a == nil && b == nil {
				close(c) // закрываем канал, чтобы не было deadlock
				return
			}
		}
	}()
	return c
}

// Программа выведет последовательность чисел в случайном порядке, в зависимости
// от порядка следования горутин (например, 12436587)
func main() {
	rand.Seed(time.Now().Unix())
	// c go 1.20 следует использовать
	//  rand.New(rand.NewSource(time.Now().Unix()))
	a := asChan(1, 3, 5, 7)
	b := asChan(2, 4, 6, 8)
	c := merge(a, b)
	for v := range c {
		fmt.Print(v) // читаем значение с канала С и выводим в stdout в строку
	}
}
