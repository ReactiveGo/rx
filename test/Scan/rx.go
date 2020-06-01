// Code generated by jig; DO NOT EDIT.

//go:generate jig

package Scan

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

//jig:name FromInt

// FromInt creates an ObservableInt from multiple int values passed in.
func FromInt(slice ...int) ObservableInt {
	var zeroInt int
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

//jig:name ObservableInt_ScanInt

// ScanInt applies a accumulator function to each item emitted by an
// ObservableInt and the previous accumulator result. The operator accepts a
// seed argument that is passed to the accumulator for the first item emitted
// by the ObservableInt. ScanInt emits every value, both intermediate and final.
func (o ObservableInt) ScanInt(accumulator func(int, int) int, seed int) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		state := seed
		observer := func(next int, err error, done bool) {
			if !done {
				state = accumulator(state, next)
				observe(state, nil, false)
			} else {
				var zeroInt int
				observe(zeroInt, err, done)
			}
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
