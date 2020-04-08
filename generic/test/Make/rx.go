// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package Make

import (
	"fmt"
	"time"

	"github.com/reactivego/scheduler"
	"github.com/reactivego/subscriber"
)

//jig:name Scheduler

type Scheduler scheduler.Scheduler

//jig:name Subscriber

type Subscriber subscriber.Subscriber

type Subscription subscriber.Subscription

//jig:name IntObserveFunc

type IntObserveFunc func(next int, err error, done bool)

//jig:name zeroInt

var zeroInt int

//jig:name ObservableInt

type ObservableInt func(IntObserveFunc, Scheduler, Subscriber)

//jig:name MakeIntFunc

type MakeIntFunc func(Next func(int), Error func(error), Complete func())

//jig:name MakeInt

func MakeInt(make MakeIntFunc) ObservableInt {
	observable := func(observe IntObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		done := false
		scheduler.ScheduleRecursive(func(self func()) {
			if subscriber.Canceled() {
				return
			}
			next := func(n int) {
				if !subscriber.Canceled() {
					observe(n, nil, false)
				}
			}
			err := func(e error) {
				done = true
				if !subscriber.Canceled() {
					observe(zeroInt, e, true)
				}
			}
			complete := func() {
				done = true
				if !subscriber.Canceled() {
					observe(zeroInt, nil, true)
				}
			}
			make(next, err, complete)
			if !done && !subscriber.Canceled() {
				self()
			}
		})
	}
	return observable
}

//jig:name MakeTimedIntFunc

type MakeTimedIntFunc func(Next func(int), Error func(error), Complete func()) time.Duration

//jig:name MakeTimedInt

func MakeTimedInt(timeout time.Duration, make MakeTimedIntFunc) ObservableInt {
	observable := func(observe IntObserveFunc, scheduler Scheduler, subscriber Subscriber) {
		done := false
		scheduler.ScheduleFutureRecursive(timeout, func(self func(time.Duration)) {
			if subscriber.Canceled() {
				return
			}
			next := func(n int) {
				if !subscriber.Canceled() {
					observe(n, nil, false)
				}
			}
			error := func(e error) {
				done = true
				if !subscriber.Canceled() {
					observe(zeroInt, e, true)
				}
			}
			complete := func() {
				done = true
				if !subscriber.Canceled() {
					observe(zeroInt, nil, true)
				}
			}
			timeout = make(next, error, complete)
			if !done && !subscriber.Canceled() {
				self(timeout)
			}
		})
	}
	return observable
}

//jig:name Schedulers

func TrampolineScheduler() Scheduler	{ return scheduler.Trampoline }

func GoroutineScheduler() Scheduler	{ return scheduler.Goroutine }

//jig:name ObservableIntPrintln

func (o ObservableInt) Println() (err error) {
	subscriber := subscriber.New()
	scheduler := TrampolineScheduler()
	observer := func(next int, e error, done bool) {
		if !done {
			fmt.Println(next)
		} else {
			err = e
			subscriber.Unsubscribe()
		}
	}
	o(observer, scheduler, subscriber)
	subscriber.Wait()
	return
}

//jig:name ObservableTake

func (o Observable) Take(n int) Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		taken := 0
		observer := func(next interface{}, err error, done bool) {
			if taken < n {
				observe(next, err, done)
				if !done {
					taken++
					if taken >= n {
						observe(nil, nil, true)
					}
				}
			}
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name ObservableIntTake

func (o ObservableInt) Take(n int) ObservableInt {
	return o.AsObservable().Take(n).AsObservableInt()
}

//jig:name ObserveFunc

type ObserveFunc func(next interface{}, err error, done bool)

//jig:name zero

var zero interface{}

//jig:name Observable

type Observable func(ObserveFunc, Scheduler, Subscriber)

//jig:name ObservableIntAsObservable

func (o ObservableInt) AsObservable() Observable {
	observable := func(observe ObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next int, err error, done bool) {
			observe(interface{}(next), err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}

//jig:name RxError

type RxError string

func (e RxError) Error() string	{ return string(e) }

//jig:name ErrTypecastToInt

const ErrTypecastToInt = RxError("typecast to int failed")

//jig:name ObservableAsObservableInt

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
