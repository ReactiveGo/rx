// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package SwitchAll

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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

//jig:name Interval

// Interval creates an ObservableInt that emits a sequence of integers spaced
// by a particular time interval.
func Interval(interval time.Duration) ObservableInt {
	observable := func(observe IntObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		i := 0
		scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Canceled() {
				return
			}
			time.Sleep(interval)
			if subscriber.Canceled() {
				return
			}
			observe(i, nil, false)
			if subscriber.Canceled() {
				return
			}
			i++
			self()
		})
	}
	return observable
}

//jig:name ObservableTake

// Take emits only the first n items emitted by an Observable.
func (o Observable) Take(n int) Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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

//jig:name ObservableIntMapObservableInt

// MapObservableInt transforms the items emitted by an ObservableInt by applying a
// function to each item.
func (o ObservableInt) MapObservableInt(project func(int) ObservableInt) ObservableObservableInt {
	observable := func(observe ObservableIntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			var mapped ObservableInt
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntObserveFunc

// ObservableIntObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type ObservableIntObserveFunc func(next ObservableInt, err error, done bool)

//jig:name zeroObservableInt

var zeroObservableInt ObservableInt

//jig:name ObservableObservableInt

// ObservableObservableInt is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableObservableInt func(ObservableIntObserveFunc, Scheduler, Subscriber)

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name ErrTypecastToInt

// ErrTypecastToInt is delivered to an observer if the generic value cannot be
// typecast to int.
const ErrTypecastToInt = RxError("typecast to int failed")

//jig:name ObservableAsObservableInt

// AsInt turns an Observable of interface{} into an ObservableInt. If during
// observing a typecast fails, the error ErrTypecastToInt will be emitted.
func (o Observable) AsObservableInt() ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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

//jig:name linkInt

type linkIntObserveFunc func(*linkInt, int, error, bool)

type linkInt struct {
	observe		linkIntObserveFunc
	state		int32
	callbackState	int32
	callbackKind	int
	callback	func()
	subscriber	Subscriber
}

func newInitialLinkInt() *linkInt {
	return &linkInt{state: linkCompleting, subscriber: subscriber.New()}
}

func newLinkInt(observe linkIntObserveFunc, subscriber Subscriber) *linkInt {
	return &linkInt{
		observe:	observe,
		subscriber:	subscriber.AddChild(),
	}
}

func (o *linkInt) Observe(next int, err error, done bool) error {
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

func (o *linkInt) SubscribeTo(observable ObservableInt, scheduler Scheduler) error {
	if !atomic.CompareAndSwapInt32(&o.state, linkUnsubscribed, linkSubscribing) {
		return RxError("Already Subscribed")
	}
	observer := func(next int, err error, done bool) {
		o.Observe(next, err, done)
	}
	observable(observer, scheduler, o.subscriber)
	if !atomic.CompareAndSwapInt32(&o.state, linkSubscribing, linkIdle) {
		return RxError("Internal Error")
	}
	return nil
}

func (o *linkInt) Cancel(callback func()) error {
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

func (o *linkInt) OnComplete(callback func()) error {
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

//jig:name ObservableObservableIntSwitchAll

// SwitchAll converts an Observable that emits Observables into a single Observable
// that emits the items emitted by the most-recently-emitted of those Observables.
func (o ObservableObservableInt) SwitchAll() ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(link *linkInt, next int, err error, done bool) {
			if !done || err != nil {
				observe(next, err, done)
			} else {
				link.subscriber.Unsubscribe()
			}
		}
		currentLink := newInitialLinkInt()
		var switcherMutex sync.Mutex
		switcherSubscriber := subscriber.AddChild()
		switcher := func(next ObservableInt, err error, done bool) {
			switch {
			case !done:
				previousLink := currentLink
				func() {
					switcherMutex.Lock()
					defer switcherMutex.Unlock()
					currentLink = newLinkInt(observer, subscriber)
				}()
				previousLink.Cancel(func() {
					switcherMutex.Lock()
					defer switcherMutex.Unlock()
					currentLink.SubscribeTo(next, subscribeOn)
				})
			case err != nil:
				currentLink.Cancel(func() {
					observe(zeroInt, err, true)
				})
				switcherSubscriber.Unsubscribe()
			default:
				currentLink.OnComplete(func() {
					observe(zeroInt, nil, true)
				})
				switcherSubscriber.Unsubscribe()
			}
		}
		o(switcher, subscribeOn, switcherSubscriber)
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
func (o ObservableInt) Println() (err error) {
	subscriber := subscriber.New()
	scheduler := TrampolineScheduler()
	observer := func(next int, e error, done bool) {
		if !done {
			fmt.Println(next)
		} else {
			err = e
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	subscriber.Wait()
	return
}

//jig:name ObservableIntSubscribeOn

// SubscribeOn specifies the scheduler an ObservableInt should use when it is
// subscribed to.
func (o ObservableInt) SubscribeOn(subscribeOn Scheduler) ObservableInt {
	observable := func(observe IntObserveFunc, _ Scheduler, subscriber Subscriber) {
		o(observe, subscribeOn, subscriber)
	}
	return observable
}
