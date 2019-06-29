// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package Merge

import (
	"fmt"
	"sync"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler scheduler.Scheduler

//jig:name Subscriber

// Subscriber is an alias for the subscriber.Subscriber interface type.
type Subscriber subscriber.Subscriber

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription subscriber.Subscription

//jig:name IntObserveFunc

// IntObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type IntObserveFunc func(next int, err error, done bool)

var zeroInt int

//jig:name ObservableInt

// ObservableInt is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableInt func(IntObserveFunc, Scheduler, Subscriber)

//jig:name FromSliceInt

// FromSliceInt creates an ObservableInt from a slice of int values passed in.
func FromSliceInt(slice []int) ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		i := 0
		subscribeOn.ScheduleRecursive(func(self func()) {
			if !subscriber.Canceled() {
				if i < len(slice) {
					observe(slice[i], nil, false)
					if !subscriber.Canceled() {
						i++
						self()
					}
				} else {
					observe(zeroInt, nil, true)
				}
			}
		})
	}
	return observable
}

//jig:name FromInts

// FromInts creates an ObservableInt from multiple int values passed in.
func FromInts(slice ...int) ObservableInt {
	return FromSliceInt(slice)
}

//jig:name JustInt

// JustInt creates an ObservableInt that emits a particular item.
func JustInt(element int) ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		done := false
		subscribeOn.ScheduleRecursive(func(self func()) {
			if !subscriber.Canceled() {
				if !done {
					observe(element, nil, false)
					if !subscriber.Canceled() {
						done = true
						self()
					}
				} else {
					observe(zeroInt, nil, true)
				}
			}
		})
	}
	return observable
}

//jig:name ObservableIntMerge

// Merge combines multiple Observables into one by merging their emissions.
// An error from any of the observables will terminate the merged observables.
func (o ObservableInt) Merge(other ...ObservableInt) ObservableInt {
	if len(other) == 0 {
		return o
	}
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		var observers struct {
			sync.Mutex
			done	bool
			len	int
		}
		observer := func(next int, err error, done bool) {
			observers.Lock()
			defer observers.Unlock()
			if !observers.done {
				switch {
				case !done:
					observe(next, nil, false)
				case err != nil:
					observers.done = true
					observe(zeroInt, err, true)
				default:
					if observers.len--; observers.len == 0 {
						observe(zeroInt, nil, true)
					}
				}
			}
		}
		subscribeOn.Schedule(func() {
			if !subscriber.Canceled() {
				observers.len = 1 + len(other)
				o(observer, subscribeOn, subscriber)
				for _, o := range other {
					if subscriber.Canceled() {
						return
					}
					o(observer, subscribeOn, subscriber)
				}
			}
		})
	}
	return observable
}

//jig:name MergeInt

// MergeInt combines multiple Observables into one by merging their emissions.
// An error from any of the observables will terminate the merged observables.
func MergeInt(observables ...ObservableInt) ObservableInt {
	if len(observables) == 0 {
		return EmptyInt()
	}
	return observables[0].Merge(observables[1:]...)
}

//jig:name EmptyInt

// EmptyInt creates an Observable that emits no items but terminates normally.
func EmptyInt() ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		subscribeOn.Schedule(func() {
			if !subscriber.Canceled() {
				observe(zeroInt, nil, true)
			}
		})
	}
	return observable
}

//jig:name ObservableIntDoOnComplete

// DoOnComplete calls a function when the stream completes.
func (o ObservableInt) DoOnComplete(f func()) ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			if err == nil && done {
				f()
			}
			observe(next, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Schedulers

func ImmediateScheduler() Scheduler	{ return scheduler.Immediate }

func CurrentGoroutineScheduler() Scheduler	{ return scheduler.CurrentGoroutine }

func NewGoroutineScheduler() Scheduler	{ return scheduler.NewGoroutine }

//jig:name ObservableIntPrintln

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
func (o ObservableInt) Println() (err error) {
	subscriber := subscriber.New()
	scheduler := CurrentGoroutineScheduler()
	observer := func(next int, e error, done bool) {
		if !done {
			fmt.Println(next)
		} else {
			err = e
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	subscriber.Wait()
	return
}
