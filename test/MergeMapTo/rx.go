// Code generated by jig; DO NOT EDIT.

//go:generate jig

package MergeMapTo

import (
	"fmt"
	"sync"
	"sync/atomic"

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

//jig:name Observable_MergeMapTo

// MergeMapTo maps every entry emitted by the Observable into a single
// Observable. The stream of Observable items is then merged into a
// single stream of  items using the MergeAll operator.
func (o Observable) MergeMapTo(inner Observable) Observable {
	project := func(interface{}) Observable { return inner }
	return o.MapObservable(project).MergeAll()
}

//jig:name Observable_MapObservable

// MapObservable transforms the items emitted by an Observable by applying a
// function to each item.
func (o Observable) MapObservable(project func(interface{}) Observable) ObservableObservable {
	observable := func(observe ObservableObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			var mapped Observable
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Observable_Println

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o Observable) Println(a ...interface{}) error {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next interface{}, err error, done bool) {
		if !done {
			fmt.Println(append(a, next)...)
		} else {
			subscriber.Done(err)
		}
	}
	subscriber.OnWait(scheduler.Wait)
	o(observer, scheduler, subscriber)
	return subscriber.Wait()
}

//jig:name ObservableObserver

// ObservableObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type ObservableObserver func(next Observable, err error, done bool)

//jig:name ObservableObservable

// ObservableObservable is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableObservable func(ObservableObserver, Scheduler, Subscriber)

//jig:name ObservableObservable_MergeAll

// MergeAll flattens a higher order observable by merging the observables it emits.
func (o ObservableObservable) MergeAll() Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		var observers struct {
			sync.Mutex
			done	bool
			len	int32
		}
		observer := func(next interface{}, err error, done bool) {
			observers.Lock()
			defer observers.Unlock()
			if !observers.done {
				switch {
				case !done:
					observe(next, nil, false)
				case err != nil:
					observers.done = true
					var zero interface{}
					observe(zero, err, true)
				default:
					if atomic.AddInt32(&observers.len, -1) == 0 {
						var zero interface{}
						observe(zero, nil, true)
					}
				}
			}
		}
		merger := func(next Observable, err error, done bool) {
			if !done {
				atomic.AddInt32(&observers.len, 1)
				next(observer, subscribeOn, subscriber)
			} else {
				var zero interface{}
				observer(zero, err, true)
			}
		}
		runner := subscribeOn.Schedule(func() {
			if subscriber.Subscribed() {
				observers.len = 1
				o(merger, subscribeOn, subscriber)
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}
