// Code generated by jig; DO NOT EDIT.

//go:generate jig

package DoOnComplete

import (
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

//jig:name EmptyInt

// EmptyInt creates an Observable that emits no items but terminates normally.
func EmptyInt() ObservableInt {
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Subscribed() {
				observe(zeroInt, nil, true)
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name FromInt

// FromInt creates an ObservableInt from multiple int values passed in.
func FromInt(slice ...int) ObservableInt {
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
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
					observe(zeroInt, nil, true)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name ObservableIntDoOnComplete

// DoOnComplete calls a function when the stream completes.
func (o ObservableInt) DoOnComplete(f func()) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
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

func TrampolineScheduler() Scheduler {
	return scheduler.Trampoline
}

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

//jig:name ObservableIntToSlice

// ToSlice collects all values from the ObservableInt into an slice. The
// complete slice and any error are returned.
func (o ObservableInt) ToSlice() (slice []int, err error) {
	subscriber := NewSubscriber()
	scheduler := TrampolineScheduler()
	observer := func(next int, e error, done bool) {
		if !done {
			slice = append(slice, next)
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

//jig:name ObservableIntSubscribeOn

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

//jig:name ObservableIntSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe by default is performed on the Trampoline scheduler.
func (o ObservableInt) Subscribe(observe IntObserver, subscribers ...Subscriber) Subscription {
	subscribers = append(subscribers, NewSubscriber())
	scheduler := TrampolineScheduler()
	observer := func(next int, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroInt, err, true)
			subscribers[0].Unsubscribe()
		}
	}
	subscribers[0].OnWait(scheduler.Wait)
	o(observer, scheduler, subscribers[0])
	return subscribers[0]
}
