package Just

import (
	"fmt"

	_ "github.com/reactivego/rx"
)

func Example_basic() {
	err := JustInt(1).Println()

	fmt.Println(err)
	// Output:
	// 1
	// <nil>
}
