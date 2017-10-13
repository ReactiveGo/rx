// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package test

import (
	"errors"
	"time"

	"github.com/reactivego/rx/channel"
	"github.com/reactivego/rx/schedulers"
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

//jig:name Range

// Range creates an ObservableInt that emits a range of sequential integers.
func Range(start, count int) ObservableInt {
	end := start + count
	return CreateInt(func(observer IntObserver) {
		for i := start; i < end; i++ {
			if observer.Closed() {
				return
			}
			observer.Next(i)
		}
		observer.Complete()
	})
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

//jig:name FromSlice

// FromSlice creates an Observable from a slice of interface{} values passed in.
func FromSlice(slice []interface{}) Observable {
	return Create(func(observer Observer) {
		for _, next := range slice {
			if observer.Closed() {
				return
			}
			observer.Next(next)
		}
		observer.Complete()
	})
}

//jig:name NextInt

// NextInt contains either the next int value (in .Next) or an error (in .Err).
// If Err is nil then Next must be valid. NextInt is meant to be used as the
// type of a channel allowing errors to be delivered in-band with the values.
type NextInt struct {
	Next	int
	Err	error
}

//jig:name FromInt

// FromInt creates an ObservableInt from multiple int values passed in.
func FromInt(slice ...int) ObservableInt {
	return FromSliceInt(slice)
}

//jig:name FromInts

// FromInts creates an ObservableInt from multiple int values passed in.
func FromInts(slice ...int) ObservableInt {
	return FromSliceInt(slice)
}

//jig:name EmptyInt

// EmptyInt creates an Observable that emits no items but terminates normally.
func EmptyInt() ObservableInt {
	return CreateInt(func(observer IntObserver) {
		observer.Complete()
	})
}

//jig:name SubjectInt

// SubjectInt is a combination of an observer and observable. Subjects are
// special because they are the only reactive constructs that support
// multicasting. The items sent to it through its observer side are
// multicasted to multiple clients subscribed to its observable side.
//
// A SubjectInt embeds ObservableInt and IntObserveFunc. This exposes the
// methods and fields of both types on SubjectInt. Use the ObservableInt
// methods to subscribe to it. Use the IntObserveFunc Next, Error and Complete
// methods to feed data to it.
//
// After a subject has been terminated by calling either Error or Complete,
// it goes into terminated state. All subsequent calls to its observer side
// will be silently ignored. All subsequent subscriptions to the observable
// side will be handled according to the specific behavior of the subject.
// There are different types of subjects, see the different NewXxxSubjectInt
// functions for more info.
//
// Important! a subject is a hot observable. This means that subscribing to
// it will block the calling goroutine while it is waiting for items and
// notifications to receive. Unless you have code on a different goroutine
// already feeding into the subject, your subscribe will deadlock.
// Alternatively, you could subscribe on a goroutine as shown in the example.
type SubjectInt struct {
	ObservableInt
	IntObserveFunc
}

//jig:name MaxReplayCapacity

// MaxReplayCapacity is the maximum size of a replay buffer. Can be modified.
var MaxReplayCapacity = 16383

//jig:name NewReplaySubjectInt

// NewReplaySubjectInt creates a new ReplaySubject. ReplaySubject ensures that
// all observers see the same sequence of emitted items, even if they
// subscribe after. When bufferCapacity argument is 0, then MaxReplayCapacity is
// used (currently 16383). When windowDuration argument is 0, then entries added
// to the buffer will remain fresh forever.
//
// Note that this implementation is non-blocking. When no subscribers are
// present the buffer fills up to bufferCapacity after which new items will
// start overwriting the oldest ones according to the FIFO principle.
// If a subscriber cannot keep up with the data rate of the source observable,
// eventually the buffer for the subscriber will overflow. At that moment the
// subscriber will receive an ErrMissingBackpressure error.
func NewReplaySubjectInt(bufferCapacity int, windowDuration time.Duration) SubjectInt {
	if bufferCapacity == 0 {
		bufferCapacity = MaxReplayCapacity
	}
	ch := channel.NewChan(bufferCapacity, 16)

	observable := Observable(func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		ep, err := ch.NewEndpoint(channel.ReplayAll)
		if err != nil {
			observe(nil, err, true)
			return
		}
		observable := Create(func(observer Observer) {
			receive := func(value interface{}, err error, closed bool) bool {
				if !closed {
					observer.Next(value)
				} else {
					observer.Error(err)
				}
				return !observer.Closed()
			}
			ep.Range(receive, windowDuration)
		})
		observable(observe, subscribeOn, subscriber.Add(ep.Cancel))
	})

	observer := func(next int, err error, done bool) {
		if !ch.Closed() {
			if !done {
				ch.Send(next)
			} else {
				ch.Close(err)
			}
		}
	}

	return SubjectInt{observable.AsInt(), observer}
}

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

//jig:name SubjectString

// SubjectString is a combination of an observer and observable. Subjects are
// special because they are the only reactive constructs that support
// multicasting. The items sent to it through its observer side are
// multicasted to multiple clients subscribed to its observable side.
//
// A SubjectString embeds ObservableString and StringObserveFunc. This exposes the
// methods and fields of both types on SubjectString. Use the ObservableString
// methods to subscribe to it. Use the StringObserveFunc Next, Error and Complete
// methods to feed data to it.
//
// After a subject has been terminated by calling either Error or Complete,
// it goes into terminated state. All subsequent calls to its observer side
// will be silently ignored. All subsequent subscriptions to the observable
// side will be handled according to the specific behavior of the subject.
// There are different types of subjects, see the different NewXxxSubjectString
// functions for more info.
//
// Important! a subject is a hot observable. This means that subscribing to
// it will block the calling goroutine while it is waiting for items and
// notifications to receive. Unless you have code on a different goroutine
// already feeding into the subject, your subscribe will deadlock.
// Alternatively, you could subscribe on a goroutine as shown in the example.
type SubjectString struct {
	ObservableString
	StringObserveFunc
}

//jig:name NewReplaySubjectString

// NewReplaySubjectString creates a new ReplaySubject. ReplaySubject ensures that
// all observers see the same sequence of emitted items, even if they
// subscribe after. When bufferCapacity argument is 0, then MaxReplayCapacity is
// used (currently 16383). When windowDuration argument is 0, then entries added
// to the buffer will remain fresh forever.
//
// Note that this implementation is non-blocking. When no subscribers are
// present the buffer fills up to bufferCapacity after which new items will
// start overwriting the oldest ones according to the FIFO principle.
// If a subscriber cannot keep up with the data rate of the source observable,
// eventually the buffer for the subscriber will overflow. At that moment the
// subscriber will receive an ErrMissingBackpressure error.
func NewReplaySubjectString(bufferCapacity int, windowDuration time.Duration) SubjectString {
	if bufferCapacity == 0 {
		bufferCapacity = MaxReplayCapacity
	}
	ch := channel.NewChan(bufferCapacity, 16)

	observable := Observable(func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		ep, err := ch.NewEndpoint(channel.ReplayAll)
		if err != nil {
			observe(nil, err, true)
			return
		}
		observable := Create(func(observer Observer) {
			receive := func(value interface{}, err error, closed bool) bool {
				if !closed {
					observer.Next(value)
				} else {
					observer.Error(err)
				}
				return !observer.Closed()
			}
			ep.Range(receive, windowDuration)
		})
		observable(observe, subscribeOn, subscriber.Add(ep.Cancel))
	})

	observer := func(next string, err error, done bool) {
		if !ch.Closed() {
			if !done {
				ch.Send(next)
			} else {
				ch.Close(err)
			}
		}
	}

	return SubjectString{observable.AsString(), observer}
}

