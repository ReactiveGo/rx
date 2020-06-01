// Code generated by jig; DO NOT EDIT.

//go:generate jig

package Range

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

//jig:name Range

// Range creates an Observable that emits a range of sequential int values.
// The generated code will do a type conversion from int to interface{}.
func Range(start, count int) Observable {
	end := start + count
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
		i := start
		runner := scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Subscribed() {
				if i < end {
					observe(interface{}(i), nil, false)
					if subscriber.Subscribed() {
						i++
						self()
					}
				} else {
					var zero interface{}
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

//jig:name RangeInt

// RangeInt creates an ObservableInt that emits a range of sequential int values.
// The generated code will do a type conversion from int to int.
func RangeInt(start, count int) ObservableInt {
	end := start + count
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		i := start
		runner := scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Subscribed() {
				if i < end {
					observe(int(i), nil, false)
					if subscriber.Subscribed() {
						i++
						self()
					}
				} else {
					var zero int
					observe(zero, nil, true)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Observable_Println

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o Observable) Println(a ...interface{}) (err error) {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next interface{}, e error, done bool) {
		if !done {
			fmt.Println(append(a, next)...)
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

//jig:name ObservableInt_Filter

// Filter emits only those items from an ObservableInt that pass a predicate test.
func (o ObservableInt) Filter(predicate func(next int) bool) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			if done || predicate(next) {
				observe(next, err, done)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableInt_MapInt

// MapInt transforms the items emitted by an ObservableInt by applying a
// function to each item.
func (o ObservableInt) MapInt(project func(int) int) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			var mapped int
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableInt_DoOnComplete

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

//jig:name ObservableInt_Println

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableInt) Println(a ...interface{}) (err error) {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next int, e error, done bool) {
		if !done {
			fmt.Println(append(a, next)...)
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
