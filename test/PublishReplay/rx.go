// Code generated by jig; DO NOT EDIT.

//go:generate jig

package PublishReplay

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

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
			if !subscriber.Subscribed() {
				return
			}
			next, ok := <-ch
			if !subscriber.Subscribed() {
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

//jig:name GoroutineScheduler

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

//jig:name Subscription

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription = subscriber.Subscription

//jig:name ObservableInt_Subscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableInt) Subscribe(observe IntObserver, subscribers ...Subscriber) Subscription {
	subscribers = append(subscribers, subscriber.New())
	scheduler := scheduler.MakeTrampoline()
	observer := func(next int, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			var zeroInt int
			observe(zeroInt, err, true)
			subscribers[0].Done(err)
		}
	}
	subscribers[0].OnWait(scheduler.Wait)
	o(observer, scheduler, subscribers[0])
	return subscribers[0]
}

//jig:name Connectable

// Connectable provides the Connect method for a Multicaster.
type Connectable func(Scheduler, Subscriber)

// Connect instructs a multicaster to subscribe to its source and begin
// multicasting items to its subscribers. Connect accepts an optional
// scheduler argument.
func (c Connectable) Connect(schedulers ...Scheduler) Subscription {
	subscriber := subscriber.New()
	schedulers = append(schedulers, scheduler.MakeTrampoline())
	if !schedulers[0].IsConcurrent() {
		subscriber.OnWait(schedulers[0].Wait)
	}
	c(schedulers[0], subscriber)
	return subscriber
}

//jig:name IntMulticaster

// IntMulticaster is a multicasting connectable observable. One or more
// IntObservers can subscribe to it simultaneously. It will subscribe to the
// source ObservableInt when Connect is called. After that, every emission
// from the source is multcast to all subscribed IntObservers.
type IntMulticaster struct {
	ObservableInt
	Connectable
}

//jig:name ObservableInt_Multicast

// Multicast converts an ordinary observable into a multicasting connectable
// observable or multicaster for short. A multicaster will only start emitting
// values after its Connect method has been called. The factory method passed
// in should return a new SubjectInt that implements the actual multicasting
// behavior.
func (o ObservableInt) Multicast(factory func() SubjectInt) IntMulticaster {
	const (
		active	int32	= iota
		notifying
		erred
		completed
	)
	var subject struct {
		state	int32
		atomic.Value
		count	int32
	}
	const (
		unsubscribed	int32	= iota
		subscribed
	)
	var source struct {
		sync.Mutex
		state		int32
		subscriber	Subscriber
	}
	subject.Store(factory())
	observer := func(next int, err error, done bool) {
		if atomic.CompareAndSwapInt32(&subject.state, active, notifying) {
			if s, ok := subject.Load().(SubjectInt); ok {
				s.IntObserver(next, err, done)
			}
			switch {
			case !done:
				atomic.CompareAndSwapInt32(&subject.state, notifying, active)
			case err != nil:
				if atomic.CompareAndSwapInt32(&subject.state, notifying, erred) {
					source.subscriber.Done(err)
				}
			default:
				if atomic.CompareAndSwapInt32(&subject.state, notifying, completed) {
					source.subscriber.Done(nil)
				}
			}
		}
	}
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		if atomic.AddInt32(&subject.count, 1) == 1 {
			if atomic.CompareAndSwapInt32(&subject.state, erred, active) {
				subject.Store(factory())
			}
		}
		if s, ok := subject.Load().(SubjectInt); ok {
			s.ObservableInt(observe, subscribeOn, subscriber)
		}
		subscriber.OnUnsubscribe(func() {
			atomic.AddInt32(&subject.count, -1)
		})
	}
	connectable := func(subscribeOn Scheduler, subscriber Subscriber) {
		source.Lock()
		if atomic.CompareAndSwapInt32(&source.state, unsubscribed, subscribed) {
			source.subscriber = subscriber
			o(observer, subscribeOn, subscriber)
			subscriber.OnUnsubscribe(func() {
				atomic.CompareAndSwapInt32(&source.state, subscribed, unsubscribed)
			})
		} else {
			source.subscriber.OnUnsubscribe(subscriber.Unsubscribe)
			subscriber.OnUnsubscribe(source.subscriber.Unsubscribe)
		}
		source.Unlock()
	}
	return IntMulticaster{ObservableInt: observable, Connectable: connectable}
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

//jig:name MakeObserverObservable

const ErrOutOfSubscriptions = RxError("out of subscriptions")

