package extchannel

import (
	"context"
	"errors"
	"sync"
)

// errors
var (
	ErrChannelClosed   = errors.New("channel is closed")
	ErrUnknownStrategy = errors.New("unknown strategy")
)

// StrategyType - type of strategy.
type StrategyType int

const (
	Blocking   StrategyType = iota // Block sending
	DropOldest                     // Delete the oldest message
	DropNewest                     // Delete the newest message
)

type AsyncChannel[T any] struct {
	ch       chan T
	capacity int
	strategy StrategyType
	mu       sync.Mutex
	closed   bool
}

// New makes a new channel with the given capacity and strategy.
func New[T any](capacity int, strategy StrategyType) *AsyncChannel[T] {
	return &AsyncChannel[T]{
		ch:       make(chan T, capacity),
		capacity: capacity,
		strategy: strategy,
	}
}

// Send sends data.
func (ac *AsyncChannel[T]) Send(ctx context.Context, value T) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.closed {
		return ErrChannelClosed
	}

	select {
	case ac.ch <- value: // Successful write
		return nil
	default: // channel is full
		switch ac.strategy {
		case Blocking:
			select {
			case ac.ch <- value:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		case DropOldest:
			<-ac.ch        // delete the oldest message
			ac.ch <- value // add a new message
			return nil
		case DropNewest:
			return nil // Ignore the new message
		default:
			return ErrUnknownStrategy
		}
	}
}

// Receive receives data.
func (ac *AsyncChannel[T]) Receive() (T, bool) {
	val, ok := <-ac.ch
	return val, ok
}

// Close closes the channel.
func (ac *AsyncChannel[T]) Close() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if !ac.closed {
		close(ac.ch)
		ac.closed = true
	}
}
