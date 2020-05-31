package Connect

import (
	"fmt"

	_ "github.com/reactivego/rx/generic"
)

func Example_connect() {
	// source is an ObservableInt
	source := FromInt(1, 2)

	// pub is a ConnectableInt
	pub := source.Publish()

	// scheduler will run a task asynchronously on a new goroutine.
	scheduler := GoroutineScheduler()

	// First subscriber (asynchronous)
	sub1 := pub.SubscribeOn(scheduler).
		MapString(func(v int) string {
			return fmt.Sprintf("value is %d", v)
		}).
		Subscribe(func(next string, err error, done bool) {
			if !done {
				fmt.Println("sub1", next)
			}
		})

	// Second subscriber (asynchronous)
	sub2 := pub.SubscribeOn(scheduler).
		MapBool(func(v int) bool {
			return v > 1
		}).
		Subscribe(func(next bool, err error, done bool) {
			if !done {
				fmt.Println("sub2", next)
			}
		})

	// Third subscriber (asynchronous)
	sub3 := pub.SubscribeOn(scheduler).
		Subscribe(func(next int, err error, done bool) {
			if !done {
				fmt.Println("sub3", next)
			}
		})

	// Connect will cause the publisher to subscribe to the source
	pub.Connect().Wait()

	sub1.Wait()
	sub2.Wait()
	sub3.Wait()

	// Unordered output:
	// sub1 value is 1
	// sub2 false
	// sub3 1
	// sub2 true
	// sub3 2
	// sub1 value is 2
}