// MakeObserverObservable actually does make an observer observable. It
// creates a buffering multicaster and returns both the Observer and the
// Observable side of it. These are then used as the core of any Subject
// implementation. The Observer side is used to pass items into the buffering
// multicaster. This then multicasts the items to every Observer that
// subscribes to the returned Observable.
//
//	age     age below which items are kept to replay to a new subscriber.
//	length  length of the item buffer, number of items kept to replay to a new subscriber.
//	[cap]   Capacity of the item buffer, number of items that can be observed before blocking.
//	[scap]  Capacity of the subscription list, max number of simultaneous subscribers.
func MakeObserverObservable(age time.Duration, length int, capacity ...int) (Observer, Observable) {
	const (
		ms	= time.Millisecond
		us	= time.Microsecond
	)

	// Access to subscriptions
	const (
		idle	uint32	= iota
		busy
	)

	// State of subscription and buffer
	const (
		active	uint64	= iota
		canceled
		closing
		closed
	)

	const (
		// Cursor is parked so it does not influence advancing the commit index.
		parked uint64 = math.MaxUint64
	)

	type subscription struct {
		Cursor		uint64
		State		uint64		// active, canceled, closed
		LastActive	time.Time	// track activity to deterime backoff
	}

	type subscriptions struct {
		sync.Mutex
		*sync.Cond
		entries	[]subscription
		access	uint32	// idle, busy
	}

	type item struct {
		Value	interface{}
		At	time.Time
	}

	type buffer struct {
		age	time.Duration
		len	uint64
		cap	uint64

		mod	uint64
		items	[]item
		begin	uint64
		end	uint64
		commit	uint64
		state	uint64	// active, closed

		subscriptions	subscriptions

		err	error
	}

	make := func(age time.Duration, length int, capacity ...int) *buffer {
		cap, scap := uint64(length), uint64(32)
		switch {
		case len(capacity) >= 2:
			cap, scap = uint64(capacity[0]), uint64(capacity[1])
		case len(capacity) == 1:
			cap = uint64(capacity[0])
		}
		len := uint64(length)
		if cap < len {
			cap = len
		}
		cap = uint64(1) << uint(math.Ceil(math.Log2(float64(cap))))
		buf := &buffer{
			len:	len,
			cap:	cap,
			age:	age,
			items:	make([]item, cap),
			mod:	cap - 1,
			end:	cap,
			subscriptions: subscriptions{
				entries: make([]subscription, 0, scap),
			},
		}
		buf.subscriptions.Cond = sync.NewCond(&buf.subscriptions.Mutex)
		return buf
	}
	buf := make(age, length, capacity...)

	accessSubscriptions := func(access func([]subscription)) bool {
		spun := false
		for !atomic.CompareAndSwapUint32(&buf.subscriptions.access, idle, busy) {
			runtime.Gosched()
			spun = true
		}
		access(buf.subscriptions.entries)
		atomic.StoreUint32(&buf.subscriptions.access, idle)
		return spun
	}

	send := func(value interface{}) {
		for buf.commit == buf.end {
			slowest := parked
			spun := accessSubscriptions(func(subscriptions []subscription) {
				for i := range subscriptions {
					cursor := atomic.LoadUint64(&subscriptions[i].Cursor)
					if cursor < slowest {
						slowest = cursor
					}
				}
				if atomic.LoadUint64(&buf.begin) < slowest && slowest <= atomic.LoadUint64(&buf.end) {
					atomic.StoreUint64(&buf.begin, slowest)
					atomic.StoreUint64(&buf.end, slowest+buf.mod+1)
				} else {
					slowest = parked
				}
			})
			if slowest == parked {

				if !spun {

					runtime.Gosched()
				}
				if atomic.LoadUint64(&buf.state) != active {
					return
				}
			}
		}
		buf.items[buf.commit&buf.mod] = item{Value: value, At: time.Now()}
		atomic.AddUint64(&buf.commit, 1)
		buf.subscriptions.Broadcast()
	}

	close := func(err error) {
		if atomic.CompareAndSwapUint64(&buf.state, active, closing) {
			buf.err = err
			if atomic.CompareAndSwapUint64(&buf.state, closing, closed) {
				accessSubscriptions(func(subscriptions []subscription) {
					for i := range subscriptions {
						atomic.CompareAndSwapUint64(&subscriptions[i].State, active, closed)
					}
				})
			}
		}
		buf.subscriptions.Broadcast()
	}

	observer := func(next interface{}, err error, done bool) {
		if atomic.LoadUint64(&buf.state) == active {
			if !done {
				send(next)
			} else {
				close(err)
			}
		}
	}

	appendSubscription := func(cursor uint64) (sub *subscription, err error) {
		accessSubscriptions(func([]subscription) {
			s := &buf.subscriptions
			if len(s.entries) < cap(s.entries) {
				s.entries = append(s.entries, subscription{Cursor: cursor})
				sub = &s.entries[len(s.entries)-1]
				return
			}
			for i := range s.entries {
				sub = &s.entries[i]
				if atomic.CompareAndSwapUint64(&sub.Cursor, parked, cursor) {
					return
				}
			}
			err = ErrOutOfSubscriptions
			return
		})
		return
	}

	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		begin := atomic.LoadUint64(&buf.begin)
		sub, err := appendSubscription(begin)
		if err != nil {
			runner := subscribeOn.Schedule(func() {
				if subscriber.Subscribed() {
					observe(nil, err, true)
				}
			})
			subscriber.OnUnsubscribe(runner.Cancel)
			return
		}
		commit := atomic.LoadUint64(&buf.commit)
		if begin+buf.len < commit {
			atomic.StoreUint64(&sub.Cursor, commit-buf.len)
		}
		atomic.StoreUint64(&sub.State, atomic.LoadUint64(&buf.state))
		sub.LastActive = time.Now()

		receiver := subscribeOn.ScheduleFutureRecursive(0, func(self func(time.Duration)) {
			commit := atomic.LoadUint64(&buf.commit)

			if sub.Cursor == commit {
				if atomic.CompareAndSwapUint64(&sub.State, canceled, canceled) {

					atomic.StoreUint64(&sub.Cursor, parked)
					return
				} else {

					now := time.Now()
					if now.Before(sub.LastActive.Add(1 * ms)) {

						self(50 * us)
						return
					} else if now.Before(sub.LastActive.Add(250 * ms)) {
						if atomic.CompareAndSwapUint64(&sub.State, closed, closed) {

							observe(nil, buf.err, true)
							atomic.StoreUint64(&sub.Cursor, parked)
							return
						}

						self(500 * us)
						return
					} else {
						if subscribeOn.IsConcurrent() {

							buf.subscriptions.Lock()
							buf.subscriptions.Wait()
							buf.subscriptions.Unlock()
							sub.LastActive = time.Now()
							self(0)
							return
						} else {

							self(5 * ms)
							return
						}
					}
				}
			}

			if atomic.LoadUint64(&sub.State) == canceled {
				atomic.StoreUint64(&sub.Cursor, parked)
				return
			}
			for ; sub.Cursor != commit; atomic.AddUint64(&sub.Cursor, 1) {
				item := &buf.items[sub.Cursor&buf.mod]
				if buf.age == 0 || item.At.IsZero() || time.Since(item.At) < buf.age {
					observe(item.Value, nil, false)
				}
				if atomic.LoadUint64(&sub.State) == canceled {
					atomic.StoreUint64(&sub.Cursor, parked)
					return
				}
			}

			sub.LastActive = time.Now()
			self(0)
		})
		subscriber.OnUnsubscribe(receiver.Cancel)

		subscriber.OnUnsubscribe(func() {
			atomic.CompareAndSwapUint64(&sub.State, active, canceled)
			buf.subscriptions.Broadcast()
		})
	}
	return observer, observable
}

