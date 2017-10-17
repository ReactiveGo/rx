package rx

import (
	"time"

	"github.com/reactivego/rx/channel"
)

//jig:template Subject<Foo>
//jig:embeds Observable<Foo>

// SubjectFoo is a combination of an observer and observable. Subjects are
// special because they are the only reactive constructs that support
// multicasting. The items sent to it through its observer side are
// multicasted to multiple clients subscribed to its observable side.
//
// A SubjectFoo embeds ObservableFoo and FooObserveFunc. This exposes the
// methods and fields of both types on SubjectFoo. Use the ObservableFoo
// methods to subscribe to it. Use the FooObserveFunc Next, Error and Complete
// methods to feed data to it.
//
// After a subject has been terminated by calling either Error or Complete,
// it goes into terminated state. All subsequent calls to its observer side
// will be silently ignored. All subsequent subscriptions to the observable
// side will be handled according to the specific behavior of the subject.
// There are different types of subjects, see the different NewXxxSubjectFoo
// functions for more info.
//
// Important! a subject is a hot observable. This means that subscribing to
// it will block the calling goroutine while it is waiting for items and
// notifications to receive. Unless you have code on a different goroutine
// already feeding into the subject, your subscribe will deadlock.
// Alternatively, you could subscribe on a goroutine as shown in the example.
type SubjectFoo struct {
	ObservableFoo
	FooObserveFunc
}

//jig:template NewSubject<Foo>
//jig:needs Subject<Foo>

// NewSubjectFoo creates a new Subject. After the subject is
// terminated, all subsequent subscriptions to the observable side will be
// terminated immediately with either an Error or Complete notification send to
// the subscribing client
//
// Note that this implementation is blocking. When no subcribers are present
// then the data can flow freely. But when there are subscribers, the observable
// goroutine is blocked until all subscribers have processed the next, error or
// complete notification.
func NewSubjectFoo() SubjectFoo {
	ch := channel.NewChan(1, 16 /*max enpoints*/)

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

	observer := func(next foo, err error, done bool) {
		if !ch.Closed() {
			if !done {
				ch.FastSend(next)
			} else {
				ch.Close(err)
			}
		}
	}

	return SubjectFoo{observable.AsObservableFoo(), observer}
}

//jig:template MaxReplayCapacity

// MaxReplayCapacity is the maximum size of a replay buffer. Can be modified.
var MaxReplayCapacity = 16383

//jig:template NewReplaySubject<Foo>
//jig:needs Subject<Foo>, MaxReplayCapacity

// NewReplaySubjectFoo creates a new ReplaySubject. ReplaySubject ensures that
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
func NewReplaySubjectFoo(bufferCapacity int, windowDuration time.Duration) SubjectFoo {
	if bufferCapacity == 0 {
		bufferCapacity = MaxReplayCapacity
	}
	ch := channel.NewChan(bufferCapacity, 16 /*max enpoints*/)

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

	observer := func(next foo, err error, done bool) {
		if !ch.Closed() {
			if !done {
				ch.Send(next)
			} else {
				ch.Close(err)
			}
		}
	}

	return SubjectFoo{observable.AsObservableFoo(), observer}
}