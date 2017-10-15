package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ToChan is used to convert the observable to a channel. The range
// ObservableInt is first converted to an Observable of interface{} so errors
// will be emitted in-band.
func Example_toChan() {
	for value := range Range(1, 9).AsObservable().ToChan() {
		switch value.(type) {
		case int:
			fmt.Print(value)
		case error:
			fmt.Println(value)
		}
	}

	// Output:
	// 123456789
}

// The source ObservableInt created by Range(1,9) will emit values in the range
// 1 to 9. ToChan subscribes asynchronously to the source which begins to emit
// the values into the channel. The channel itself is retured by ToChan and can
// be used by the caller in e.g. a for loop. ToChan changes the scheduler it
// uses for subscribing from the default synchronous trampoline to an
// asynchronous one. The subscription runs concurrently with the channel reading
// for loop.
func Example_toChanInt() {
	for value := range Range(1, 9).ToChan() {
		fmt.Print(value)
	}

	// Output:
	// 123456789
}

// First a source Observable is created that emits 1 to 8 followed by an error.
// ToChan will take the error and output it as the last item emitted by the
// channel.
func Example_toChanError() {
	source := Create(func(observer Observer) {
		for i := 1; i < 9; i++ {
			observer.Next(i)
		}
		observer.Error(errors.New("sad"))
	})

	for value := range source.ToChan() {
		switch value.(type) {
		case int:
			fmt.Print(value)
		case error:
			fmt.Printf("<%v>", value)
		}
	}

	// Output:
	// 12345678<sad>
}

func TestToChanAny(t *testing.T) {
	expect := []interface{}{1, 2, 3, 4, 5, 4, 3, 2, 1}
	source := FromSlice(expect).ToChan()
	result := []interface{}{}
	var err error
	for next := range source {
		switch v := next.(type) {
		case int:
			result = append(result, v)
		case error:
			err = v
		}
	}
	assert.Equal(t, expect, result)
	assert.Nil(t, err)
}

func TestToChan(t *testing.T) {
	expected := []int{1, 2, 3, 4, 5, 4, 3, 2, 1}
	a := FromSliceInt(expected).ToChan()
	b := []int{}
	for i := range a {
		b = append(b, i)
	}
	assert.Equal(t, expected, b)
}
