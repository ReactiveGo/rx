// Code generated by jig; DO NOT EDIT.

//go:generate jig

package PublishReplay

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/reactivego/multicast"
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

//jig:name FromChanInt

// FromChanInt creates an ObservableInt from a Go channel of int values.
// It's not possible for the code feeding into the channel to send an error.
// The feeding code can send nil or more int items and then closing the
// channel will be seen as completion.
func FromChanInt(ch <-chan int) ObservableInt {
	var zeroInt int
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Canceled() {
				return
			}
			next, ok := <-ch
			if subscriber.Canceled() {
				return
			}
			if ok {
				observe(next, nil, false)
				if subscriber.Subscribed() {
					self()
				}
			} else {
				observe(zeroInt, nil, true)
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name Subscription

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription = subscriber.Subscription

//jig:name ObservableIntSubscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe by default is performed on the Trampoline scheduler.
func (o ObservableInt) Subscribe(observe IntObserver, subscribers ...Subscriber) Subscription {
	subscribers = append(subscribers, subscriber.New())
	scheduler := scheduler.Trampoline
	observer := func(next int, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			var zeroInt int
			observe(zeroInt, err, true)
			subscribers[0].Unsubscribe()
		}
	}
	subscribers[0].OnWait(scheduler.Wait)
	o(observer, scheduler, subscribers[0])
	return subscribers[0]
}

//jig:name ConnectableInt

// ConnectableInt is an ObservableInt that has an additional method Connect()
// used to Subscribe to the parent observable and then multicasting values to
// all subscribers of ConnectableInt.
type ConnectableInt struct {
	ObservableInt
	connect	func() Subscription
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
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		if s, ok := subjectValue.Load().(SubjectInt); ok {
			s.ObservableInt(observe, subscribeOn, subscriber)
		}
	}
	observer := func(next int, err error, done bool) {
		if atomic.CompareAndSwapInt32(&subjectValue.state, active, notifying) {
			if s, ok := subjectValue.Load().(SubjectInt); ok {
				s.IntObserver(next, err, done)
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
	connect := func() Subscription {
		if atomic.CompareAndSwapInt32(&subjectValue.state, terminated, active) {
			subjectValue.Store(factory())
		}
		if atomic.CompareAndSwapInt32(&subscriberValue.state, unsubscribed, subscribed) {
			scheduler := GoroutineScheduler()
			subscriber := subscriber.New()
			o.SubscribeOn(scheduler).Subscribe(observer, subscriber)
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
// A SubjectInt embeds ObservableInt and IntObserver. This exposes the
// methods and fields of both types on SubjectInt. Use the ObservableInt
// methods to subscribe to it. Use the IntObserver Next, Error and Complete
// methods to feed data to it.
//
// After a subject has been terminated by calling either Error or Complete,
// it goes into terminated state. All subsequent calls to its observer side
// will be silently ignored. All subsequent subscriptions to the observable
// side will be handled according to the specific behavior of the subject.
// There are different types of subjects, see the different NewXxxSubjectInt
// functions for more info.
type SubjectInt struct {
	ObservableInt
	IntObserver
}

// Next is called by an ObservableInt to emit the next int value to the
// Observer.
func (f IntObserver) Next(next int) {
	f(next, nil, false)
}

// Error is called by an ObservableInt to report an error to the Observer.
func (f IntObserver) Error(err error) {
	var zeroInt int
	f(zeroInt, err, true)
}

// Complete is called by an ObservableInt to signal that no more data is
// forthcoming to the Observer.
func (f IntObserver) Complete() {
	var zeroInt int
	f(zeroInt, nil, true)
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
func NewReplaySubjectInt(bufferCapacity int, windowDuration time.Duration) SubjectInt {
	if bufferCapacity == 0 {
		bufferCapacity = MaxReplayCapacity
	}
	ch := multicast.NewChan(bufferCapacity, 16)

	observable := Observable(func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		ep, err := ch.NewEndpoint(multicast.ReplayAll)
		if err != nil {
			observe(nil, err, true)
			return
		}
		observable := Create(func(Next Next, Error Error, Complete Complete, Canceled Canceled) {
			receive := func(next interface{}, err error, closed bool) bool {
				switch {
				case !closed:
					Next(next)
				case err != nil:
					Error(err)
				default:
					Complete()
				}
				return !Canceled()
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

	return SubjectInt{observable.AsObservableInt(), observer}
}

//jig:name ObservableIntPublishReplay

// Replay uses Multicast to control the subscription of a ReplaySubject to a
// source observable and turns the subject into a connectable observable.
// A ReplaySubject emits to any observer all of the items that were emitted by
// the source observable, regardless of when the observer subscribes.
//
// If the source completed and as a result the internal ReplaySubject
// terminated, then calling Connect again will replace the old ReplaySubject
// with a newly created one.
func (o ObservableInt) PublishReplay(bufferCapacity int, windowDuration time.Duration) ConnectableInt {
	factory := func() SubjectInt {
		return NewReplaySubjectInt(bufferCapacity, windowDuration)
	}
	return o.Multicast(factory)
}

//jig:name ConnectableIntConnect

// Connect instructs a connectable Observable to begin emitting items to its
// subscribers. All values will then be passed on to the observers that
// subscribed to this connectable observable
func (c ConnectableInt) Connect() Subscription {
	return c.connect()
}

//jig:name ObservableIntPrintln

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println is performed on the Trampoline scheduler.
func (o ObservableInt) Println() (err error) {
	subscriber := subscriber.New()
	scheduler := scheduler.Trampoline
	observer := func(next int, e error, done bool) {
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

//jig:name ConnectableIntAutoConnect

// AutoConnect makes a ConnectableInt behave like an ordinary ObservableInt that
// automatically connects when the specified number of clients subscribe to it.
// If count is 0, then AutoConnect will immediately call connect on the
// ConnectableInt before returning the ObservableInt part of the ConnectableInt.
func (o ConnectableInt) AutoConnect(count int) ObservableInt {
	if count == 0 {
		o.connect()
		return o.ObservableInt
	}
	var refcount int32
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		if atomic.AddInt32(&refcount, 1) == int32(count) {
			o.connect()
		}
		o.ObservableInt(observe, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntWait

// Wait subscribes to the Observable and waits for completion or error.
// Returns either the error or nil when the Observable completed normally.
// Subscribing is performed on the Trampoline scheduler.
func (o ObservableInt) Wait() (err error) {
	subscriber := subscriber.New()
	scheduler := scheduler.Trampoline
	observer := func(next int, e error, done bool) {
		if done {
			err = e
			subscriber.Unsubscribe()
		}
	}
	subscriber.OnWait(scheduler.Wait)
	o(observer, scheduler, subscriber)
	subscriber.Wait()
	return
}

//jig:name GoroutineScheduler

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

//jig:name ObservableIntSubscribeOn

// SubscribeOn specifies the scheduler an ObservableInt should use when it is
// subscribed to.
func (o ObservableInt) SubscribeOn(subscribeOn Scheduler) ObservableInt {
	observable := func(observe IntObserver, _ Scheduler, subscriber Subscriber) {
		subscriber.OnWait(subscribeOn.Wait)
		o(observe, subscribeOn, subscriber)
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

//jig:name Error

// Error signals an error condition.
type Error func(error)

//jig:name Complete

// Complete signals that no more data is to be expected.
type Complete func()

//jig:name Canceled

// Canceled returns true when the observer has unsubscribed.
type Canceled func() bool

//jig:name Next

// Next can be called to emit the next value to the IntObserver.
type Next func(interface{})

//jig:name Create

// Create provides a way of creating an Observable from
// scratch by calling observer methods programmatically.
//
// The create function provided to Create will be called once
// to implement the observable. It is provided with a Next, Error,
// Complete and Canceled function that can be called by the code that
// implements the Observable.
func Create(create func(Next, Error, Complete, Canceled)) Observable {
	var zero interface{}
	observable := func(observe Observer, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Canceled() {
				return
			}
			n := func(next interface{}) {
				if subscriber.Subscribed() {
					observe(next, nil, false)
				}
			}
			e := func(err error) {
				if subscriber.Subscribed() {
					observe(zero, err, true)
				}
			}
			c := func() {
				if subscriber.Subscribed() {
					observe(zero, nil, true)
				}
			}
			x := func() bool {
				return subscriber.Canceled()
			}
			create(n, e, c, x)
		})
		subscriber.OnUnsubscribe(runner.Cancel)
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
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next interface{}, err error, done bool) {
			if !done {
				if nextInt, ok := next.(int); ok {
					observe(nextInt, err, done)
				} else {
					var zeroInt int
					observe(zeroInt, ErrTypecastToInt, true)
				}
			} else {
				var zeroInt int
				observe(zeroInt, err, true)
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}
