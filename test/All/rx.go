// Code generated by jig; DO NOT EDIT.

//go:generate jig

package All

import (
	"fmt"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Error

// Error signals an error condition.
type Error func(error)

//jig:name Complete

// Complete signals that no more data is to be expected.
type Complete func()

//jig:name Canceled

// Canceled returns true when the observer has unsubscribed.
type Canceled func() bool

//jig:name NextInt

// NextInt can be called to emit the next value to the IntObserver.
type NextInt func(int)

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

//jig:name CreateInt

// CreateInt provides a way of creating an ObservableInt from
// scratch by calling observer methods programmatically.
//
// The create function provided to CreateInt will be called once
// to implement the observable. It is provided with a NextInt, Error,
// Complete and Canceled function that can be called by the code that
// implements the Observable.
func CreateInt(create func(NextInt, Error, Complete, Canceled)) ObservableInt {
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Canceled() {
				return
			}
			n := func(next int) {
				if subscriber.Subscribed() {
					observe(next, nil, false)
				}
			}
			e := func(err error) {
				if subscriber.Subscribed() {
					observe(zeroInt, err, true)
				}
			}
			c := func() {
				if subscriber.Subscribed() {
					observe(zeroInt, nil, true)
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

//jig:name ObservableIntAll

// All determines whether all items emitted by an ObservableInt meet some
// criteria.
//
// Pass a predicate function to the All operator that accepts an item emitted
// by the source ObservableInt and returns a boolean value based on an
// evaluation of that item. All returns an ObservableBool that emits a single
// boolean value: true if and only if the source ObservableInt terminates
// normally and every item emitted by the source ObservableInt evaluated as
// true according to this predicate; false if any item emitted by the source
// ObservableInt evaluates as false according to this predicate.
func (o ObservableInt) All(predicate func(next int) bool) ObservableBool {
	condition := func(next interface{}) bool {
		return predicate(next.(int))
	}
	return o.AsObservable().All(condition)
}

//jig:name BoolObserver

// BoolObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type BoolObserver func(next bool, err error, done bool)

//jig:name ObservableBool

// ObservableBool is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableBool func(BoolObserver, Scheduler, Subscriber)

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

//jig:name ObservableIntAsObservable

// AsObservable turns a typed ObservableInt into an Observable of interface{}.
func (o ObservableInt) AsObservable() Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			observe(interface{}(next), err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Subscription

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription = subscriber.Subscription

//jig:name zeroBool

var zeroBool bool

//jig:name ObservableBoolSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe by default is performed on the Trampoline scheduler.
func (o ObservableBool) Subscribe(observe BoolObserver, subscribers ...Subscriber) Subscription {
	subscribers = append(subscribers, NewSubscriber())
	scheduler := TrampolineScheduler()
	observer := func(next bool, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroBool, err, true)
			subscribers[0].Unsubscribe()
		}
	}
	subscribers[0].OnWait(scheduler.Wait)
	o(observer, scheduler, subscribers[0])
	return subscribers[0]
}

//jig:name ObservableAll

// All determines whether all items emitted by an Observable meet some
// criteria.
//
// Pass a predicate function to the All operator that accepts an item emitted
// by the source Observable and returns a boolean value based on an
// evaluation of that item. All returns an ObservableBool that emits a single
// boolean value: true if and only if the source Observable terminates
// normally and every item emitted by the source Observable evaluated as
// true according to this predicate; false if any item emitted by the source
// Observable evaluates as false according to this predicate.
func (o Observable) All(predicate func(next interface{}) bool) ObservableBool {
	observable := func(observe BoolObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			switch {
			case !done:
				if !predicate(next) {
					observe(false, nil, false)
					observe(zeroBool, nil, true)
				}
			case err != nil:
				observe(zeroBool, err, true)
			default:
				observe(true, nil, false)
				observe(zeroBool, nil, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}
