// Code generated by jig; DO NOT EDIT.

//go:generate jig

package All

import (
	"fmt"

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

//jig:name IntObserveFuncMethods

// Next is called by an ObservableInt to emit the next int value to the
// observer.
func (f IntObserveFunc) Next(next int) {
	f(next, nil, false)
}

// Error is called by an ObservableInt to report an error to the observer.
func (f IntObserveFunc) Error(err error) {
	f(zeroInt, err, true)
}

// Complete is called by an ObservableInt to signal that no more data is
// forthcoming to the observer.
func (f IntObserveFunc) Complete() {
	f(zeroInt, nil, true)
}

//jig:name IntObserver

// IntObserver is the interface used with CreateInt when implementing a custom
// observable.
type IntObserver interface {
	// Next emits the next int value.
	Next(int)
	// Error signals an error condition.
	Error(error)
	// Complete signals that no more data is to be expected.
	Complete()
	// Subscribed returns true when the subscription is currently valid.
	Subscribed() bool
}

//jig:name CreateInt

// CreateInt creates an Observable from scratch by calling observer methods
// programmatically.
func CreateInt(f func(IntObserver)) ObservableInt {
	observable := func(observe IntObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if !subscriber.Subscribed() {
				return
			}
			observer := func(next int, err error, done bool) {
				if subscriber.Subscribed() {
					observe(next, err, done)
				}
			}
			type ObserverSubscriber struct {
				IntObserveFunc
				Subscriber
			}
			f(&ObserverSubscriber{observer, subscriber})
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Schedulers

func TrampolineScheduler() Scheduler	{ return scheduler.Trampoline }

func GoroutineScheduler() Scheduler	{ return scheduler.Goroutine }

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

//jig:name BoolObserveFunc

// BoolObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type BoolObserveFunc func(next bool, err error, done bool)

//jig:name zeroBool

var zeroBool bool

//jig:name ObservableBool

// ObservableBool is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableBool func(BoolObserveFunc, Scheduler, Subscriber)

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

//jig:name ObservableBoolSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe by default is performed on the Trampoline scheduler.
func (o ObservableBool) Subscribe(observe BoolObserveFunc, subscribers ...Subscriber) Subscription {
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
	observable := func(observe BoolObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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
