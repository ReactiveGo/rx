// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package SwitchMap

import (
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

//jig:name Never

// Never creates an Observable that emits no items and does't terminate.
func Never() Observable {
	observable := func(observe ObserveFunc, scheduler Scheduler, subscriber Subscriber) {
	}
	return observable
}

//jig:name StringObserveFunc

// StringObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type StringObserveFunc func(next string, err error, done bool)

//jig:name zeroString

var zeroString string

//jig:name ObservableString

// ObservableString is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableString func(StringObserveFunc, Scheduler, Subscriber)

//jig:name FromSlice

// FromSlice creates an Observable from a slice of interface{} values passed in.
func FromSlice(slice []interface{}) Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		i := 0
		subscribeOn.ScheduleRecursive(func(self func()) {
			if !subscriber.Canceled() {
				if i < len(slice) {
					observe(slice[i], nil, false)
					if !subscriber.Canceled() {
						i++
						self()
					}
				} else {
					observe(zero, nil, true)
				}
			}
		})
	}
	return observable
}

//jig:name From

// From creates an Observable from multiple interface{} values passed in.
func From(slice ...interface{}) Observable {
	return FromSlice(slice)
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

//jig:name EmptyString

// EmptyString creates an Observable that emits no items but terminates normally.
func EmptyString() ObservableString {
	observable := func(observe StringObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		subscribeOn.Schedule(func() {
			if !subscriber.Canceled() {
				observe(zeroString, nil, true)
			}
		})
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
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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
//
// This observer starts a goroutine for every subscription to monitor the
// timeout deadline. It is guaranteed that calls to the observer for this
// subscription will never be called concurrently. It is however almost certain
// that any timeout error will be delivered on a goroutine other than the one
// delivering the next values.
func (o Observable) Timeout(timeout time.Duration) Observable {
	observable := Observable(func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		deadline := time.NewTimer(timeout)
		unsubscribe := make(chan struct{})
		observer := func(next interface{}, err error, done bool) {
			if deadline.Stop() {
				if subscriber.Closed() {
					return
				}
				observe(next, err, done)
				if done {
					return
				}
				deadline.Reset(timeout)
			}
		}
		timeouter := func() {
			select {
			case <-deadline.C:
				if subscriber.Closed() {
					return
				}
				observe(nil, ErrTimeout, true)
			case <-unsubscribe:
			}
		}
		go timeouter()
		o(observer, subscribeOn, subscriber.Add(func() { close(unsubscribe) }))
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
	observable := func(observe StringObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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

//jig:name ObservableCatch

// Catch recovers from an error notification by continuing the sequence without
// emitting the error but by switching to the catch Observable to provide
// items.
func (o Observable) Catch(catch Observable) Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			observe(interface{}(next), err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Schedulers

func TrampolineScheduler() Scheduler	{ return scheduler.Trampoline }

func GoroutineScheduler() Scheduler	{ return scheduler.Goroutine }

//jig:name SubscribeOption

// SubscribeOption is an option that can be passed to the Subscribe method.
type SubscribeOption func(options *subscribeOptions)

type subscribeOptions struct {
	scheduler	Scheduler
	subscriber	Subscriber
	onSubscribe	func(subscription Subscription)
	onUnsubscribe	func()
}

// SubscribeOn returns an option that can be passed to the Subscribe method.
// It takes the scheduler to subscribe the observable on. The tasks that
// actually perform the observable functionality are scheduled on this
// scheduler. The other options that can be passed here are applied after the
// scheduler was set so any schedulers passed in via other will override
// the scheduler passed here.
func SubscribeOn(scheduler Scheduler, other ...SubscribeOption) SubscribeOption {
	return func(options *subscribeOptions) {
		options.scheduler = scheduler
		for _, setter := range other {
			setter(options)
		}
	}
}

// WithSubscriber returns an option that can be passed to the Subscribe method.
// The Subscribe method will use the subscriber passed here instead of creating
// a new one.
func WithSubscriber(subscriber Subscriber) SubscribeOption {
	return func(options *subscribeOptions) {
		options.subscriber = subscriber
	}
}

// OnSubscribe returns an option that can be passed to the Subscribe method.
// It takes a callback that is called from the Subscribe method just before
// subscribing continues further.
func OnSubscribe(callback func(Subscription)) SubscribeOption {
	return func(options *subscribeOptions) { options.onSubscribe = callback }
}

// OnUnsubscribe returns an option that can be passed to the Subscribe method.
// It takes a callback that is called by the Subscribe method to notify the
// client that the subscription has been canceled.
func OnUnsubscribe(callback func()) SubscribeOption {
	return func(options *subscribeOptions) { options.onUnsubscribe = callback }
}

// newSchedulerAndSubscriber will return either return the scheduler and subscriber
// passed in through the SubscribeOn() and WithSubscriber() options or it will
// return newly created scheduler and subscriber. Before returning the callback
// passed in through OnSubscribe() will already have been called.
func newSchedulerAndSubscriber(setters []SubscribeOption) (Scheduler, Subscriber) {
	options := &subscribeOptions{scheduler: TrampolineScheduler()}
	for _, setter := range setters {
		setter(options)
	}
	if options.subscriber == nil {
		options.subscriber = subscriber.New()
	}
	options.subscriber.OnUnsubscribe(options.onUnsubscribe)
	if options.onSubscribe != nil {
		options.onSubscribe(options.subscriber)
	}
	return options.scheduler, options.subscriber
}

//jig:name ObservableStringSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
func (o ObservableString) Subscribe(observe StringObserveFunc, options ...SubscribeOption) Subscription {
	scheduler, subscriber := newSchedulerAndSubscriber(options)
	observer := func(next string, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroString, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	return subscriber
}

//jig:name ObservableStringToSlice

// ToSlice collects all values from the ObservableString into an slice. The
// complete slice and any error are returned.
//
// This function subscribes to the source observable on the Goroutine
// scheduler. The Goroutine scheduler works in more situations for
// complex chains of observables, like when merging the output of multiple
// observables.
func (o ObservableString) ToSlice(options ...SubscribeOption) (slice []string, err error) {
	scheduler := GoroutineScheduler()
	o.Subscribe(func(next string, e error, done bool) {
		if !done {
			slice = append(slice, next)
		} else {
			err = e
		}
	}, SubscribeOn(scheduler, options...)).Wait()
	return
}

//jig:name ObservableIntMapObservableString

// MapObservableString transforms the items emitted by an ObservableInt by applying a
// function to each item.
func (o ObservableInt) MapObservableString(project func(int) ObservableString) ObservableObservableString {
	observable := func(observe ObservableStringObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
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

//jig:name ObservableStringObserveFunc

// ObservableStringObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type ObservableStringObserveFunc func(next ObservableString, err error, done bool)

//jig:name zeroObservableString

var zeroObservableString ObservableString

//jig:name ObservableObservableString

// ObservableObservableString is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableObservableString func(ObservableStringObserveFunc, Scheduler, Subscriber)

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

type linkStringObserveFunc func(*linkString, string, error, bool)

type linkString struct {
	observe		linkStringObserveFunc
	state		int32
	callbackState	int32
	callbackKind	int
	callback	func()
	subscriber	Subscriber
}

func newInitialLinkString() *linkString {
	return &linkString{state: linkCompleting, subscriber: subscriber.New()}
}

func newLinkString(observe linkStringObserveFunc, subscriber Subscriber) *linkString {
	return &linkString{
		observe:	observe,
		subscriber:	subscriber.AddChild(),
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
	observable := func(observe StringObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(link *linkString, next string, err error, done bool) {
			if !done || err != nil {
				observe(next, err, done)
			} else {
				link.subscriber.Unsubscribe()
			}
		}
		currentLink := newInitialLinkString()
		var switcherMutex sync.Mutex
		switcherSubscriber := subscriber.AddChild()
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
