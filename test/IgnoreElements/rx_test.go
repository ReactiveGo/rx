package IgnoreElements

import _ "github.com/reactivego/rx"

func Example_ignoreElements() {
	FromInt(1, 2, 3, 4, 5).IgnoreElements().Println()
	// Output:
}
