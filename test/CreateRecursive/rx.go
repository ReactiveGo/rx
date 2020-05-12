// Code generated by jig; DO NOT EDIT.

//go:generate jig

package CreateRecursive

import (
	"fmt"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler = scheduler.Scheduler

//jig:name Subscriber

// Subscriber is an interface that can be passed in when subscribing to an
// Observable. It allows a set of observable subscriptions to be canceled
// from a single subscriber at the root of the subscription tree.
type Subscriber = subscriber.Subscriber

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

//jig:name Error

// Error signals an error condition.
type Error func(error)

//jig:name Complete

// Complete signals that no more data is to be expected.
type Complete func()

//jig:name NextInt

// NextInt can be called to emit the next value to the IntObserver.
type NextInt func(int)

//jig:name CreateRecursiveInt

// CreateRecursiveInt provides a way of creating an ObservableInt from
// scratch by calling observer methods programmatically.
//
// The create function provided to CreateRecursiveInt will be called
// repeatedly to implement the observable. It is provided with a NextInt, Error
// and Complete function that can be called by the code that implements the
// Observable.
func CreateRecursiveInt(create func(NextInt, Error, Complete)) ObservableInt {
	var zeroInt int
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		done := false
		runner := scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Canceled() {
				return
			}
			n := func(next int) {
				if subscriber.Subscribed() {
					observe(next, nil, false)
				}
			}
			e := func(err error) {
				done = true
				if subscriber.Subscribed() {
					observe(zeroInt, err, true)
				}
			}
			c := func() {
				done = true
				if subscriber.Subscribed() {
					observe(zeroInt, nil, true)
				}
			}
			create(n, e, c)
			if !done && subscriber.Subscribed() {
				self()
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name ObservableIntPrintln

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println is performed on the Trampoline scheduler.
func (o ObservableInt) Println() (err error) {
	subscriber := subscriber.New()
	scheduler := scheduler.Trampoline
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
