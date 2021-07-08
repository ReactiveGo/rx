// Code generated by jig; DO NOT EDIT.

//go:generate jig

package BufferTime

import (
	"fmt"
	"sync"
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

//jig:name TimerInt

// TimerInt creates an ObservableInt that emits a sequence of integers
// (starting at zero) after an initialDelay has passed. Subsequent values are
// emitted using  a schedule of intervals passed in. If only the initialDelay
// is given, Timer will emit only once.
func TimerInt(initialDelay time.Duration, intervals ...time.Duration) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		i := 0
		runner := subscribeOn.ScheduleFutureRecursive(initialDelay, func(self func(time.Duration)) {
			if subscriber.Subscribed() {
				if i == 0 || (i > 0 && len(intervals) > 0) {
					observe(int(i), nil, false)
				}
				if subscriber.Subscribed() {
					if len(intervals) > 0 {
						self(intervals[i%len(intervals)])
					} else {
						if i == 0 {
							self(0)
						} else {
							var zero int
							observe(zero, nil, true)
						}
					}
				}
				i++
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
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

//jig:name From

// From creates an Observable from multiple interface{} values passed in.
func From(slice ...interface{}) Observable {
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
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
					var zero interface{}
					observe(zero, nil, true)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Empty

// Empty creates an Observable that emits no items but terminates normally.
func Empty() Observable {
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Subscribed() {
				var zero interface{}
				observe(zero, nil, true)
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

//jig:name ObservableInt_ConcatMap

// ConcatMap transforms the items emitted by an ObservableInt by applying a
// function to each item and returning an Observable. The stream of
// Observable items is then flattened by concattenating the emissions from
// the observables without interleaving.
func (o ObservableInt) ConcatMap(project func(int) Observable) Observable {
	return o.MapObservable(project).ConcatAll()
}

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name TypecastFailed

// ErrTypecast is delivered to an observer if the generic value cannot be
// typecast to a specific type.
const TypecastFailed = RxError("typecast failed")

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
					var zero int
					observe(zero, TypecastFailed, true)
				}
			} else {
				var zero int
				observe(zero, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableInt_MapObservable

// MapObservable transforms the items emitted by an ObservableInt by applying a
// function to each item.
func (o ObservableInt) MapObservable(project func(int) Observable) ObservableObservable {
	observable := func(observe ObservableObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
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

//jig:name Observable_BufferTime

// BufferTime buffers the source Observable values for a specific time period
// and emits those as a slice periodically in time.
func (o Observable) BufferTime(period time.Duration) ObservableSlice {
	return o.Buffer(Interval(period))
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

//jig:name SliceObserver

// SliceObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type SliceObserver func(next Slice, err error, done bool)

//jig:name ObservableSlice

// ObservableSlice is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableSlice func(SliceObserver, Scheduler, Subscriber)

//jig:name Observable_Buffer

// Buffer buffers the source Observable values until closingNotifier emits.
func (o Observable) Buffer(closingNotifier Observable) ObservableSlice {
	observable := func(observe SliceObserver, subscribeOn Scheduler, subscriber Subscriber) {
		var serializer struct {
			sync.Mutex
			next	[]interface{}
			done	bool
		}

		notifier := func(next interface{}, err error, done bool) {
			serializer.Lock()
			defer serializer.Unlock()
			if !serializer.done {
				serializer.done = done
				switch {
				case !done:
					observe(serializer.next, nil, false)
					serializer.next = serializer.next[:0]
				case err != nil:
					observe(nil, err, true)
					serializer.next = nil
				default:
					observe(serializer.next, nil, false)
					observe(nil, nil, true)
					serializer.next = nil
				}
			}
		}
		closingNotifier(notifier, subscribeOn, subscriber)

		observer := func(next interface{}, err error, done bool) {
			serializer.Lock()
			defer serializer.Unlock()
			if !serializer.done {
				serializer.done = done
				switch {
				case !done:
					serializer.next = append(serializer.next, next)
				case err != nil:
					observe(nil, err, true)
					serializer.next = nil
				default:
					observe(serializer.next, nil, false)
					observe(nil, nil, true)
					serializer.next = nil
				}
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Interval

// Interval creates an Observable that emits a sequence of integers spaced
// by a particular time interval. First integer is not emitted immediately, but
// only after the first time interval has passed. The generated code will do a type
// conversion from int to interface{}.
func Interval(interval time.Duration) Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		i := 0
		runner := subscribeOn.ScheduleFutureRecursive(interval, func(self func(time.Duration)) {
			if subscriber.Subscribed() {
				observe(interface{}(i), nil, false)
				i++
				if subscriber.Subscribed() {
					self(interval)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Slice

type Slice = []interface{}

//jig:name ObservableObservable_ConcatAll

// ConcatAll flattens a higher order observable by concattenating the observables it emits.
func (o ObservableObservable) ConcatAll() Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		var concat struct {
			sync.Mutex
			observables	[]Observable
			observer	Observer
			subscriber	Subscriber
		}

		var source struct {
			observer	ObservableObserver
			subscriber	Subscriber
		}

		concat.observer = func(next interface{}, err error, done bool) {
			concat.Lock()
			if !done || err != nil {
				observe(next, err, done)
			} else {
				if len(concat.observables) == 0 {
					if !source.subscriber.Subscribed() {
						var zero interface{}
						observe(zero, nil, true)
					}
					concat.observables = nil
				} else {
					observable := concat.observables[0]
					concat.observables = concat.observables[1:]
					observable(concat.observer, subscribeOn, subscriber)
				}
			}
			concat.Unlock()
		}

		source.observer = func(next Observable, err error, done bool) {
			if !done {
				concat.Lock()
				initial := concat.observables == nil
				concat.observables = append(concat.observables, next)
				concat.Unlock()
				if initial {
					var zero interface{}
					concat.observer(zero, nil, true)
				}
			} else {
				concat.Lock()
				initial := concat.observables == nil
				source.subscriber.Done(err)
				concat.Unlock()
				if initial || err != nil {
					var zero interface{}
					concat.observer(zero, err, true)
				}
			}
		}
		source.subscriber = subscriber.Add()
		o(source.observer, subscribeOn, source.subscriber)
	}
	return observable
}

//jig:name ObservableSlice_Println

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableSlice) Println(a ...interface{}) error {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next Slice, err error, done bool) {
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
