// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package rx

import (
	"sync"
	"sync/atomic"
)

//jig:name BarObserveFunc

// BarObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type BarObserveFunc func(bar, error, bool)

var zeroBar bar

// Next is called by an ObservableBar to emit the next bar value to the
// observer.
func (f BarObserveFunc) Next(next bar) {
	f(next, nil, false)
}

// Error is called by an ObservableBar to report an error to the observer.
func (f BarObserveFunc) Error(err error) {
	f(zeroBar, err, true)
}

// Complete is called by an ObservableBar to signal that no more data is
// forthcoming to the observer.
func (f BarObserveFunc) Complete() {
	f(zeroBar, nil, true)
}

//jig:name ObservableBar

// ObservableBar is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableBar func(BarObserveFunc, Scheduler, Subscriber)

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

//jig:name ObservableFooObserveFunc

// ObservableFooObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type ObservableFooObserveFunc func(ObservableFoo, error, bool)

var zeroObservableFoo ObservableFoo

// Next is called by an ObservableObservableFoo to emit the next ObservableFoo value to the
// observer.
func (f ObservableFooObserveFunc) Next(next ObservableFoo) {
	f(next, nil, false)
}

// Error is called by an ObservableObservableFoo to report an error to the observer.
func (f ObservableFooObserveFunc) Error(err error) {
	f(zeroObservableFoo, err, true)
}

// Complete is called by an ObservableObservableFoo to signal that no more data is
// forthcoming to the observer.
func (f ObservableFooObserveFunc) Complete() {
	f(zeroObservableFoo, nil, true)
}

//jig:name ObservableObservableFoo

// ObservableObservableFoo is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableObservableFoo func(ObservableFooObserveFunc, Scheduler, Subscriber)

//jig:name ObservableFooAsObservable

// AsObservable turns a typed ObservableFoo into an Observable of interface{}.
func (o ObservableFoo) AsObservable() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next foo, err error, done bool) {
			observe(interface{}(next), err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableFooMapObservableBar

// MapObservableBar transforms the items emitted by an ObservableFoo by applying a
// function to each item.
func (o ObservableFoo) MapObservableBar(project func(foo) ObservableBar) ObservableObservableBar {
	observable := func(observe ObservableBarObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next foo, err error, done bool) {
			var mapped ObservableBar
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

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

//jig:name Empty

// Empty creates an Observable that emits no items but terminates normally.
func Empty() Observable {
	return Create(func(observer Observer) {
		observer.Complete()
	})
}

//jig:name ObservableBarObserveFunc

// ObservableBarObserveFunc is essentially the observer, a function that gets called
// whenever the observable has something to report.
type ObservableBarObserveFunc func(ObservableBar, error, bool)

var zeroObservableBar ObservableBar

// Next is called by an ObservableObservableBar to emit the next ObservableBar value to the
// observer.
func (f ObservableBarObserveFunc) Next(next ObservableBar) {
	f(next, nil, false)
}

// Error is called by an ObservableObservableBar to report an error to the observer.
func (f ObservableBarObserveFunc) Error(err error) {
	f(zeroObservableBar, err, true)
}

// Complete is called by an ObservableObservableBar to signal that no more data is
// forthcoming to the observer.
func (f ObservableBarObserveFunc) Complete() {
	f(zeroObservableBar, nil, true)
}

//jig:name ObservableObservableBar

// ObservableObservableBar is essentially a subscribe function taking an observe
// function, scheduler and an subscriber.
type ObservableObservableBar func(ObservableBarObserveFunc, Scheduler, Subscriber)

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

//jig:name ObservableSerialize

// Serialize forces an Observable to make serialized calls and to be
// well-behaved.
func (o Observable) Serialize() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		var (
			mutex		sync.Mutex
			alreadyDone	bool
		)
		observer := func(next interface{}, err error, done bool) {
			mutex.Lock()
			if !alreadyDone {
				alreadyDone = done
				observe(next, err, done)
			}
			mutex.Unlock()
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableObservableBarMergeAll

// MergeAll flattens a higher order observable by merging the observables it emits.
func (o ObservableObservableBar) MergeAll() ObservableBar {
	observable := func(observe BarObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		var (
			mutex	sync.Mutex
			count	int32	= 1
		)
		observer := func(next bar, err error, done bool) {
			mutex.Lock()
			defer mutex.Unlock()
			if !done || err != nil {
				observe(next, err, done)
			} else {
				if atomic.AddInt32(&count, -1) == 0 {
					observe(zeroBar, nil, true)
				}
			}
		}
		merger := func(next ObservableBar, err error, done bool) {
			if !done {
				atomic.AddInt32(&count, 1)
				next(observer, subscribeOn, subscriber)
			} else {
				observer(zeroBar, err, true)
			}
		}
		o(merger, subscribeOn, subscriber)
	}
	return observable
}
