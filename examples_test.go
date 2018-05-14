package multiplex

import (
	"fmt"
	"sync"
)

func Example() {
	chLow := make(chan []byte, 10)
	chHigh := make(chan []byte, 10)

	plexer := New(chHigh, chLow)

	chLow <- []byte("low1")
	chLow <- []byte("low2")
	chHigh <- []byte("high1")
	chHigh <- []byte("high2")

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

	close(chLow)
	close(chHigh)

	wg.Wait()
}
