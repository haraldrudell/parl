/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"
)

const (
	TimerCh TriggeringChan = iota
	DoneCh
	ValueCh
)

type TriggeringChan uint8

// Value is an event value that is being debounced
type Value interface{}

// SenderFunc takes an untyped value, type asserts and sends on a typed channel
type SenderFunc func([]Value)

// ReceiverFunc takes two channels, listens to them and a typed channel, returns what channel triggered and a possible untyped value
type ReceiverFunc func(c <-chan time.Time, done <-chan struct{}) (which TriggeringChan, value Value)

// Debouncer debounces event streams of Value
func NewDebouncer(d time.Duration, receiver ReceiverFunc, sender SenderFunc, ctx context.Context) (err error) {
	var values []Value
	var timer *time.Timer
	c := make(<-chan time.Time)
	C := c
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		switch which, value := receiver(C, ctx.Done()); which {
		case ValueCh:
			if timer == nil {
				timer = time.NewTimer(d)
				C = timer.C
			}
			values = append(values, value)
			continue
		case TimerCh:
			timer = nil
			C = c
			sender(values)
			values = nil
			continue
		case DoneCh:
		}
		break
	}
	return
}