//jig:name SubjectInt

// SubjectInt is a combination of an IntObserver and ObservableInt.
// Subjects are special because they are the only reactive constructs that
// support multicasting. The items sent to it through its observer side are
// multicasted to multiple clients subscribed to its observable side.
//
// The SubjectInt exposes all methods from the embedded IntObserver and
// ObservableInt. Use the IntObserver Next, Error and Complete methods to feed
// data to it. Use the ObservableInt methods to subscribe to it.
//
// After a subject has been terminated by calling either Error or Complete,
// it goes into terminated state. All subsequent calls to its observer side
// will be silently ignored. All subsequent subscriptions to the observable
// side will be handled according to the specific behavior of the subject.
// There are different types of subjects, see the different NewXxxSubjectInt
// functions for more info.
type SubjectInt struct {
	IntObserver
	ObservableInt
}

// Next is called by an ObservableInt to emit the next int value to the
// Observer.
func (o IntObserver) Next(next int) {
	o(next, nil, false)
}

// Error is called by an ObservableInt to report an error to the Observer.
func (o IntObserver) Error(err error) {
	var zero int
	o(zero, err, true)
}

// Complete is called by an ObservableInt to signal that no more data is
// forthcoming to the Observer.
func (o IntObserver) Complete() {
	var zero int
	o(zero, nil, true)
}

//jig:name Observer_AsIntObserver

// AsIntObserver converts an observer of interface{} items to an observer of
// int items.
func (o Observer) AsIntObserver() IntObserver {
	observer := func(next int, err error, done bool) {
		o(next, err, done)
	}
	return observer
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
	observer, observable := MakeObserverObservable(windowDuration, bufferCapacity)
	return SubjectInt{observer.AsIntObserver(), observable.AsObservableInt()}
}

//jig:name ObservableInt_PublishReplay

