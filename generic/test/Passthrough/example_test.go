package Passthrough

func Example_passthrough() {
	Range(1, 3).Passthrough().Println()
	// Output:
	// 1
	// 2
	// 3
}