//jig:name NewSubjectInt

// NewSubjectInt creates a new Subject. After the subject is
// terminated, all subsequent subscriptions to the observable side will be
// terminated immediately with either an Error or Complete notification send to
// the subscribing client
//
// Note that this implementation is blocking. When no subcribers are present
// then the data can flow freely. But when there are subscribers, the observable
// goroutine is blocked until all subscribers have processed the next, error or
// complete notification.
func NewSubjectInt() SubjectInt {
	ch := channel.NewChan(1, 16)

	observable := Observable(func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		ep, err := ch.NewEndpoint(0)
		if err != nil {
			observe(nil, err, true)
			return
		}
		observable := Create(func(observer Observer) {
			receive := func(value interface{}, err error, closed bool) bool {
				if !closed {
					observer.Next(value)
				} else {
					observer.Error(err)
				}
				return !observer.Closed()
			}
			ep.Range(receive, 0)
		})
		observable(observe, subscribeOn, subscriber.Add(ep.Cancel))
	})

	observer := func(next int, err error, done bool) {
		if !ch.Closed() {
			if !done {
				ch.FastSend(next)
			} else {
				ch.Close(err)
			}
		}
	}

	return SubjectInt{observable.AsInt(), observer}
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

//jig:name ObservableIntSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscriber.
func (o ObservableInt) Subscribe(observe IntObserveFunc, setters ...SubscribeOptionSetter) Subscriber {
	scheduler := NewTrampoline()
	setter := SubscribeOn(scheduler, setters...)
	options := NewSubscribeOptions(setter)
	subscriber := options.NewSubscriber()
	observer := func(next int, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroInt, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, options.SubscribeOn, subscriber)
	return subscriber
}