// Replay uses the Multicast operator to control the subscription of a
// ReplaySubject to a source observable and turns the subject into a
// connectable observable. A ReplaySubject emits to any observer all of the
// items that were emitted by the source observable, regardless of when the
// observer subscribes.
//
// If the source completed and as a result the internal ReplaySubject
// terminated, then calling Connect again will replace the old ReplaySubject
// with a newly created one.
func (o ObservableInt) PublishReplay(bufferCapacity int, windowDuration time.Duration) IntMulticaster {
	factory := func() SubjectInt {
		return NewReplaySubjectInt(bufferCapacity, windowDuration)
	}
	return o.Multicast(factory)
}

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name ObservableInt_Println

// Println subscribes to the Observable and prints every item to os.Stdout
// while it waits for completion or error. Returns either the error or nil
// when the Observable completed normally.
// Println uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableInt) Println(a ...interface{}) error {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next int, err error, done bool) {
		if !done {
			fmt.Println(append(a, next)...)
		} else {
			subscriber.Done(err)
		}
	}
	subscriber.OnWait(scheduler.Wait)
	o(observer, scheduler, subscriber)
	return subscriber.Wait()
}

//jig:name ThrowInt

// ThrowInt creates an Observable that emits no items and terminates with an
// error.
func ThrowInt(err error) ObservableInt {
	var zeroInt int
	observable := func(observe IntObserver, scheduler Scheduler, subscriber Subscriber) {
		runner := scheduler.Schedule(func() {
			if subscriber.Subscribed() {
				observe(zeroInt, err, true)
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

//jig:name ErrAutoConnect

const ErrAutoConnectInvalidCount = RxError("invalid count")

//jig:name IntMulticaster_AutoConnect

// AutoConnect makes a IntMulticaster behave like an ordinary ObservableInt
// that automatically connects the multicaster to its source when the
// specified number of observers have subscribed to it. If the count is less
// than 1 it will return a ThrowInt(ErrAutoConnectInvalidCount). After
// connecting, when the number of subscribed observers eventually drops to 0,
// AutoConnect will cancel the source connection if it hasn't terminated yet.
// When subsequently the next observer subscribes, AutoConnect will connect to
// the source only when it was previously canceled or because the source
// terminated with an error. So it will not reconnect when the source
// completed succesfully. This specific behavior allows for implementing a
// caching observable that can be retried until it succeeds. Another thing to
// notice is that AutoConnect will disconnect an active connection when the
// number of observers drops to zero. The reason for this is that not doing so
// would leak a task and leave it hanging in the scheduler.
func (o IntMulticaster) AutoConnect(count int) ObservableInt {
	if count < 1 {
		return ThrowInt(ErrAutoConnectInvalidCount)
	}
	var source struct {
		sync.Mutex
		refcount	int32
		subscriber	Subscriber
	}
	observable := func(observe IntObserver, subscribeOn Scheduler, withSubscriber Subscriber) {
		withSubscriber.OnUnsubscribe(func() {
			source.Lock()
			if atomic.AddInt32(&source.refcount, -1) == 0 {
				if source.subscriber != nil {
					source.subscriber.Unsubscribe()
				}
			}
			source.Unlock()
		})
		o.ObservableInt(observe, subscribeOn, withSubscriber)
		source.Lock()
		if atomic.AddInt32(&source.refcount, 1) == int32(count) {
			if source.subscriber == nil || source.subscriber.Error() != nil {
				source.subscriber = subscriber.New()
				source.Unlock()
				o.Connectable(subscribeOn, source.subscriber)
				source.Lock()
			}
		}
		source.Unlock()
	}
	return observable
}

//jig:name ObservableInt_Wait

// Wait subscribes to the Observable and waits for completion or error.
// Returns either the error or nil when the Observable completed normally.
// Wait uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o ObservableInt) Wait() error {
	subscriber := subscriber.New()
	scheduler := scheduler.MakeTrampoline()
	observer := func(next int, err error, done bool) {
		if done {
			subscriber.Done(err)
		}
	}
	subscriber.OnWait(scheduler.Wait)
	o(observer, scheduler, subscriber)
	return subscriber.Wait()
}

//jig:name ErrTypecastToInt

// ErrTypecastToInt is delivered to an observer if the generic value cannot be
// typecast to int.
const ErrTypecastToInt = RxError("typecast to int failed")

//jig:name Observable_AsObservableInt

// AsObservableInt turns an Observable of interface{} into an ObservableInt.
// If during observing a typecast fails, the error ErrTypecastToInt will be
// emitted.
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

//jig:name ObservableInt_SubscribeOn

// SubscribeOn specifies the scheduler an ObservableInt should use when it is
// subscribed to.
func (o ObservableInt) SubscribeOn(scheduler Scheduler) ObservableInt {
	observable := func(observe IntObserver, _ Scheduler, subscriber Subscriber) {
		if scheduler.IsConcurrent() {
			subscriber.OnWait(nil)
		} else {
			subscriber.OnWait(scheduler.Wait)
		}
		o(observe, scheduler, subscriber)
	}
	return observable
}
