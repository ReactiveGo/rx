package Merge

import _ "github.com/reactivego/rx"

// Merging two streams of ints generated by FromInt. The result will be
// perfectly interleaved. The Trampoline scheduler used by Println for
// subscribing together with the specific way FromInt is implemented
// will create a repeatable (deterministic) test run with ideal merge
// behavior.
func Example_simple() {
	a := FromInt(0, 2, 4)
	b := FromInt(1, 3, 5)
	MergeInt(a, b).Println()
	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
}