//jig:name ObservableIntSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscriber.
func (o ObservableInt) SubscribeNext(f func(next int), setters ...SubscribeOptionSetter) Subscription {
	return o.Subscribe(func(next int, err error, done bool) {
		if !done {
			f(next)
		}
	}, setters...)
}

//jig:name ObservableStringSubscribeOn

// SubscribeOn specifies the scheduler an ObservableString should use when it is
// subscribed to.
func (o ObservableString) SubscribeOn(subscribeOn Scheduler) ObservableString {
	observable := func(observe StringObserveFunc, _ Scheduler, subscriber Subscriber) {
		o(observe, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntWait

// Wait subscribes to the Observable and waits for completion or error.
// Returns either the error or nil when the Observable completed normally.
func (o ObservableInt) Wait(setters ...SubscribeOptionSetter) (e error) {
	o.Subscribe(func(next int, err error, done bool) {
		if done {
			e = err
		}
	}, setters...).Wait()
	return e
}

//jig:name ObservableIntToChanNext

// ToChanNext returns a channel that emits NextInt values. If the source
// observable does not emit values but emits complete, then the returned channel
// will close without emitting anything. If the source emitted an error, then
// that error is also emitted by the channel before closing immediately after.
// A NextInt has two fields Next (of type int) and Err (error). Valid items have
// either Next or Err set.
//
// Because the channel is fed by subscribing to the observable, ToChanNext would
// block when subscribed on the standard Trampoline scheduler which is initially
// synchronous. That's why the subscribing is done on the Goroutine scheduler.
// It is not possible to cancel the subscription created internally by ToChanNext.
func (o ObservableInt) ToChanNext(setters ...SubscribeOptionSetter) <-chan NextInt {
	scheduler := NewGoroutine()
	nextch := make(chan NextInt, 1)
	o.Subscribe(func(next int, err error, done bool) {
		if !done {
			nextch <- NextInt{Next: next}
		} else {
			if err != nil {
				nextch <- NextInt{Err: err}
			}
			close(nextch)
		}
	}, SubscribeOn(scheduler, setters...))
	return nextch
}

//jig:name ObservableIntToSingle

// ToSingle blocks until the ObservableInt emits exactly one value or an error.
// The value and any error are returned.
//
// This function subscribes to the source observable on the Goroutine scheduler.
// The Goroutine scheduler works in more situations for complex chains of
// observables, like when merging the output of multiple observables.
func (o ObservableInt) ToSingle(setters ...SubscribeOptionSetter) (v int, e error) {
	scheduler := NewGoroutine()
	o.Single().Subscribe(func(next int, err error, done bool) {
		if !done {
			v = next
		} else {
			e = err
		}
	}, SubscribeOn(scheduler, setters...)).Wait()
	return v, e
}

//jig:name ErrTypecastToInt

// ErrTypecastToInt is delivered to an observer if the generic value cannot be
// typecast to int.
var ErrTypecastToInt = errors.New("typecast to int failed")

//jig:name ObservableAsInt

// AsInt turns an Observable of interface{} into an ObservableInt. If during
// observing a typecast fails, the error ErrTypecastToInt will be emitted.
func (o Observable) AsInt() ObservableInt {
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

//jig:name ErrTypecastToString

// ErrTypecastToString is delivered to an observer if the generic value cannot be
// typecast to string.
var ErrTypecastToString = errors.New("typecast to string failed")

//jig:name ObservableAsString

// AsString turns an Observable of interface{} into an ObservableString. If during
// observing a typecast fails, the error ErrTypecastToString will be emitted.
func (o Observable) AsString() ObservableString {
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

//jig:name ObservableIntToChan

// ToChan returns a channel that emits int values. If the source observable does
// not emit values but emits an error or complete, then the returned channel
// will close without emitting any values.
//
// There is no way to determine whether the observable feeding into the
// channel terminated with an error or completed normally.
// Because the channel is fed by subscribing to the observable, ToChan would
// block when subscribed on the standard Trampoline scheduler which is initially
// synchronous. That's why the subscribing is done on the Goroutine scheduler.
// It is not possible to cancel the subscription created internally by ToChan.
func (o ObservableInt) ToChan(setters ...SubscribeOptionSetter) <-chan int {
	scheduler := NewGoroutine()
	nextch := make(chan int, 1)
	o.Subscribe(func(next int, err error, done bool) {
		if !done {
			nextch <- next
		} else {
			close(nextch)
		}
	}, SubscribeOn(scheduler, setters...))
	return nextch
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

//jig:name ObservableToChan

// ToChan returns a channel that emits interface{} values. If the source
// observable does not emit values but emits an error or complete, then the
// returned channel will enit any error and then close without emitting any
// values.
//
// Because the channel is fed by subscribing to the observable, ToChan would
// block when subscribed on the standard Trampoline scheduler which is initially
// synchronous. That's why the subscribing is done on the Goroutine scheduler.
//
// To cancel the subscription created internally by ToChan you will need access
// to the subscription used internnally by ToChan. To get at this subscription,
// pass the result of a call to option OnSubscribe(func(Subscription)) as a
// parameter to ToChan. On suscription the callback will be called with the
// subscription that was created.
func (o Observable) ToChan(setters ...SubscribeOptionSetter) <-chan interface{} {
	scheduler := NewGoroutine()
	nextch := make(chan interface{}, 1)
	o.Subscribe(func(next interface{}, err error, done bool) {
		if !done {
			nextch <- next
		} else {
			if err != nil {
				nextch <- err
			}
			close(nextch)
		}
	}, SubscribeOn(scheduler, setters...))
	return nextch
}

//jig:name ObservableIntToSlice

// ToSlice collects all values from the ObservableInt into an slice. The
// complete slice and any error are returned.
//
// This function subscribes to the source observable on the Goroutine scheduler.
// The Goroutine scheduler works in more situations for complex chains of
// observables, like when merging the output of multiple observables.
func (o ObservableInt) ToSlice(setters ...SubscribeOptionSetter) (a []int, e error) {
	scheduler := NewGoroutine()
	o.Subscribe(func(next int, err error, done bool) {
		if !done {
			a = append(a, next)
		} else {
			e = err
		}
	}, SubscribeOn(scheduler, setters...)).Wait()
	return a, e
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

//jig:name ObservableIntSingle

// Single enforces that the observableInt sends exactly one data item and then
// completes. If the observable sends no data before completing or sends more
// than 1 item before completing  this reported as an error to the observer.
func (o ObservableInt) Single() ObservableInt {
	return o.AsAny().Single().AsInt()
}

//jig:name ObservableIntAsAny

// AsAny turns a typed ObservableInt into an Observable of interface{}.
func (o ObservableInt) AsAny() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			observe(next, err, done)
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
