// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package Only

import (
	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler interface {
	Schedule(task func())
}

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

var zero interface{}

// Next is called by an Observable to emit the next interface{} value to the
// observer.
func (f ObserveFunc) Next(next interface{}) {
	f(next, nil, false)
}

// Error is called by an Observable to report an error to the observer.
func (f ObserveFunc) Error(err error) {
	f(zero, err, true)
}

// Complete is called by an Observable to signal that no more data is
// forthcoming to the observer.
func (f ObserveFunc) Complete() {
	f(zero, nil, true)
}

//jig:name Observable

// Observable is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type Observable func(ObserveFunc, Scheduler, Subscriber)

//jig:name Observer

// Observer is the interface used with Create when implementing a custom
// observable.
type Observer interface {
	// Next emits the next interface{} value.
	Next(interface{})
	// Error signals an error condition.
	Error(error)
	// Complete signals that no more data is to be expected.
	Complete()
	// Closed returns true when the subscription has been canceled.
	Closed() bool
}

//jig:name Create

// Create creates an Observable from scratch by calling observer methods
// programmatically.
func Create(f func(Observer)) Observable {
	observable := func(observe ObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		scheduler.Schedule(func() {
			if subscriber.Closed() {
				return
			}
			observer := func(next interface{}, err error, done bool) {
				if !subscriber.Closed() {
					observe(next, err, done)
				}
			}
			type ObserverSubscriber struct {
				ObserveFunc
				Subscriber
			}
			f(&ObserverSubscriber{observer, subscriber})
		})
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

var zeroString string

// Next is called by an ObservableString to emit the next string value to the
// observer.
func (f StringObserveFunc) Next(next string) {
	f(next, nil, false)
}

// Error is called by an ObservableString to report an error to the observer.
func (f StringObserveFunc) Error(err error) {
	f(zeroString, err, true)
}

// Complete is called by an ObservableString to signal that no more data is
// forthcoming to the observer.
func (f StringObserveFunc) Complete() {
	f(zeroString, nil, true)
}

//jig:name ObservableString

// ObservableString is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableString func(StringObserveFunc, Scheduler, Subscriber)

//jig:name ObservableOnlyString

// OnlyString filters the value stream of an Observable of interface{} and outputs only the
// string typed values.
func (o Observable) OnlyString() ObservableString {
	observable := func(observe StringObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextString, ok := next.(string); ok {
					observe(nextString, err, done)
				}
			} else {
				observe(zeroString, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name SizeObserveFunc

// SizeObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type SizeObserveFunc func(next Size, err error, done bool)

var zeroSize Size

// Next is called by an ObservableSize to emit the next Size value to the
// observer.
func (f SizeObserveFunc) Next(next Size) {
	f(next, nil, false)
}

// Error is called by an ObservableSize to report an error to the observer.
func (f SizeObserveFunc) Error(err error) {
	f(zeroSize, err, true)
}

// Complete is called by an ObservableSize to signal that no more data is
// forthcoming to the observer.
func (f SizeObserveFunc) Complete() {
	f(zeroSize, nil, true)
}

//jig:name ObservableSize

// ObservableSize is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableSize func(SizeObserveFunc, Scheduler, Subscriber)

//jig:name ObservableOnlySize

// OnlySize filters the value stream of an Observable of interface{} and outputs only the
// Size typed values.
func (o Observable) OnlySize() ObservableSize {
	observable := func(observe SizeObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextSize, ok := next.(Size); ok {
					observe(nextSize, err, done)
				}
			} else {
				observe(zeroSize, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name PointObserveFunc

// PointObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type PointObserveFunc func(next []point, err error, done bool)

var zeroPoint []point

// Next is called by an ObservablePoint to emit the next []point value to the
// observer.
func (f PointObserveFunc) Next(next []point) {
	f(next, nil, false)
}

// Error is called by an ObservablePoint to report an error to the observer.
func (f PointObserveFunc) Error(err error) {
	f(zeroPoint, err, true)
}

// Complete is called by an ObservablePoint to signal that no more data is
// forthcoming to the observer.
func (f PointObserveFunc) Complete() {
	f(zeroPoint, nil, true)
}

//jig:name ObservablePoint

// ObservablePoint is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservablePoint func(PointObserveFunc, Scheduler, Subscriber)

//jig:name ObservableOnlyPoint

// OnlyPoint filters the value stream of an Observable of interface{} and outputs only the
// []point typed values.
func (o Observable) OnlyPoint() ObservablePoint {
	observable := func(observe PointObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextPoint, ok := next.([]point); ok {
					observe(nextPoint, err, done)
				}
			} else {
				observe(zeroPoint, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name NewScheduler

func NewGoroutineScheduler() Scheduler	{ return scheduler.NewGoroutine }

func CurrentGoroutineScheduler() Scheduler	{ return scheduler.CurrentGoroutine }

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
	options := &subscribeOptions{scheduler: CurrentGoroutineScheduler()}
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

//jig:name ObservableStringSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscription.
func (o ObservableString) SubscribeNext(f func(next string), options ...SubscribeOption) Subscription {
	return o.Subscribe(func(next string, err error, done bool) {
		if !done {
			f(next)
		}
	}, options...)
}

//jig:name ObservableSizeSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
func (o ObservableSize) Subscribe(observe SizeObserveFunc, options ...SubscribeOption) Subscription {
	scheduler, subscriber := newSchedulerAndSubscriber(options)
	observer := func(next Size, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroSize, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	return subscriber
}

//jig:name ObservableSizeSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscription.
func (o ObservableSize) SubscribeNext(f func(next Size), options ...SubscribeOption) Subscription {
	return o.Subscribe(func(next Size, err error, done bool) {
		if !done {
			f(next)
		}
	}, options...)
}

//jig:name ObservablePointSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
func (o ObservablePoint) Subscribe(observe PointObserveFunc, options ...SubscribeOption) Subscription {
	scheduler, subscriber := newSchedulerAndSubscriber(options)
	observer := func(next []point, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroPoint, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	return subscriber
}

//jig:name ObservablePointSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscription.
func (o ObservablePoint) SubscribeNext(f func(next []point), options ...SubscribeOption) Subscription {
	return o.Subscribe(func(next []point, err error, done bool) {
		if !done {
			f(next)
		}
	}, options...)
}
