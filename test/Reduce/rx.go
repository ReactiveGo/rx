// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package Reduce

import (
	"errors"

	"github.com/reactivego/rx/schedulers"
	"github.com/reactivego/rx/subscriber"
)

//jig:name IntObserveFunc

// IntObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type IntObserveFunc func(int, error, bool)

var zeroInt int

// Next is called by an ObservableInt to emit the next int value to the
// observer.
func (f IntObserveFunc) Next(next int) {
	f(next, nil, false)
}

// Error is called by an ObservableInt to report an error to the observer.
func (f IntObserveFunc) Error(err error) {
	f(zeroInt, err, true)
}

// Complete is called by an ObservableInt to signal that no more data is
// forthcoming to the observer.
func (f IntObserveFunc) Complete() {
	f(zeroInt, nil, true)
}

//jig:name ObservableInt

// ObservableInt is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableInt func(IntObserveFunc, Scheduler, Subscriber)

//jig:name IntObserver

// IntObserver is the interface used with CreateInt when implementing a custom
// observable.
type IntObserver interface {
	// Next emits the next int value.
	Next(int)
	// Error signals an error condition.
	Error(error)
	// Complete signals that no more data is to be expected.
	Complete()
	// Closed returns true when the subscription has been canceled.
	Closed() bool
}

//jig:name CreateInt

// CreateInt creates an Observable from scratch by calling observer methods
// programmatically.
func CreateInt(f func(IntObserver)) ObservableInt {
	observable := func(observe IntObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		scheduler.Schedule(func() {
			if subscriber.Closed() {
				return
			}
			observer := func(next int, err error, done bool) {
				if !subscriber.Closed() {
					observe(next, err, done)
				}
			}
			type ObserverSubscriber struct {
				IntObserveFunc
				Subscriber
			}
			f(&ObserverSubscriber{observer, subscriber})
		})
	}
	return observable
}

//jig:name FromSliceInt

// FromSliceInt creates an ObservableInt from a slice of int values passed in.
func FromSliceInt(slice []int) ObservableInt {
	return CreateInt(func(observer IntObserver) {
		for _, next := range slice {
			if observer.Closed() {
				return
			}
			observer.Next(next)
		}
		observer.Complete()
	})
}

//jig:name FromInts

// FromInts creates an ObservableInt from multiple int values passed in.
func FromInts(slice ...int) ObservableInt {
	return FromSliceInt(slice)
}

//jig:name FromInt

// FromInt creates an ObservableInt from multiple int values passed in.
func FromInt(slice ...int) ObservableInt {
	return FromSliceInt(slice)
}

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler interface {
	Schedule(task func())
}

//jig:name Subscriber

// Subscriber is an alias for the subscriber.Subscriber interface type.
type Subscriber subscriber.Subscriber

//jig:name ObservableIntReduceFloat32

