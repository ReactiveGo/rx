package Passthrough

import _ "github.com/reactivego/rx"

func Example_passthrough() {
	Range(1, 3).Passthrough().Println()
	// Output:
	// 1
	// 2
	// 3
}
