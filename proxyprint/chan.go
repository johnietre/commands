package main

import (
	"sync"

	utils "github.com/johnietre/utils/go"
)

type Chan[T any] struct {
	ch       chan T
	mtx      sync.RWMutex
	closed   bool
	closedCh *utils.AValue[chan utils.Unit]
}

func NewChan[T any](capacity int) *Chan[T] {
	return &Chan[T]{
		ch:       make(chan T, capacity),
		closedCh: utils.NewAValue(make(chan utils.Unit)),
	}
}

func (ch *Chan[T]) Recv() (t T, isOpen bool) {
	t, isOpen = <-ch.ch
	return
}

func (ch *Chan[T]) TryRecv() (t T, recvd, isOpen bool) {
	select {
	case t, isOpen = <-ch.ch:
		recvd = true
	default:
	}
	return
}

func (ch *Chan[T]) Send(t T) (isOpen bool) {
	closed := ch.closedCh.Load()
	if closed == nil {
		return false
	}
	ch.mtx.RLock()
	defer ch.mtx.RUnlock()
	if ch.closed {
		return false
	}
	select {
	case <-closed:
		return false
	case ch.ch <- t:
	}
	return true
}

func (ch *Chan[T]) TrySend(t T) (sent, isOpen bool) {
	closed := ch.closedCh.Load()
	if closed == nil {
		return false, false
	}
	ch.mtx.RLock()
	defer ch.mtx.RUnlock()
	if ch.closed {
		return
	}
	isOpen = true
	select {
	case <-closed:
		return false, false
	case ch.ch <- t:
		sent = true
	default:
	}
	return
}

func (ch *Chan[T]) Close() bool {
	closedCh, _ := ch.closedCh.Swap(nil)
	if closedCh == nil {
		return false
	}
	close(closedCh)

	ch.mtx.Lock()
	defer ch.mtx.Unlock()
	if ch.closed {
		return false
	}
	ch.closed = true
	close(ch.ch)
	return true
}

func (ch *Chan[T]) Len() int {
	return len(ch.ch)
}

func (ch *Chan[T]) Cap() int {
	return cap(ch.ch)
}

func (ch *Chan[T]) Chan() chan T {
	return ch.ch
}
