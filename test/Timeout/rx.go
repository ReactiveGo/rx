// Code generated by jig; DO NOT EDIT.

//go:generate jig

package Timeout

import (
	"fmt"
	"sync"
	"time"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler = scheduler.Scheduler

//jig:name Subscriber

// Subscriber is an alias for the subscriber.Subscriber interface type.
type Subscriber = subscriber.Subscriber

// NewSubscriber creates a new subscriber.
func NewSubscriber() Subscriber {
	return subscriber.New()
}

//jig:name IntObserveFunc

// IntObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type IntObserveFunc func(next int, err error, done bool)

//jig:name zeroInt

var zeroInt int

//jig:name ObservableInt

// ObservableInt is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableInt func(IntObserveFunc, Scheduler, Subscriber)

//jig:name MakeTimedIntFunc

// MakeTimedIntFunc is the signature of a function that can be passed to
// MakeTimedInt to implement an ObservableInt.
type MakeTimedIntFunc func(Next func(int), Error func(error), Complete func()) time.Duration

//jig:name MakeTimedInt

// MakeTimedInt provides a way of creating an ObservableInt from scratch by
// calling observer methods programmatically. A make function conforming to the
// MakeIntFunc signature will be called by MakeInt provining a Next, Error and
// Complete function that can be called by the code that implements the
// Observable. The timeout passed in determines the time between calling the
// make function. The time.Duration returned by MakeTimedIntFunc determines when
// to reschedule the next iteration.
func MakeTimedInt(timeout time.Duration, make MakeTimedIntFunc) ObservableInt {
	observable := func(observe IntObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		done := false
		runner := scheduler.ScheduleFutureRecursive(timeout, func(self func(time.Duration)) {
			if subscriber.Canceled() {
				return
			}
			next := func(n int) {
				if subscriber.Subscribed() {
					observe(n, nil, false)
				}
			}
			error := func(e error) {
				done = true
				if subscriber.Subscribed() {
					observe(zeroInt, e, true)
				}
			}
			complete := func() {
				done = true
				if subscriber.Subscribed() {
					observe(zeroInt, nil, true)
				}
			}
			timeout = make(next, error, complete)
			if !done && subscriber.Subscribed() {
				self(timeout)
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name ObservableSerialize

// Serialize forces an Observable to make serialized calls and to be
// well-behaved.
func (o Observable) Serialize() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		var observer struct {
			sync.Mutex
			done	bool
		}
		serializer := func(next interface{}, err error, done bool) {
			observer.Lock()
			defer observer.Unlock()
			if !observer.done {
				observer.done = done
				observe(next, err, done)
			}
		}
		o(serializer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableTimeout

// ErrTimeout is delivered to an observer if the stream times out.
const ErrTimeout = RxError("timeout")

// Timeout mirrors the source Observable, but issues an error notification if a
// particular period of time elapses without any emitted items.
// Timeout schedules tasks on the scheduler passed to this
func (o Observable) Timeout(timeout time.Duration) Observable {
	observable := Observable(func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		if subscriber.Canceled() {
			return
		}
		var last struct {
			sync.Mutex
			at	time.Time
			done	bool
		}
		last.at = subscribeOn.Now()
		timer := func(self func(time.Duration)) {
			last.Lock()
			defer last.Unlock()
			if last.done || subscriber.Canceled() {
				return
			}
			deadline := last.at.Add(timeout)
			now := subscribeOn.Now()
			if now.Before(deadline) {
				self(deadline.Sub(now))
				return
			}
			last.done = true
			observe(zero, ErrTimeout, true)
		}
		runner := subscribeOn.ScheduleFutureRecursive(timeout, timer)
		subscriber.OnUnsubscribe(runner.Cancel)
		observer := func(next interface{}, err error, done bool) {
			last.Lock()
			defer last.Unlock()
			if last.done || subscriber.Canceled() {
				return
			}
			now := subscribeOn.Now()
			deadline := last.at.Add(timeout)
			if !now.Before(deadline) {
				return
			}
			last.done = done
			last.at = now
			observe(next, err, done)
		}
		o(observer, subscribeOn, subscriber)
	})
	return observable.Serialize()
}

//jig:name ObservableIntTimeout

// Timeout mirrors the source ObservableInt, but issues an error notification if
// a particular period of time elapses without any emitted items.
//
// This observer starts a goroutine for every subscription to monitor the
// timeout deadline. It is guaranteed that calls to the observer for this
// subscription will never be called concurrently. It is however almost certain
// that any timeout error will be delivered on a goroutine other than the one
// delivering the next values.
func (o ObservableInt) Timeout(timeout time.Duration) ObservableInt {
	return o.AsObservable().Timeout(timeout).AsObservableInt()
}

//jig:name ObserveFunc

// ObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type ObserveFunc func(next interface{}, err error, done bool)

//jig:name zero

var zero interface{}

//jig:name Observable

// Observable is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type Observable func(ObserveFunc, Scheduler, Subscriber)

//jig:name ObservableIntAsObservable

// AsObservable turns a typed ObservableInt into an Observable of interface{}.
func (o ObservableInt) AsObservable() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			observe(interface{}(next), err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntSubscribeOn

// SubscribeOn specifies the scheduler an ObservableInt should use when it is
// subscribed to.
func (o ObservableInt) SubscribeOn(subscribeOn Scheduler) ObservableInt {
	observable := func(observe IntObserveFunc, _ Scheduler, subscriber Subscriber) {
		subscriber.OnWait(subscribeOn.Wait)
		o(observe, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Schedulers

func TrampolineScheduler() Scheduler {
	return scheduler.Trampoline
}

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

//jig:name ObservableIntPrintln

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println is performed on the Trampoline scheduler.
func (o ObservableInt) Println() (err error) {
	subscriber := NewSubscriber()
	scheduler := TrampolineScheduler()
	observer := func(next int, e error, done bool) {
		if !done {
			fmt.Println(next)
		} else {
			err = e
			subscriber.Unsubscribe()
		}
	}
	subscriber.OnWait(scheduler.Wait)
	o(observer, scheduler, subscriber)
	subscriber.Wait()
	return
}

//jig:name ErrTypecastToInt

// ErrTypecastToInt is delivered to an observer if the generic value cannot be
// typecast to int.
const ErrTypecastToInt = RxError("typecast to int failed")

//jig:name ObservableAsObservableInt

// AsInt turns an Observable of interface{} into an ObservableInt. If during
// observing a typecast fails, the error ErrTypecastToInt will be emitted.
func (o Observable) AsObservableInt() ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextInt, ok := next.(int); ok {
					observe(nextInt, err, done)
				} else {
					observe(zeroInt, ErrTypecastToInt, true)
				}
			} else {
				observe(zeroInt, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}