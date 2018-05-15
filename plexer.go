package multiplex

import "reflect"

// Plexer is a priority channel N:1 multiplexer
type Plexer struct {
	channels []<-chan []byte
	output   chan []byte
}

// New constructs a new Plexer. Channels should be added in highest to lowest priority order.
func New(channels ...<-chan []byte) *Plexer {
	return &Plexer{
		channels: channels,
		output:   make(chan []byte),
	}
}

// NumChannels returns the number of channels we know about
func (p Plexer) NumChannels() int {
	return len(p.channels)
}

// Out returns the output channel. This is the unified channel to read from
func (p *Plexer) Out() <-chan []byte {
	return p.output
}

// read reads from the given channel. If a message was read, it returns the
// message plus true, else returns nil and a boolean indicating if the channel
// was closed
func read(ch <-chan []byte) ([]byte, bool) {
	select {
	case msg, ok := <-ch:
		if !ok {
			// channel was closed
			return nil, true
		}
		return msg, false
	default:
		return nil, false
	}
}

// Run starts the muliplexer reading from the input channels. Should likely be
// run in a goroutine. To stop it, close all of the input channels
func (p *Plexer) Run() {
	for {
		allClosed, handledMsg := p.rangeOverChannels()
		if allClosed {
			close(p.output)
			return
		}

		if handledMsg {
			continue
		}

		// if after ranging over all the channels we didn't have any work, that means
		// that there were no messages on any of the channels. We now want to block
		// here and wait for a message from ANY channel
		cases := make([]reflect.SelectCase, 0, len(p.channels))
		for _, ch := range p.channels {
			if ch == nil {
				continue
			}
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
		}

		remaining := len(cases)
		for remaining > 0 {
			chosen, value, ok := reflect.Select(cases)
			if !ok {
				// The chosen channel has been closed, so zero out the channel to disable the case
				cases[chosen].Chan = reflect.ValueOf(nil)
				remaining--
				continue
			}

			p.output <- value.Bytes()
			break
		}
	}
}

// The first bool return value indicates if all channels have been closed or not.
// The second boolean return value indicates if a message was read from one of
// the channels or not
func (p *Plexer) rangeOverChannels() (bool, bool) {
	allClosed := true
	for i := range p.channels {
		if p.channels[i] == nil {
			continue
		}
		msg, doClose := read(p.channels[i])
		if doClose {
			// If the channel has been closed, mark as nil
			p.channels[i] = nil
			continue
		}
		allClosed = false
		if msg == nil {
			continue
		}
		p.output <- msg
		return false, true
	}

	return allClosed, false
}
