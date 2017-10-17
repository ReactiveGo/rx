package MergeAll

import (
	"fmt"
	"time"
)

func Example_mergeAll() {
	source := CreateObservableString(func(observer ObservableStringObserver) {
		for i := 0; i < 3; i++ {
			time.Sleep(100 * time.Millisecond)
			observer.Next(JustString(fmt.Sprintf("First %d", i)))
			observer.Next(JustString(fmt.Sprintf("Second %d", i)))
		}
		observer.Complete()
	}).MergeAll()

	source.SubscribeNext(func(next string) { fmt.Println(next) })
	// Output:
	// First 0
	// Second 0
	// First 1
	// Second 1
	// First 2
	// Second 2
}