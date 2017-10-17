package test

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Subject is both an observer and an observable.
func Example_subject() {
	subject := NewSubjectInt()

	subscription := subject.SubscribeNext(func(next int) {
		fmt.Println(next)
	}, SubscribeOn(NewGoroutine()))

	subject.Next(123)
	subject.Next(456)
	subject.Complete()

	subscription.Wait()

	// Output:
	// 123
	// 456
}

// Subject will forward errors from the observer to the observable side.
func Example_subjectError() {
	subject := NewSubjectInt()

	// feed the subject...
	go func() {
		subject.Error(errors.New("something bad happened"))
	}()

	err := subject.Wait()
	fmt.Println(err)

	// Output:
	// something bad happened
}

// Subject has an observable side that provides multicasting. This means that
// two subscribers will receive the same data at approximately the same time.
func Example_subjectMultiple() {
	scheduler := NewGoroutine()
	subject := NewSubjectInt()

	var messages struct {
		sync.Mutex
		values []string
	}

	var subscriptions []Subscription
	for i := 0; i < 5; i++ {
		index := i
		subscription := subject.SubscribeNext(func(next int) {

			message := fmt.Sprint(index, next)

			messages.Lock()
			messages.values = append(messages.values, message)
			messages.Unlock()
		}, SubscribeOn(scheduler))
		subscriptions = append(subscriptions, subscription)
	}

	subject.Next(123)
	subject.Next(456)
	subject.Complete()
	subject.Next(111)
	subject.Next(222)

	for i := 0; i < 5; i++ {
		subscriptions[i].Wait()
	}

	sort.Sort(sort.StringSlice(messages.values))
	for _, message := range messages.values {
		fmt.Println(message)
	}

	// Output:
	// 0 123
	// 0 456
	// 1 123
	// 1 456
	// 2 123
	// 2 456
	// 3 123
	// 3 456
	// 4 123
	// 4 456
}