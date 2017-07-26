package main

////////////////////////////////////////////////////////
// main
////////////////////////////////////////////////////////

func main() {
	println("hello")

	observable := CreateInt(func(s IntObserver) {
		println("subscription received...")
		for i := 0; i < 15; i++ {
			for j := 0; j < 3; j++ {
				if s.Unsubscribed() {
					println("subscriber has left...")
					return
				}
				s.Next(i)
			}
		}
		s.Complete()
	})

	fobservable := observable.Distinct().MapFloat64(func(v int) float64 { return float64(v) * 1.62 })

	term := make(chan struct{})
	subscription := fobservable.Subscribe(func(v float64, err error, complete bool) {
		if err != nil || complete {
			close(term)
			return
		}
		println(v)
	})
	<-term
	if subscription.Unsubscribed() {
		println("subscription unsubscribed....")
	} else {
		println("subscription still alive....")
	}

	println("goodbye")
}