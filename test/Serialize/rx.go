// Code generated by jig; DO NOT EDIT.

//go:generate jig

package Serialize

import (
	"sync"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler = scheduler.Scheduler

//jig:name GoroutineScheduler

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
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

//jig:name NextInt

// NextInt can be called to emit the next value to the IntObserver.
type NextInt func(int)

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

//jig:name CreateInt

// CreateInt provides a way of creating an ObservableInt from
// scratch by calling observer methods programmatically.
//
// The create function provided to CreateInt will be called once
// to implement the observable. It is provided with a NextInt, Error,
// Complete and Canceled function that can be called by the code that
// implements the Observable.
func CreateInt(create func(NextInt, Error, Complete, Canceled)) ObservableInt {
	var zeroInt int
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

//jig:name ObservableInt_SubscribeOn

// SubscribeOn specifies the scheduler an ObservableInt should use when it is
// subscribed to.
func (o ObservableInt) SubscribeOn(subscribeOn Scheduler) ObservableInt {
	observable := func(observe IntObserver, _ Scheduler, subscriber Subscriber) {
		subscriber.OnWait(subscribeOn.Wait)
		o(observe, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Subscription

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription = subscriber.Subscription

//jig:name ObservableInt_Subscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableInt) Subscribe(observe IntObserver, subscribers ...Subscriber) Subscription {
	subscribers = append(subscribers, subscriber.New())
	scheduler := scheduler.MakeTrampoline()
	observer := func(next int, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			var zeroInt int
			observe(zeroInt, err, true)
			subscribers[0].Unsubscribe()
		}
	}
	subscribers[0].OnWait(scheduler.Wait)
	o(observer, scheduler, subscribers[0])
	return subscribers[0]
}

//jig:name ObservableInt_Serialize

// Serialize forces an ObservableInt to make serialized calls and to be
// well-behaved.
func (o ObservableInt) Serialize() ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		var observer struct {
			sync.Mutex
			done	bool
		}
		serializer := func(next int, err error, done bool) {
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
