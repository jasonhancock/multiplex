package multiplex

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlexer(t *testing.T) {
	var chans []chan []byte
	for i := 0; i < 10; i++ {
		chans = append(chans, make(chan []byte, 1000))
	}

	plexer := New(chConv(chans...)...)
	for i := 0; i < 100; i++ {
		chans[i%10] <- []byte(fmt.Sprintf("%d", i))
	}

	go func() {
		plexer.Run()
	}()

	var values [][]byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range plexer.Out() {
			values = append(values, msg)
		}
	}()

	for i := 0; i < 10; i++ {
		close(chans[i])
	}

	wg.Wait()

	// Ensure that we read the values out in the right order
	mod := 0
	for i := 0; i < 100; i++ {
		if i%10 == 0 && i != 0 {
			mod++
		}
		val, err := strconv.Atoi(string(values[i]))
		require.NoError(t, err)
		require.True(t, val%10 == mod)
	}
}

func TestPlexerCloseChan(t *testing.T) {
	var chans []chan []byte

	for i := 0; i < 10; i++ {
		chans = append(chans, make(chan []byte, 1000))
	}

	plexer := New(chConv(chans...)...)
	go func() {
		plexer.Run()
	}()

	var values [][]byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range plexer.Out() {
			values = append(values, msg)
		}
	}()

	// Publish a message, close a channel, publish another message
	time.Sleep(100 * time.Millisecond)
	close(chans[1])
	chans[0] <- []byte("foo")
	time.Sleep(100 * time.Millisecond)
	chans[2] <- []byte("bar")

	for i := 0; i < 10; i++ {
		if i == 1 {
			continue
		}
		close(chans[i])
	}
	wg.Wait()

	require.Equal(t, []byte("foo"), values[0])
	require.Equal(t, []byte("bar"), values[1])
}

func chConv(channels ...chan []byte) []<-chan []byte {
	ret := make([]<-chan []byte, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}
