// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package RefCount

import (
	"sync/atomic"
	"time"

	"github.com/reactivego/multicast"
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

//jig:name IntObserveFunc

// IntObserveFunc is the observer, a function that gets called whenever the
// observable has something to report. The next argument is the item value that
// is only valid when the done argument is false. When done is true and the err
// argument is not nil, then the observable has terminated with an error.
// When done is true and the err argument is nil, then the observable has
// completed normally.
type IntObserveFunc func(next int, err error, done bool)

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

//jig:name FromChanInt

// FromChanInt creates an ObservableInt from a Go channel of int values.
// It's not possible for the code feeding into the channel to send an error.
// The feeding code can send zero or more int items and then closing the
// channel will be seen as completion.
func FromChanInt(ch <-chan int) ObservableInt {
	return CreateInt(func(observer IntObserver) {
		for next := range ch {
			if observer.Closed() {
				return
			}
			observer.Next(next)
		}
		observer.Complete()
	})
}

//jig:name Interval

// Interval creates an ObservableInt that emits a sequence of integers spaced
// by a particular time interval.
func Interval(interval time.Duration) ObservableInt {
	return CreateInt(func(observer IntObserver) {
		for i := 0; ; i++ {
			time.Sleep(interval)
			if observer.Closed() {
				return
			}
			observer.Next(i)
		}
	})
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

//jig:name ObservableIntSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
func (o ObservableInt) Subscribe(observe IntObserveFunc, options ...SubscribeOption) Subscription {
	scheduler, subscriber := newSchedulerAndSubscriber(options)
	observer := func(next int, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			observe(zeroInt, err, true)
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	return subscriber
}

//jig:name ConnectableInt

// ConnectableInt is an ObservableInt that has an additional method Connect()
// used to Subscribe to the parent observable and then multicasting values to
// all subscribers of ConnectableInt.
type ConnectableInt struct {
	ObservableInt
	connect	func(options []SubscribeOption) Subscription
}

//jig:name ObservableIntMulticast

// Multicast converts an ordinary Observable into a connectable Observable.
// A connectable observable will only start emitting values after its Connect
// method has been called. The factory method passed in should return a
// new SubjectInt that implements the actual multicasting behavior.
func (o ObservableInt) Multicast(factory func() SubjectInt) ConnectableInt {
	const (
		active	int32	= iota
		notifying
		terminated
	)
	var subjectValue struct {
		state	int32
		atomic.Value
	}
	subjectValue.Store(factory())
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		if s, ok := subjectValue.Load().(SubjectInt); ok {
			s.ObservableInt(observe, subscribeOn, subscriber)
		}
	}
	observer := func(next int, err error, done bool) {
		if atomic.CompareAndSwapInt32(&subjectValue.state, active, notifying) {
			if s, ok := subjectValue.Load().(SubjectInt); ok {
				s.IntObserveFunc(next, err, done)
			}
			if !done {
				atomic.CompareAndSwapInt32(&subjectValue.state, notifying, active)
			} else {
				atomic.CompareAndSwapInt32(&subjectValue.state, notifying, terminated)
			}
		}
	}
	const (
		unsubscribed	int32	= iota
		subscribed
	)
	var subscriberValue struct {
		state	int32
		atomic.Value
	}
	connect := func(options []SubscribeOption) Subscription {
		if atomic.CompareAndSwapInt32(&subjectValue.state, terminated, active) {
			subjectValue.Store(factory())
		}
		if atomic.CompareAndSwapInt32(&subscriberValue.state, unsubscribed, subscribed) {
			scheduler := NewGoroutineScheduler()
			subscriber := subscriber.New()
			o.Subscribe(observer, SubscribeOn(scheduler, options...), WithSubscriber(subscriber))
			subscriberValue.Store(subscriber)
			subscriber.OnUnsubscribe(func() {
				atomic.CompareAndSwapInt32(&subscriberValue.state, subscribed, unsubscribed)
			})
		}
		subscription := subscriberValue.Load().(Subscriber)
		return subscription.Add(func() { subscription.Unsubscribe() })
	}
	return ConnectableInt{ObservableInt: observable, connect: connect}
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
	ch := multicast.NewChan(1, 16)

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

	return SubjectInt{observable.AsObservableInt(), observer}
}

//jig:name ObservableIntPublish

// Publish uses Multicast to control the subscription of a Subject to a
// source observable and turns the subject it into a connnectable observable.
// A Subject emits to an observer only those items that are emitted by
// the source Observable subsequent to the time of the subscription.
//
// If the source completed and as a result the internal Subject terminated, then
// calling Connect again will replace the old Subject with a newly created one.
// So this Publish operator is re-connectable, unlike the RxJS 5 behavior that
// isn't. To simulate the RxJS 5 behavior use Publish().AutoConnect(1) this will
// connect on the first subscription but will never re-connect.
func (o ObservableInt) Publish() ConnectableInt {
	return o.Multicast(NewSubjectInt)
}

//jig:name ObservableIntDo

// Do calls a function for each next value passing through the observable.
func (o ObservableInt) Do(f func(next int)) ObservableInt {
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			if !done {
				f(next)
			}
			observe(next, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
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

//jig:name ObservableIntSubscribeNext

// SubscribeNext operates upon the emissions from an Observable only.
// This method returns a Subscription.
func (o ObservableInt) SubscribeNext(f func(next int), options ...SubscribeOption) Subscription {
	return o.Subscribe(func(next int, err error, done bool) {
		if !done {
			f(next)
		}
	}, options...)
}

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

//jig:name ConnectableIntRefCount

// RefCount makes a ConnectableInt behave like an ordinary ObservableInt. On
// first Subscribe it will call Connect on its ConnectableInt and when its last
// subscriber is Unsubscribed it will cancel the connection by calling
// Unsubscribe on the subscription returned by the call to Connect.
func (o ConnectableInt) RefCount(options ...SubscribeOption) ObservableInt {
	var (
		refcount	int32
		connection	Subscription
	)
	observable := func(observe IntObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		if atomic.AddInt32(&refcount, 1) == 1 {
			connection = o.connect(options)
		}
		subscriber.OnUnsubscribe(func() {
			if atomic.AddInt32(&refcount, -1) == 0 {
				connection.Unsubscribe()
			}
		})
		o.ObservableInt(observe, subscribeOn, subscriber)
	}
	return observable
}

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
