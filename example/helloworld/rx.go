// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package main

import (
	"github.com/reactivego/rx/schedulers"
	"github.com/reactivego/rx/subscriber"
)

//jig:name StringObserveFunc

// StringObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type StringObserveFunc func(string, error, bool)

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

//jig:name StringObserver

// StringObserver is the interface used with CreateString when implementing a custom
// observable.
type StringObserver interface {
	// Next emits the next string value.
	Next(string)
	// Error signals an error condition.
	Error(error)
	// Complete signals that no more data is to be expected.
	Complete()
	// Closed returns true when the subscription has been canceled.
	Closed() bool
}

//jig:name CreateString

// CreateString creates an Observable from scratch by calling observer methods
// programmatically.
func CreateString(f func(StringObserver)) ObservableString {
	observable := func(observe StringObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		scheduler.Schedule(func() {
			if subscriber.Closed() {
				return
			}
			observer := func(next string, err error, done bool) {
				if !subscriber.Closed() {
					observe(next, err, done)
				}
			}
			type ObserverSubscriber struct {
				StringObserveFunc
				Subscriber
			}
			f(&ObserverSubscriber{observer, subscriber})
		})
	}
	return observable
}

//jig:name FromSliceString

// FromSliceString creates an ObservableString from a slice of string values passed in.
func FromSliceString(slice []string) ObservableString {
	return CreateString(func(observer StringObserver) {
		for _, next := range slice {
			if observer.Closed() {
				return
			}
			observer.Next(next)
		}
		observer.Complete()
	})
}

//jig:name FromStrings

// FromStrings creates an ObservableString from multiple string values passed in.
func FromStrings(slice ...string) ObservableString {
	return FromSliceString(slice)
}

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler interface {
	Schedule(task func())
}

//jig:name Subscriber

// Subscriber is an alias for the subscriber.Subscriber interface type.
type Subscriber subscriber.Subscriber

//jig:name ObservableStringMapString

// MapString transforms the items emitted by an ObservableString by applying a
// function to each item.
func (o ObservableString) MapString(project func(string) string) ObservableString {
	observable := func(observe StringObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next string, err error, done bool) {
			var mapped string
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

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

//jig:name ObservableStringSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscriber.
func (o ObservableString) Subscribe(observe StringObserveFunc, setters ...SubscribeOptionSetter) Subscriber {
	scheduler := NewTrampoline()
	setter := SubscribeOn(scheduler, setters...)
	options := NewSubscribeOptions(setter)
	subscriber := options.NewSubscriber()
	observer := func(next string, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroString, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, options.SubscribeOn, subscriber)
	return subscriber
}

//jig:name ObservableStringSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscriber.
func (o ObservableString) SubscribeNext(f func(next string), setters ...SubscribeOptionSetter) Subscription {
	return o.Subscribe(func(next string, err error, done bool) {
		if !done {
			f(next)
		}
	}, setters...)
}
