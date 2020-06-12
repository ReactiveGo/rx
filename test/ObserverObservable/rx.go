// Code generated by jig; DO NOT EDIT.

//go:generate jig

package ObserverObservable

import (
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Observer

// Observer is a function that gets called whenever the Observable has
// something to report. The next argument is the item value that is only
// valid when the done argument is false. When done is true and the err
// argument is not nil, then the Observable has terminated with an error.
// When done is true and the err argument is nil, then the Observable has
// completed normally.
type Observer func(next interface{}, err error, done bool)

//jig:name Scheduler

// Scheduler is used to schedule tasks to support subscribing and observing.
type Scheduler = scheduler.Scheduler

//jig:name Subscriber

// Subscriber is an interface that can be passed in when subscribing to an
// Observable. It allows a set of observable subscriptions to be canceled
// from a single subscriber at the root of the subscription tree.
type Subscriber = subscriber.Subscriber

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

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name Observable_SubscribeOn

// SubscribeOn specifies the scheduler an Observable should use when it is
// subscribed to.
func (o Observable) SubscribeOn(scheduler Scheduler) Observable {
	observable := func(observe Observer, _ Scheduler, subscriber Subscriber) {
		if scheduler.IsConcurrent() {
			subscriber.OnWait(nil)
		} else {
			subscriber.OnWait(scheduler.Wait)
		}
		o(observe, scheduler, subscriber)
	}
	return observable
}

//jig:name Subscription

// Subscription is an alias for the subscriber.Subscription interface type.
type Subscription = subscriber.Subscription

//jig:name Observable_Subscribe

// Subscribe operates upon the emissions and notifications from an Observable.
// This method returns a Subscription.
// Subscribe uses a trampoline scheduler created with scheduler.MakeTrampoline().
func (o Observable) Subscribe(observe Observer, subscribers ...Subscriber) Subscription {
	subscribers = append(subscribers, subscriber.New())
	scheduler := scheduler.MakeTrampoline()
	observer := func(next interface{}, err error, done bool) {
		if !done {
			observe(next, err, done)
		} else {
			var zero interface{}
			observe(zero, err, true)
			subscribers[0].Done(err)
		}
	}
	subscribers[0].OnWait(scheduler.Wait)
	o(observer, scheduler, subscribers[0])
	return subscribers[0]
}