// ReduceFloat32 applies a reducer function to each item emitted by an ObservableInt
// and the previous reducer result. The operator accepts a seed argument that
// is passed to the reducer for the first item emitted by the ObservableInt.
// ReduceFloat32 emits only the final value.
func (o ObservableInt) ReduceFloat32(reducer func(float32, int) float32, seed float32) ObservableFloat32 {
	observable := func(observe Float32ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		state := seed
		observer := func(next int, err error, done bool) {
			if !done {
				state = reducer(state, next)
			} else {
				if err == nil {
					observe(state, nil, false)
				}
				observe(zeroFloat32, err, done)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntReduce

// Reduce applies a reducer function to each item emitted by an ObservableInt
// and the previous reducer result. The operator accepts a seed argument that
// is passed to the reducer for the first item emitted by the ObservableInt.
// Reduce emits only the final value.
func (o ObservableInt) Reduce(reducer func(interface{}, int) interface{}, seed interface{}) Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		state := seed
		observer := func(next int, err error, done bool) {
			if !done {
				state = reducer(state, next)
			} else {
				if err == nil {
					observe(state, nil, false)
				}
				observe(zero, err, done)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name Float32ObserveFunc

// Float32ObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type Float32ObserveFunc func(float32, error, bool)

var zeroFloat32 float32

// Next is called by an ObservableFloat32 to emit the next float32 value to the
// observer.
func (f Float32ObserveFunc) Next(next float32) {
	f(next, nil, false)
}

// Error is called by an ObservableFloat32 to report an error to the observer.
func (f Float32ObserveFunc) Error(err error) {
	f(zeroFloat32, err, true)
}

// Complete is called by an ObservableFloat32 to signal that no more data is
// forthcoming to the observer.
func (f Float32ObserveFunc) Complete() {
	f(zeroFloat32, nil, true)
}

//jig:name ObservableFloat32

// ObservableFloat32 is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableFloat32 func(Float32ObserveFunc, Scheduler, Subscriber)

//jig:name ObserveFunc

// ObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type ObserveFunc func(interface{}, error, bool)

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

//jig:name NewScheduler

func NewGoroutine() Scheduler	{ return &schedulers.Goroutine{} }

func NewTrampoline() Scheduler	{ return &schedulers.Trampoline{} }

//jig:name SubscribeOptions

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription subscriber.Subscription

// SubscribeOptions is a struct with options for Subscribe related methods.
type SubscribeOptions struct {
	// SubscribeOn is the scheduler to run the observable subscription on.
	SubscribeOn	Scheduler
	// OnSubscribe is called right after the subscription is created and before
	// subscribing continues further.
	OnSubscribe	func(subscription Subscription)
	// OnUnsubscribe is called by the subscription to notify the client that the
	// subscription has been canceled.
	OnUnsubscribe	func()
}

// NewSubscriber will return a newly created subscriber. Before returning the
// subscription the OnSubscribe callback (if set) will already have been called.
func (options SubscribeOptions) NewSubscriber() Subscriber {
	subscription := subscriber.NewWithCallback(options.OnUnsubscribe)
	if options.OnSubscribe != nil {
		options.OnSubscribe(subscription)
	}
	return subscription
}

// SubscribeOptionSetter is the type of a function for setting SubscribeOptions.
type SubscribeOptionSetter func(options *SubscribeOptions)

// SubscribeOn takes the scheduler to run the observable subscription on and
// additional setters. It will first set the SubscribeOn option before
// calling the other setters provided as a parameter.
func SubscribeOn(subscribeOn Scheduler, setters ...SubscribeOptionSetter) SubscribeOptionSetter {
	return func(options *SubscribeOptions) {
		options.SubscribeOn = subscribeOn
		for _, setter := range setters {
			setter(options)
		}
	}
}

// OnSubscribe takes a callback to be called on subscription.
func OnSubscribe(callback func(Subscription)) SubscribeOptionSetter {
	return func(options *SubscribeOptions) { options.OnSubscribe = callback }
}

// OnUnsubscribe takes a callback to be called on subscription cancelation.
func OnUnsubscribe(callback func()) SubscribeOptionSetter {
	return func(options *SubscribeOptions) { options.OnUnsubscribe = callback }
}

// NewSubscribeOptions will create a new SubscribeOptions struct and then call
// the setter on it to recursively set all the options. It then returns a
// pointer to the created SubscribeOptions struct.
func NewSubscribeOptions(setter SubscribeOptionSetter) *SubscribeOptions {
	options := &SubscribeOptions{}
	setter(options)
	return options
}

//jig:name ObservableFloat32Subscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscriber.
func (o ObservableFloat32) Subscribe(observe Float32ObserveFunc, setters ...SubscribeOptionSetter) Subscriber {
	scheduler := NewTrampoline()
	setter := SubscribeOn(scheduler, setters...)
	options := NewSubscribeOptions(setter)
	subscriber := options.NewSubscriber()
	observer := func(next float32, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroFloat32, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, options.SubscribeOn, subscriber)
	return subscriber
}

//jig:name ObservableFloat32SubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscriber.
func (o ObservableFloat32) SubscribeNext(f func(next float32), setters ...SubscribeOptionSetter) Subscription {
	return o.Subscribe(func(next float32, err error, done bool) {
		if !done {
			f(next)
		}
	}, setters...)
}

//jig:name ObservableSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscriber.
func (o Observable) Subscribe(observe ObserveFunc, setters ...SubscribeOptionSetter) Subscriber {
	scheduler := NewTrampoline()
	setter := SubscribeOn(scheduler, setters...)
	options := NewSubscribeOptions(setter)
	subscriber := options.NewSubscriber()
	observer := func(next interface{}, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zero, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, options.SubscribeOn, subscriber)
	return subscriber
}

//jig:name ObservableSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscriber.
func (o Observable) SubscribeNext(f func(next interface{}), setters ...SubscribeOptionSetter) Subscription {
	return o.Subscribe(func(next interface{}, err error, done bool) {
		if !done {
			f(next)
		}
	}, setters...)
}

//jig:name ObservableFloat32ToSingle

// ToSingle blocks until the ObservableFloat32 emits exactly one value or an error.
// The value and any error are returned.
//
// This function subscribes to the source observable on the Goroutine scheduler.
// The Goroutine scheduler works in more situations for complex chains of
// observables, like when merging the output of multiple observables.
func (o ObservableFloat32) ToSingle(setters ...SubscribeOptionSetter) (v float32, e error) {
	scheduler := NewGoroutine()
	o.Single().Subscribe(func(next float32, err error, done bool) {
		if !done {
			v = next
		} else {
			e = err
		}
	}, SubscribeOn(scheduler, setters...)).Wait()
	return v, e
}

//jig:name ObservableFloat32Single

// Single enforces that the observableFloat32 sends exactly one data item and then
// completes. If the observable sends no data before completing or sends more
// than 1 item before completing  this reported as an error to the observer.
func (o ObservableFloat32) Single() ObservableFloat32 {
	return o.AsObservable().Single().AsObservableFloat32()
}

//jig:name ObservableFloat32AsObservable

// AsObservable turns a typed ObservableFloat32 into an Observable of interface{}.
func (o ObservableFloat32) AsObservable() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next float32, err error, done bool) {
			observe(interface{}(next), err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableSingle

// Single enforces that the observable sends exactly one data item and then
// completes. If the observable sends no data before completing or sends more
// than 1 item before completing  this reported as an error to the observer.
func (o Observable) Single() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		var (
			count	int
			latest	interface{}
		)
		observer := func(next interface{}, err error, done bool) {
			if count < 2 {
				if done {
					if err != nil {
						observe(nil, err, true)
					} else {
						if count == 1 {
							observe(latest, nil, false)
							observe(nil, nil, true)
						} else {
							observe(nil, errors.New("expected one value, got none"), true)
						}
					}
				} else {
					count++
					if count == 1 {
						latest = next
					} else {
						observe(nil, errors.New("expected one value, got multiple"), true)
					}
				}
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ErrTypecastToFloat32

// ErrTypecastToFloat32 is delivered to an observer if the generic value cannot be
// typecast to float32.
var ErrTypecastToFloat32 = errors.New("typecast to float32 failed")

//jig:name ObservableAsObservableFloat32

// AsFloat32 turns an Observable of interface{} into an ObservableFloat32. If during
// observing a typecast fails, the error ErrTypecastToFloat32 will be emitted.
func (o Observable) AsObservableFloat32() ObservableFloat32 {
	observable := func(observe Float32ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextFloat32, ok := next.(float32); ok {
					observe(nextFloat32, err, done)
				} else {
					observe(zeroFloat32, ErrTypecastToFloat32, true)
				}
			} else {
				observe(zeroFloat32, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}
