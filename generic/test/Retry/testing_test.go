package Retry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryTrampoline(t *testing.T) {
	errored := false
	a := CreateInt(func(observer IntObserver) {
		observer.Next(1)
		observer.Next(2)
		observer.Next(3)
		if errored {
			observer.Complete()
		} else {
			// Error triggers subscribe and subscribe is scheduled on trampoline....
			observer.Error(RxError("error"))
			errored = true
		}
	}).SubscribeOn(TrampolineScheduler())
	b, e := a.Retry().ToSlice()
	assert.NoError(t, e)
	assert.Equal(t, []int{1, 2, 3, 1, 2, 3}, b)
	assert.True(t, errored)
}

func TestRetryGoroutine(t *testing.T) {
	errored := false
	a := CreateInt(func(observer IntObserver) {
		observer.Next(1)
		observer.Next(2)
		observer.Next(3)
		if errored {
			observer.Complete()
		} else {
			observer.Error(RxError("error"))
			errored = true
		}
	}).SubscribeOn(GoroutineScheduler())
	b, e := a.Retry().ToSlice()
	assert.NoError(t, e)
	assert.Equal(t, []int{1, 2, 3, 1, 2, 3}, b)
	assert.True(t, errored)
}
