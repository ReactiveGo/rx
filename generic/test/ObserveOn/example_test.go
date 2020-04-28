package ObserveOn

import "fmt"

type TaskList struct {
	Tasks []func()
}

func (s *TaskList) Schedule(task func()) {
	s.Tasks = append(s.Tasks, task)
}

func (s *TaskList) Run() {
	for _, task := range s.Tasks {
		task()
	}
}

func Example_observeOn() {
	tasks := &TaskList{}

	// Create a source that uses ObserveOn to park all next calls and the
	// complete call on a custom tasklist.
	source := FromInts(1, 2, 3, 4, 5).ObserveOn(tasks.Schedule)

	// Subscribe to the source and wait for it to complete.
	subscription := source.Subscribe(func(next int, err error, done bool) {
		if !done {
			fmt.Printf("task %d\n", next)
		} else {
			fmt.Println("complete")
		}
	})
	subscription.Wait()

	// Source ran to completion but nothing happended yet, all tasks have been
	// parked.
	fmt.Printf("%d parked tasks\n", len(tasks.Tasks))
	if subscription.Subscribed() {
		fmt.Println("subscribed") // complete has not yet been delivered.
	}

	// Now let's run those tasks
	fmt.Println("---Hey Ho Let's Go!---")
	tasks.Run()
	fmt.Println("--------")

	if subscription.Canceled() {
		fmt.Println("unsubscribed") // complete has been delivered.
	}

	// Output:
	// 6 parked tasks
	// subscribed
	// ---Hey Ho Let's Go!---
	// task 1
	// task 2
	// task 3
	// task 4
	// task 5
	// complete
	// --------
	// unsubscribed
}
