// Code generated by jig; DO NOT EDIT.

//go:generate jig

package Timer

import (
	"fmt"
	"time"

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

//jig:name Timer

// Timer creates an ObservableInt that emits a sequence of integers (starting
// at zero) after an initialDelay has passed. Subsequent values are emitted
// using  a schedule of intervals passed in. If only the initialDelay is
// given, Timer will emit only once.
func Timer(initialDelay time.Duration, intervals ...time.Duration) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		i := 0
		runner := subscribeOn.ScheduleFutureRecursive(initialDelay, func(self func(time.Duration)) {
			if subscriber.Subscribed() {
				observe(i, nil, false)
				if subscriber.Subscribed() {
					if len(intervals) > 0 {
						self(intervals[i%len(intervals)])
					}
				}
				i++
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Observable_Take

// Take emits only the first n items emitted by an Observable.
func (o Observable) Take(n int) Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		taken := 0
		observer := func(next interface{}, err error, done bool) {
			if taken < n {
				observe(next, err, done)
				if !done {
					taken++
					if taken >= n {
						observe(nil, nil, true)
					}
				}
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableInt_Take

// Take emits only the first n items emitted by an ObservableInt.
func (o ObservableInt) Take(n int) ObservableInt {
	return o.AsObservable().Take(n).AsObservableInt()
}

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

//jig:name ObservableInt_AsObservable

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

//jig:name TimeIntervalInt

type TimeIntervalInt struct {
	Value		int
	Interval	time.Duration
}

//jig:name ObservableInt_TimeInterval

// TimeInterval intercepts the items from the source Observable and emits in
// their place a struct that indicates the amount of time that elapsed between
// pairs of emissions.
func (o ObservableInt) TimeInterval() ObservableTimeIntervalInt {
	observable := func(observe TimeIntervalIntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		begin := subscribeOn.Now()
		observer := func(next int, err error, done bool) {
			if subscriber.Subscribed() {
				if !done {
					now := subscribeOn.Now()
					observe(TimeIntervalInt{next, now.Sub(begin).Round(time.Millisecond)}, nil, false)
					begin = now
				} else {
					var zero TimeIntervalInt
					observe(zero, err, done)
				}
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name TimeIntervalIntObserver

// TimeIntervalIntObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type TimeIntervalIntObserver func(next TimeIntervalInt, err error, done bool)

//jig:name ObservableTimeIntervalInt

// ObservableTimeIntervalInt is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableTimeIntervalInt func(TimeIntervalIntObserver, Scheduler, Subscriber)

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name ErrTypecastToInt

// ErrTypecastToInt is delivered to an observer if the generic value cannot be
// typecast to int.
const ErrTypecastToInt = RxError("typecast to int failed")

//jig:name Observable_AsObservableInt

// AsObservableInt turns an Observable of interface{} into an ObservableInt.
// If during observing a typecast fails, the error ErrTypecastToInt will be
// emitted.
func (o Observable) AsObservableInt() ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextInt, ok := next.(int); ok {
					observe(nextInt, err, done)
				} else {
					var zeroInt int
					observe(zeroInt, ErrTypecastToInt, true)
				}
			} else {
				var zeroInt int
				observe(zeroInt, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableTimeIntervalInt_Println

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableTimeIntervalInt) Println(a ...interface{}) (err error) {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next TimeIntervalInt, e error, done bool) {
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
