// Code generated by jig; DO NOT EDIT.

//go:generate jig

package RefCount

import (
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

//jig:name GoroutineScheduler

func GoroutineScheduler() Scheduler {
	return scheduler.Goroutine
}

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

//jig:name Interval

// Interval creates an ObservableInt that emits a sequence of integers spaced
// by a particular time interval. First integer is not emitted immediately, but
// only after the first time interval has passed.
func Interval(interval time.Duration) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
		i := 0
		runner := subscribeOn.ScheduleFutureRecursive(interval, func(self func(time.Duration)) {
			if subscriber.Subscribed() {
				observe(i, nil, false)
				i++
				if subscriber.Subscribed() {
					self(interval)
				}
			}
		})
		subscriber.OnUnsubscribe(runner.Cancel)
	}
	return observable
}

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

//jig:name ObservableInt_Do

// Do calls a function for each next value passing through the observable.
func (o ObservableInt) Do(f func(next int)) ObservableInt {
	observable := func(observe IntObserver, subscribeOn Scheduler, subscriber Subscriber) {
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

//jig:name NewBuffer

const ErrOutOfEndpoints = RxError("out of endpoints")

// NewBuffer creates a buffer to be used as the core of any Subject
// implementation. It returns both an Observer as well as an Observable. Items
// are placed in the buffer through the returned Observer. The buffer then
// multicasts the item to every subscriber of the returned Observable.
//
//	age     age below which items are kept to replay to a new subscriber.
//	length  length of the item buffer, number of items kept to replay to a new subscriber.
//	[cap]   Capacity of the item buffer, number of items that can be observed before blocking.
//	[ecap]  Capacity of the endpoints slice.
func NewBuffer(age time.Duration, length int, capacity ...int) (Observer, Observable) {
	const (
		ms	= time.Millisecond
		us	= time.Microsecond
	)

	// Access to endpoints
	const (
		idle	uint32	= iota
		busy
	)

	// State of endpoint and Chan
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

	type endpoint struct {
		Cursor		uint64
		State		uint64		// active, canceled, closed
		LastActive	time.Time	// track activity to deterime backoff
	}

	type endpoints struct {
		sync.Mutex
		*sync.Cond
		entries	[]endpoint
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

		endpoints	endpoints

		err	error
	}

	make := func(age time.Duration, length int, capacity ...int) *buffer {
		cap, ecap := uint64(length), uint64(32)
		switch {
		case len(capacity) >= 2:
			cap, ecap = uint64(capacity[0]), uint64(capacity[1])
		case len(capacity) == 1:
			cap = uint64(capacity[0])
		}
		len := uint64(length)
		if cap < len {
			cap = len
		}
		cap = uint64(1) << uint(math.Ceil(math.Log2(float64(cap))))
		ch := &buffer{
			len:	len,
			cap:	cap,
			age:	age,
			items:	make([]item, cap),
			mod:	cap - 1,
			end:	cap,
			endpoints: endpoints{
				entries: make([]endpoint, 0, ecap),
			},
		}
		ch.endpoints.Cond = sync.NewCond(&ch.endpoints.Mutex)
		return ch
	}
	ch := make(age, length, capacity...)

	accessEndpoints := func(access func([]endpoint)) bool {
		spun := false
		for !atomic.CompareAndSwapUint32(&ch.endpoints.access, idle, busy) {
			runtime.Gosched()
			spun = true
		}
		access(ch.endpoints.entries)
		atomic.StoreUint32(&ch.endpoints.access, idle)
		return spun
	}

	send := func(value interface{}) {
		for ch.commit == ch.end {
			slowest := parked
			spun := accessEndpoints(func(endpoints []endpoint) {
				for i := range endpoints {
					cursor := atomic.LoadUint64(&endpoints[i].Cursor)
					if cursor < slowest {
						slowest = cursor
					}
				}
				if atomic.LoadUint64(&ch.begin) < slowest && slowest <= atomic.LoadUint64(&ch.end) {
					atomic.StoreUint64(&ch.begin, slowest)
					atomic.StoreUint64(&ch.end, slowest+ch.mod+1)
				} else {
					slowest = parked
				}
			})
			if slowest == parked {

				if !spun {

					runtime.Gosched()
				}
				if atomic.LoadUint64(&ch.state) != active {
					return
				}
			}
		}
		ch.items[ch.commit&ch.mod] = item{Value: value, At: time.Now()}
		atomic.AddUint64(&ch.commit, 1)
		ch.endpoints.Broadcast()
	}

	close := func(err error) {
		if atomic.CompareAndSwapUint64(&ch.state, active, closing) {
			ch.err = err
			if atomic.CompareAndSwapUint64(&ch.state, closing, closed) {
				accessEndpoints(func(endpoints []endpoint) {
					for i := range endpoints {
						atomic.CompareAndSwapUint64(&endpoints[i].State, active, closed)
					}
				})
			}
		}
		ch.endpoints.Broadcast()
	}

	observer := func(next interface{}, err error, done bool) {
		if atomic.LoadUint64(&ch.state) == active {
			if !done {
				send(next)
			} else {
				close(err)
			}
		}
	}

	appendEndpoint := func(cursor uint64) (ep *endpoint, err error) {
		accessEndpoints(func([]endpoint) {
			e := &ch.endpoints
			if len(e.entries) < cap(e.entries) {
				e.entries = append(e.entries, endpoint{Cursor: cursor})
				ep = &e.entries[len(e.entries)-1]
				return
			}
			for i := range e.entries {
				ep = &e.entries[i]
				if atomic.CompareAndSwapUint64(&ep.Cursor, parked, cursor) {
					return
				}
			}
			err = ErrOutOfEndpoints
			return
		})
		return
	}

	observable := func(observe Observer, subscribeOn Scheduler, subscriber Subscriber) {
		cursor := atomic.LoadUint64(&ch.begin)
		ep, err := appendEndpoint(cursor)
		if err != nil {
			runner := subscribeOn.Schedule(func() {
				if subscriber.Subscribed() {
					observe(nil, err, true)
				}
			})
			subscriber.OnUnsubscribe(runner.Cancel)
			return
		}
		commit := atomic.LoadUint64(&ch.commit)
		begin := atomic.LoadUint64(&ch.begin)
		if begin+ch.len < commit {
			atomic.StoreUint64(&ep.Cursor, commit-ch.len)
		}
		atomic.StoreUint64(&ep.State, atomic.LoadUint64(&ch.state))
		ep.LastActive = time.Now()

		receiver := subscribeOn.ScheduleFutureRecursive(0, func(self func(time.Duration)) {
			commit := atomic.LoadUint64(&ch.commit)

			if ep.Cursor == commit {
				if atomic.CompareAndSwapUint64(&ep.State, canceled, canceled) {

					atomic.StoreUint64(&ep.Cursor, parked)
					return
				} else {

					now := time.Now()
					if now.Before(ep.LastActive.Add(1 * ms)) {

						self(50 * us)
						return
					} else if now.Before(ep.LastActive.Add(250 * ms)) {
						if atomic.CompareAndSwapUint64(&ep.State, closed, closed) {

							observe(nil, ch.err, true)
							atomic.StoreUint64(&ep.Cursor, parked)
							return
						}

						self(500 * us)
						return
					} else {
						if subscribeOn.IsConcurrent() {

							ch.endpoints.Lock()
							ch.endpoints.Wait()
							ch.endpoints.Unlock()
							ep.LastActive = time.Now()
							self(0)
							return
						} else {

							self(5 * ms)
							return
						}
					}
				}
			}

			if atomic.LoadUint64(&ep.State) == canceled {
				atomic.StoreUint64(&ep.Cursor, parked)
				return
			}
			for ; ep.Cursor != commit; atomic.AddUint64(&ep.Cursor, 1) {
				item := &ch.items[ep.Cursor&ch.mod]
				if ch.age == 0 || item.At.IsZero() || time.Since(item.At) < ch.age {
					observe(item.Value, nil, false)
				}
				if atomic.LoadUint64(&ep.State) == canceled {
					atomic.StoreUint64(&ep.Cursor, parked)
					return
				}
			}

			ep.LastActive = time.Now()
			self(0)
		})
		subscriber.OnUnsubscribe(receiver.Cancel)

		subscriber.OnUnsubscribe(func() {
			atomic.CompareAndSwapUint64(&ep.State, active, canceled)
			ch.endpoints.Broadcast()
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

//jig:name NewSubjectInt

// NewSubjectInt creates a new Subject. After the subject is terminated, all
// subsequent subscriptions to the observable side will be terminated
// immediately with either an Error or Complete notification send to the
// subscribing client
//
// Note that this implementation is blocking. When there are subscribers, the
// observable goroutine is blocked until all subscribers have processed the
// next, error or complete notification.
func NewSubjectInt() SubjectInt {
	observer, observable := NewBuffer(0, 0, 1, 16)
	return SubjectInt{observer.AsIntObserver(), observable.AsObservableInt()}
}

//jig:name ObservableInt_Publish

// Publish uses the Multicast operator to control the subscription of a
// Subject to a source observable and turns the subject it into a connnectable
// observable. A Subject emits to an observer only those items that are emitted
// by the source Observable subsequent to the time of the observer subscribes.
//
// If the source completed and as a result the internal Subject terminated, then
// calling Connect again will replace the old Subject with a newly created one.
// So this Publish operator is re-connectable, unlike the RxJS 5 behavior that
// isn't. To simulate the RxJS 5 behavior use Publish().AutoConnect(1) this will
// connect on the first subscription but will never re-connect.
func (o ObservableInt) Publish() IntMulticaster {
	return o.Multicast(NewSubjectInt)
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

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

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

//jig:name IntMulticaster_RefCount

// RefCount makes a IntMulticaster behave like an ordinary ObservableInt. On
// first Subscribe it will call Connect on its IntMulticaster and when its last
// subscriber is Unsubscribed it will cancel the source connection by calling
// Unsubscribe on the subscription returned by the call to Connect.
func (o IntMulticaster) RefCount() ObservableInt {
	var source struct {
		sync.Mutex
		refcount	int32
		subscriber	Subscriber
	}
	observable := func(observe IntObserver, subscribeOn Scheduler, withSubscriber Subscriber) {
		withSubscriber.OnUnsubscribe(func() {
			source.Lock()
			if atomic.AddInt32(&source.refcount, -1) == 0 {
				source.subscriber.Unsubscribe()
			}
			source.Unlock()
		})
		o.ObservableInt(observe, subscribeOn, withSubscriber)
		source.Lock()
		if atomic.AddInt32(&source.refcount, 1) == 1 {
			source.subscriber = subscriber.New()
			source.Unlock()
			o.Connectable(subscribeOn, source.subscriber)
			source.Lock()
		}
		source.Unlock()
	}
	return observable
}
