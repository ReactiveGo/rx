// Code generated by jig; DO NOT EDIT.

//go:generate jig

package SwitchMap

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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

//jig:name Never

// Never creates an Observable that emits no items and does't terminate.
func Never() Observable {
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
	}
	return observable
}

//jig:name StringObserver

// StringObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type StringObserver func(next string, err error, done bool)

//jig:name ObservableString

// ObservableString is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableString func(StringObserver, Scheduler, Subscriber)

//jig:name zero

var zero interface{}

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

//jig:name Interval

// Interval creates an ObservableInt that emits a sequence of integers spaced
// by a particular time interval. First integer is emitted after the first time
// interval expires.
func Interval(interval time.Duration) ObservableInt {
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		i := 0
		runner := scheduler.ScheduleFutureRecursive(interval, func(self func(time.Duration)) {
			if subscriber.Canceled() {
				return
			}
			observe(i, nil, false)
			if subscriber.Canceled() {
				return
			}
			i++
			self(interval)
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name zeroString

var zeroString string

//jig:name EmptyString

// EmptyString creates an Observable that emits no items but terminates normally.
func EmptyString() ObservableString {
	observable := func(observe StringObserver, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Subscribed() {
				observe(zeroString, nil, true)
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name ObservableSerialize

// Serialize forces an Observable to make serialized calls and to be
// well-behaved.
func (o Observable) Serialize() Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		var observer struct {
			sync.Mutex
			done	bool
		}
		serializer := func(next interface{}, err error, done bool) {
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

//jig:name ObservableTimeout

// ErrTimeout is delivered to an observer if the stream times out.
const ErrTimeout = RxError("timeout")

// Timeout mirrors the source Observable, but issues an error notification if a
// particular period of time elapses without any emitted items.
// Timeout schedules tasks on the scheduler passed to this
func (o Observable) Timeout(timeout time.Duration) Observable {
	observable := Observable(func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		if subscriber.Canceled() {
			return
		}
		var last struct {
			sync.Mutex
			at	time.Time
			done	bool
		}
		last.at = subscribeOn.Now()
		timer := func(self func(time.Duration)) {
			last.Lock()
			defer last.Unlock()
			if last.done || subscriber.Canceled() {
				return
			}
			deadline := last.at.Add(timeout)
			now := subscribeOn.Now()
			if now.Before(deadline) {
				self(deadline.Sub(now))
				return
			}
			last.done = true
			observe(zero, ErrTimeout, true)
		}
		runner := subscribeOn.ScheduleFutureRecursive(timeout, timer)
		subscriber.OnUnsubscribe(runner.Cancel)
		observer := func(next interface{}, err error, done bool) {
			last.Lock()
			defer last.Unlock()
			if last.done || subscriber.Canceled() {
				return
			}
			now := subscribeOn.Now()
			deadline := last.at.Add(timeout)
			if !now.Before(deadline) {
				return
			}
			last.done = done
			last.at = now
			observe(next, err, done)
		}
		o(observer, subscribeOn, subscriber)
	})
	return observable.Serialize()
}

//jig:name ErrTypecastToString

// ErrTypecastToString is delivered to an observer if the generic value cannot be
// typecast to string.
const ErrTypecastToString = RxError("typecast to string failed")

//jig:name ObservableAsObservableString

// AsString turns an Observable of interface{} into an ObservableString. If during
// observing a typecast fails, the error ErrTypecastToString will be emitted.
func (o Observable) AsObservableString() ObservableString {
	observable := func(observe StringObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextString, ok := next.(string); ok {
					observe(nextString, err, done)
				} else {
					observe(zeroString, ErrTypecastToString, true)
				}
			} else {
				observe(zeroString, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableTake

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

//jig:name ObservableIntTake

// Take emits only the first n items emitted by an ObservableInt.
func (o ObservableInt) Take(n int) ObservableInt {
	return o.AsObservable().Take(n).AsObservableInt()
}

//jig:name ObservableCatch

// Catch recovers from an error notification by continuing the sequence without
// emitting the error but by switching to the catch Observable to provide
// items.
func (o Observable) Catch(catch Observable) Observable {
	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if err != nil {
				catch(observe, subscribeOn, subscriber)
			} else {
				observe(next, err, done)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntSwitchMapString

// SwitchMapString transforms the items emitted by an ObservableInt by applying a
// function to each item an returning an ObservableString. In doing so, it behaves much like
// MergeMap (previously FlatMap), except that whenever a new ObservableString is emitted
// SwitchMap will unsubscribe from the previous ObservableString and begin emitting items
// from the newly emitted one.
func (o ObservableInt) SwitchMapString(project func(int) ObservableString) ObservableString {
	return o.MapObservableString(project).SwitchAll()
}

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

//jig:name Schedulers

func TrampolineScheduler() Scheduler {
	return scheduler.Trampoline
}

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

//jig:name ObservableStringPrintln

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println is performed on the Trampoline scheduler.
func (o ObservableString) Println() (err error) {
	subscriber := NewSubscriber()
	scheduler := TrampolineScheduler()
	observer := func(next string, e error, done bool) {
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

//jig:name ErrTypecastToInt

// ErrTypecastToInt is delivered to an observer if the generic value cannot be
// typecast to int.
const ErrTypecastToInt = RxError("typecast to int failed")

//jig:name zeroInt

var zeroInt int

//jig:name ObservableAsObservableInt

// AsInt turns an Observable of interface{} into an ObservableInt. If during
// observing a typecast fails, the error ErrTypecastToInt will be emitted.
func (o Observable) AsObservableInt() ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextInt, ok := next.(int); ok {
					observe(nextInt, err, done)
				} else {
					observe(zeroInt, ErrTypecastToInt, true)
				}
			} else {
				observe(zeroInt, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntMapObservableString

// MapObservableString transforms the items emitted by an ObservableInt by applying a
// function to each item.
func (o ObservableInt) MapObservableString(project func(int) ObservableString) ObservableObservableString {
	observable := func(observe ObservableStringObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			var mapped ObservableString
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableStringObserver

// ObservableStringObserver is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type ObservableStringObserver func(next ObservableString, err error, done bool)

//jig:name ObservableObservableString

// ObservableObservableString is a function taking an Observer, Scheduler and Subscriber.
// Calling it will subscribe the Observer to events from the Observable.
type ObservableObservableString func(ObservableStringObserver, Scheduler, Subscriber)

//jig:name LinkEnums

// state
const (
	linkUnsubscribed	= iota
	linkSubscribing
	linkIdle
	linkBusy
	linkError	// done:error
	linkCanceled	// externally:canceled
	linkCompleting
	linkComplete	// done:complete
)

// callbackState
const (
	callbackNil	= iota
	settingCallback
	callbackSet
)

// callbackKind
const (
	linkCallbackOnComplete	= iota
	linkCancelOrCompleted
)

//jig:name linkString

type linkStringObserver func(*linkString, string, error, bool)

type linkString struct {
	observe		linkStringObserver
	state		int32
	callbackState	int32
	callbackKind	int
	callback	func()
	subscriber	Subscriber
}

func newInitialLinkString() *linkString {
	return &linkString{state: linkCompleting, subscriber: subscriber.New()}
}

func newLinkString(observe linkStringObserver, subscriber Subscriber) *linkString {
	return &linkString{
		observe:	observe,
		subscriber:	subscriber.Add(),
	}
}

func (o *linkString) Observe(next string, err error, done bool) error {
	if !atomic.CompareAndSwapInt32(&o.state, linkIdle, linkBusy) {
		if atomic.LoadInt32(&o.state) > linkBusy {
			return RxError("Already Done")
		}
		return RxError("Recursion Error")
	}
	o.observe(o, next, err, done)
	if done {
		if err != nil {
			if !atomic.CompareAndSwapInt32(&o.state, linkBusy, linkError) {
				return RxError("Internal Error: 'busy' -> 'error'")
			}
		} else {
			if !atomic.CompareAndSwapInt32(&o.state, linkBusy, linkCompleting) {
				return RxError("Internal Error: 'busy' -> 'completing'")
			}
		}
	} else {
		if !atomic.CompareAndSwapInt32(&o.state, linkBusy, linkIdle) {
			return RxError("Internal Error: 'busy' -> 'idle'")
		}
	}
	if atomic.LoadInt32(&o.callbackState) != callbackSet {
		return nil
	}
	if atomic.CompareAndSwapInt32(&o.state, linkCompleting, linkComplete) {
		o.callback()
	}
	if o.callbackKind == linkCancelOrCompleted {
		if atomic.CompareAndSwapInt32(&o.state, linkIdle, linkCanceled) {
			o.callback()
		}
	}
	return nil
}

func (o *linkString) SubscribeTo(observable ObservableString, scheduler Scheduler) error {
	if !atomic.CompareAndSwapInt32(&o.state, linkUnsubscribed, linkSubscribing) {
		return RxError("Already Subscribed")
	}
	observer := func(next string, err error, done bool) {
		o.Observe(next, err, done)
	}
	observable(observer, scheduler, o.subscriber)
	if !atomic.CompareAndSwapInt32(&o.state, linkSubscribing, linkIdle) {
		return RxError("Internal Error")
	}
	return nil
}

func (o *linkString) Cancel(callback func()) error {
	if !atomic.CompareAndSwapInt32(&o.callbackState, callbackNil, settingCallback) {
		return RxError("Already Waiting")
	}
	o.callbackKind = linkCancelOrCompleted
	o.callback = callback
	if !atomic.CompareAndSwapInt32(&o.callbackState, settingCallback, callbackSet) {
		return RxError("Internal Error")
	}
	o.subscriber.Unsubscribe()
	if atomic.CompareAndSwapInt32(&o.state, linkCompleting, linkComplete) {
		o.callback()
	}
	if atomic.CompareAndSwapInt32(&o.state, linkIdle, linkCanceled) {
		o.callback()
	}
	return nil
}

func (o *linkString) OnComplete(callback func()) error {
	if !atomic.CompareAndSwapInt32(&o.callbackState, callbackNil, settingCallback) {
		return RxError("Already Waiting")
	}
	o.callbackKind = linkCallbackOnComplete
	o.callback = callback
	if !atomic.CompareAndSwapInt32(&o.callbackState, settingCallback, callbackSet) {
		return RxError("Internal Error")
	}
	if atomic.CompareAndSwapInt32(&o.state, linkCompleting, linkComplete) {
		o.callback()
	}
	return nil
}

//jig:name ObservableObservableStringSwitchAll

// SwitchAll converts an Observable that emits Observables into a single Observable
// that emits the items emitted by the most-recently-emitted of those Observables.
func (o ObservableObservableString) SwitchAll() ObservableString {
	observable := func(observe StringObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(link *linkString, next string, err error, done bool) {
			if !done || err != nil {
				observe(next, err, done)
			} else {
				link.subscriber.Unsubscribe()
			}
		}
		currentLink := newInitialLinkString()
		var switcherMutex sync.Mutex
		switcherSubscriber := subscriber.Add()
		switcher := func(next ObservableString, err error, done bool) {
			switch {
			case !done:
				previousLink := currentLink
				func() {
					switcherMutex.Lock()
					defer switcherMutex.Unlock()
					currentLink = newLinkString(observer, subscriber)
				}()
				previousLink.Cancel(func() {
					switcherMutex.Lock()
					defer switcherMutex.Unlock()
					currentLink.SubscribeTo(next, subscribeOn)
				})
			case err != nil:
				currentLink.Cancel(func() {
					observe(zeroString, err, true)
				})
				switcherSubscriber.Unsubscribe()
			default:
				currentLink.OnComplete(func() {
					observe(zeroString, nil, true)
				})
				switcherSubscriber.Unsubscribe()
			}
		}
		o(switcher, subscribeOn, switcherSubscriber)
	}
	return observable
}
