# rx

    import _ "github.com/reactivego/rx"

[![](https://godoc.org/github.com/reactivego/rx?status.png)](http://godoc.org/github.com/reactivego/rx)

Library `rx` provides [Reactive eXtensions](http://reactivex.io/) for [Go](https://golang.org/). It's a generics library for composing asynchronous and event-based programs using observable sequences. The library consists of more than a 100 templates to enable type-safe programming with observable streams. To use it, you will need the *jig* tool from [Just-in-time Generics for Go](https://github.com/reactivego/jig).

Using the library is very simple. Import the library with the blank identifier `_` as the package name. The side effect of this import is that generics from the library can now be accessed by the *jig* tool. Then start using generics from the library and run *jig* to generate code. The following is a minimal *Hello, World!* program:

```go
package main

import _ "github.com/reactivego/rx"

func main() {
	FromStrings("You!", "Gophers!", "World!").
		MapString(func(x string) string {
			return "Hello, " + x
		}).
		SubscribeNext(func(next string) {
			println(next)
		})

	// Output:
	// Hello, You!
	// Hello, Gophers!
	// Hello, World!
}
```

Take a look at the [Quick Start](doc/quickstart.md) guide to see how it all fits together.

## Why?
ReactiveX observables are somewhat similar to Go channels but have much richer semantics. Observables can be hot or cold, can complete normally or with an error, use subscriptions that can be cancelled from the subscriber side. Where a normal variable is just a place where you read and write values from, an observable captures how the value of this variable changes over time. Concurrency follows naturally from the fact that an observable is an ever changing stream of values.

`rx` is a library of operators that work on one or more observables. The way in which observables can be combined using operators to form new observables is the real strength of ReactiveX. Operators specify how observables representing streams of values are e.g. merged, transformed, concatenated, split, multicasted, replayed, delayed and debounced. My observataion is, that [RxJS 5](https://github.com/ReactiveX/rxjs) and [RxJava 2](https://github.com/ReactiveX/RxJava) have been pushing the envelope in evolving ReactiveX operator semantics. The whole field is still in flux, but the more Rx is applied, the more patterns are emerging. I would like Go to be a participant in this field as well, but for that to happen we need....

## Generic Programming

`rx` is a generics library of templates compatible with [Just-in-time Generics for Go](https://github.com/reactivego/jig). In generic programming you need to specify place-holder types like e.g. the `T` in `Map<T>`. Because we want our generic code to build normally, we work with so called metasyntactic type names like *Foo* and *Bar* so e.g. `MapFoo` instead of `Map<T>`. To actually use the generics just replace *Foo* with the actual type you need e.g. for `int` use `MapInt` and for `string` use `MapString`.

## Operators 

**This implementation of ReactiveX is highly experimental!**
![Gophers are doing experiments](doc/caution.png)

Folowing is a list of [ReactiveX operators](http://reactivex.io/documentation/operators.html) that have been implemented. Operators that are most commonly used got a :star:.

### Creating Operators
Operators that originate new Observables.

- [**CreateFoo :star:**](test/Create/) -> ObservableFoo
- [**DeferFoo**](test/Defer/) -> ObservableFoo
- [**EmptyFoo**](test/Empty/) -> ObservableFoo
- [**FromChanFoo**](test/From/) -> ObservableFoo
- [**FromSliceFoo**](test/From/) -> ObservableFoo
- [**FromFoos**](test/From/) -> ObservableFoo
- [**FromFoo :star:**](test/From/) -> ObservableFoo
- [**Interval**](test/Interval/) -> ObservableInt
- [**JustFoo**](test/Just/) -> ObservableFoo
- [**NeverFoo**](test/Never/) -> ObservableFoo
- Of :star:
- [**Range**](test/Range/) -> ObservableInt
- [**RepeatFoo**](test/Repeat/) -> ObservableFoo
- ObservableFoo -> [**Repeat**](test/Repeat/) -> ObservableFoo
- [**StartFoo**](test/Start/) -> ObservableFoo
- [**ThrowFoo**](test/Throw/) -> ObservableFoo

### Transforming Operators
Operators that transform items that are emitted by an Observable.

- BufferTime :star:
- ConcatMap :star:
- ObservableFoo -> [**MapBar :star:**](test/Map/) -> ObservableBar
- ObservableFoo -> [**MergeMapBar :star:**](test/MergeMap/) -> ObservableBar
- ObservableFoo -> [**ScanBar :star:**](test/Scan/) -> ObservableBar
- SwitchMap :star:

### Filtering Operators
Operators that selectively emit items from a source Observable.

- ObservableFoo -> [**Debounce**](test/Debounce/) -> ObservableFoo
- DebounceTime :star:
- ObservableFoo -> [**Distinct**](test/Distinct/) -> ObservableFoo
- DistinctUntilChanged :star:
- ObservableFoo -> [**ElementAt**](test/ElementAt/) -> ObservableFoo
- ObservableFoo -> [**Filter :star:**](test/Filter/) -> ObservableFoo
- ObservableFoo -> [**First**](test/First/) -> ObservableFoo
- ObservableFoo -> [**IgnoreElements**](test/IgnoreElements/) -> ObservableFoo
- ObservableFoo -> [**IgnoreCompletion**](test/IgnoreCompletion/) -> ObservableFoo
- ObservableFoo -> [**Last**](test/Last/) -> ObservableFoo
- ObservableFoo -> [**Sample**](test/Sample/) -> ObservableFoo
- ObservableFoo -> [**Single**](test/Single/) -> ObservableFoo
- ObservableFoo -> [**Skip**](test/Skip/) -> ObservableFoo
- ObservableFoo -> [**SkipLast**](test/SkipLast/) -> ObservableFoo
- ObservableFoo -> [**Take :star:**](test/Take/) -> ObservableFoo
- TakeUntil :star:
- ObservableFoo -> [**TakeLast**](test/TakeLast/) -> ObservableFoo

### Combining Operators
Operators that work with multiple source Observables to create a single Observable.

- CombineLatest :star:
- [**ConcatFoo :star:**](test/Concat/) -> ObservableFoo
- ObservableFoo -> [**Concat :star:**](test/Concat/) -> ObservableFoo
- Observable<sup>2</sup>Foo -> [**ConcatAll**](test/ConcatAll/) -> ObservableFoo
- [**MergeFoo**](test/Merge/) -> ObservableFoo
- ObservableFoo -> [**Merge :star:**](test/Merge/) -> ObservableFoo
- Observable<sup>2</sup>Foo -> [**MergeAll**](test/MergeAll/) -> ObservableFoo
- [**MergeDelayErrorFoo**](test/MergeDelayError/) -> ObservableFoo
- ObservableFoo -> [**MergeDelayError**](test/MergeDelayError/) -> ObservableFoo
- StartWith :star:
- WithLatestFrom :star:

### Multicasting Operators
Operators that provide subscription multicasting from 1 to multiple subscribers.

- (ObservableFoo) [**Publish**](test/Publish/)() ConnectableFoo
- (ObservableFoo) [**PublishReplay**](test/PublishReplay/)() ConnectableFoo
- PublishLast
- PublishBehavior
- (ConnectableFoo) [**RefCount**](test/RefCount/)() ObservableFoo
- (ConnectableFoo) [**AutoConnect**](test/AutoConnect/)() ObservableFoo
- Share :star:

### Error Handling Operators
Operators that help to recover from error notifications from an Observable.

- ObservableFoo [**Catch :star:**](test/Catch/) -> ObservableFoo
- ObservableFoo [**Retry**](test/Retry/) -> ObservableFoo

### Utility Operators
A toolbox of useful Operators for working with Observables.

- ObservableFoo [**Do :star:**](test/Do/) -> ObservableFoo
- ObservableFoo [**DoOnError**](test/Do/) -> ObservableFoo
- ObservableFoo [**DoOnComplete**](test/Do/) -> ObservableFoo
- ObservableFoo [**Finally**](test/Do/) -> ObservableFoo
- ObservableFoo [**Passthrough**](test/Passthrough/) -> ObservableFoo
- ObservableFoo [**Serialize**](test/Serialize/) -> ObservableFoo
- ObservableFoo [**Timeout**](test/Timeout/) -> ObservableFoo

### Conditional and Boolean Operators
Operators that evaluate one or more Observables or items emitted by Observables.

None yet. Who needs logic anyway?

### Mathematical and Aggregate Operators
Operators that operate on the entire sequence of items emitted by an Observable.

- ObservableFoo  -> [**Average**](test/Average/) -> ObservableFoo
- ObservableFoo  -> [**Count**](test/Count/) -> ObservableInt
- ObservableFoo  -> [**Max**](test/Max/) -> ObservableFoo
- ObservableFoo  -> [**Min**](test/Min/) -> ObservableFoo
- ObservableFoo  -> [**ReduceBar**](test/Reduce/) -> ObservableBar
- ObservableFoo  -> [**Sum**](test/Sum/) -> ObservableFoo

### Scheduling Operators
Change the scheduler for subscribing and observing.

- ObservableFoo -> [**ObserveOn**](test/ObserveOn/) -> ObservableFoo
- ObservableFoo -> [**SubscribeOn**](test/SubscribeOn/) -> ObservableFoo

### Type Casting, Converting and Filtering Operators
Operators to type cast, type convert and type filter observables.

- (ObservableFoo) [**AsObservableBar**](test/AsObservable/)() ObservableBar
- (Observable) [**OnlyFoo**](test/Only/)() ObservableFoo

## Subjects
A *Subject* is both a multicasting *Observable* as well as an *Observer*. The *Observable* side allows multiple simultaneous subscribers. The *Observer* side allows you to directly feed it data or subscribe it to another *Observable*.

- [**NewSubjectFoo**](test/)() SubjectFoo
- [**NewReplaySubjectFoo**](test/)() SubjectFoo

## Subscribing
Subscribing breathes life into a chain of observables.

- ObservableFoo -> [**Subscribe**](test/) Subscriber
	Following methods call Subscribe internally:
	- ConnectableFoo -> [**Connect**](test/) Subscriber
		Following operators call Connect internally:
		- ConnectableFoo -> [**RefCount**](test/) -> ObservableFoo
		- ConnectableFoo -> [**AutoConnect**](test/) -> ObservableFoo
	- ObservableFoo -> [**SubscribeNext**](test/) Subsciber
	- ObservableFoo -> [**ToChan**](test/) -> chan foo
	- ObservableFoo -> [**ToSingle**](test/) -> (foo, error)
	- ObservableFoo -> [**ToSlice**](test/) -> ([]foo, error)
	- ObservableFoo -> [**Wait**](test/) -> error

## Obligatory Dijkstra Quote

Our intellectual powers are rather geared to master static relations and our powers to visualize processes evolving in time are relatively poorly developed. For that reason we should do our utmost to shorten the conceptual gap between the static program and the dynamic process, to make the correspondence between the program (spread out in text space) and the process (spread out in time) as trivial as possible.

*Edsger W. Dijkstra*, March 1968

## Acknowledgements
This library started life as the [Reactive eXtensions for Go](https://github.com/alecthomas/gorx) library by *Alec Thomas*. Although the library has been through the metaphorical meat grinder a few times, its DNA is still clearly present in this library and I owe Alec a debt of grattitude for the work he has made so generously available.

The image *Gophers are doing experiments* was borrowed from the [Go on Mobile](https://github.com/golang/mobile) project and uploaded there by Google's *Hana Kim*.

## License
This library is licensed under the terms of the MIT License. See [LICENSE](LICENSE) file in this repository for copyright notice and exact wording.