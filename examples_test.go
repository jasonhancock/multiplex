package multiplex

import (
	"fmt"
	"sync"
)

func Example() {
	var chans []chan []byte
	for i := 0; i < 10; i++ {
		chans = append(chans, make(chan []byte, 1000))
	}

	plexer := New(chans...)
	for i := 0; i < 100; i++ {
		chans[i%10] <- []byte(fmt.Sprintf("%d", i))
	}

	go func() {
		plexer.Run()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range plexer.Out() {
			fmt.Println(string(msg))
		}
	}()

	for i := 0; i < 10; i++ {
		close(chans[i])
	}

	wg.Wait()
}
