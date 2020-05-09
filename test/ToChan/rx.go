// Code generated by jig; DO NOT EDIT.

//go:generate jig

package ToChan

import (
	"sync/atomic"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Subscriber

// Subscriber is an interface that can be passed in when subscribing to an
// Observable. It allows a set of observable subscriptions to be canceled
// from a single subscriber at the root of the subscription tree.
type Subscriber = subscriber.Subscriber

// NewSubscriber creates a new subscriber.
func NewSubscriber() Subscriber {
	return subscriber.New()
}

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler = scheduler.Scheduler

//jig:name Observer

// Observer is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type Observer func(next interface{}, err error, done bool)

//jig:name Observable

// Observable is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type Observable func(Observer, Scheduler, Subscriber)

//jig:name zero

var zero interface{}

//jig:name From

// From creates an Observable from multiple interface{} values passed in.
func From(slice ...interface{}) Observable {
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
		i := 0
		runner := scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Subscribed() {
				if i < len(slice) {
					observe(slice[i], nil, false)
					if subscriber.Subscribed() {
						i++
						self()
					}
				} else {
					observe(zero, nil, true)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name IntObserver

// IntObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type IntObserver func(next int, err error, done bool)

//jig:name ObservableInt

// ObservableInt is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableInt func(IntObserver, Scheduler, Subscriber)

//jig:name zeroInt

var zeroInt int

//jig:name Range

// Range creates an ObservableInt that emits a range of sequential integers.
func Range(start, count int) ObservableInt {
	end := start + count
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		i := start
		runner := scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Subscribed() {
				if i < end {
					observe(i, nil, false)
					if subscriber.Subscribed() {
						i++
						self()
					}
				} else {
					observe(zeroInt, nil, true)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Error

// Error signals an error condition.
type Error func(error)

//jig:name Complete

// Complete signals that no more data is to be expected.
type Complete func()

//jig:name Canceled

// Canceled returns true when the observer has unsubscribed.
type Canceled func() bool

//jig:name Next

// Next can be called to emit the next value to the IntObserver.
type Next func(interface{})

//jig:name Create

// Create provides a way of creating an Observable from
// scratch by calling observer methods programmatically.
//
// The create function provided to Create will be called once
// to implement the observable. It is provided with a Next, Error,
// Complete and Canceled function that can be called by the code that
// implements the Observable.
func Create(create func(Next, Error, Complete, Canceled)) Observable {
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Canceled() {
				return
			}
			n := func(next interface{}) {
				if subscriber.Subscribed() {
					observe(next, nil, false)
				}
			}
			e := func(err error) {
				if subscriber.Subscribed() {
					observe(zero, err, true)
				}
			}
			c := func() {
				if subscriber.Subscribed() {
					observe(zero, nil, true)
				}
			}
			x := func() bool {
				return subscriber.Canceled()
			}
			create(n, e, c, x)
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name Schedulers

func TrampolineScheduler() Scheduler {
	return scheduler.Trampoline
}

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

//jig:name ObservableToChan

// ToChan returns a channel that emits interface{} values. If the source
// observable does not emit values but emits an error or complete, then the
// returned channel will enit any error and then close without emitting any
// values.
//
// This method subscribes to the observable on the Goroutine scheduler because
// it needs the concurrency so the returned channel can be used by used
// by the calling code directly. To cancel ToChan you will need to supply a
// subscriber that you hold on to.
func (o Observable) ToChan(subscribers ...Subscriber) <-chan interface{} {
	scheduler := GoroutineScheduler()
	subscribers = append(subscribers, NewSubscriber())
	donech := make(chan struct{})
	nextch := make(chan interface{})
	const (
		idle	= iota
		busy
		closed
	)
	state := int32(idle)
	observer := func(next interface{}, err error, done bool) {
		if atomic.CompareAndSwapInt32(&state, idle, busy) {
			if err != nil {
				next = err
			}
			if !done || err != nil {
				select {
				case <-donech:
					atomic.StoreInt32(&state, closed)
				default:
					select {
					case <-donech:
						atomic.StoreInt32(&state, closed)
					case nextch <- next:
					}
				}
			}
			if done {
				atomic.StoreInt32(&state, closed)
				subscribers[0].Unsubscribe()
			}
			if !atomic.CompareAndSwapInt32(&state, busy, idle) {
				close(nextch)
			}
		}
	}
	subscribers[0].OnUnsubscribe(func() {
		close(donech)
		if atomic.CompareAndSwapInt32(&state, busy, closed) {
			return
		}
		if atomic.CompareAndSwapInt32(&state, idle, closed) {
			close(nextch)
			return
		}
	})
	o(observer, scheduler, subscribers[0])
	return nextch
}

//jig:name ObservableIntToChan

// ToChan returns a channel that emits int values. If the source observable does
// not emit values but emits an error or complete, then the returned channel
// will close without emitting any values.
//
// This method subscribes to the observable on the Goroutine scheduler because
// it needs the concurrency so the returned channel can be used by used
// by the calling code directly. To cancel ToChan you will need to supply a
// subscriber that you hold on to.
func (o ObservableInt) ToChan(subscribers ...Subscriber) <-chan int {
	scheduler := GoroutineScheduler()
	subscribers = append(subscribers, NewSubscriber())
	donech := make(chan struct{})
	nextch := make(chan int)
	const (
		idle	= iota
		busy
		closed
	)
	state := int32(idle)
	observer := func(next int, err error, done bool) {
		if atomic.CompareAndSwapInt32(&state, idle, busy) {
			if !done {
				select {
				case <-donech:
					atomic.StoreInt32(&state, closed)
				default:
					select {
					case <-donech:
						atomic.StoreInt32(&state, closed)
					case nextch <- next:
					}
				}
			} else {
				atomic.StoreInt32(&state, closed)
				subscribers[0].Unsubscribe()
			}
			if !atomic.CompareAndSwapInt32(&state, busy, idle) {
				close(nextch)
			}
		}
	}
	subscribers[0].OnUnsubscribe(func() {
		close(donech)
		if atomic.CompareAndSwapInt32(&state, busy, closed) {
			return
		}
		if atomic.CompareAndSwapInt32(&state, idle, closed) {
			close(nextch)
			return
		}
	})
	o(observer, scheduler, subscribers[0])
	return nextch
}
