# WithLatestFrom

[![](../../../assets/godev.svg?raw=true)](https://pkg.go.dev/github.com/reactivego/rx/test/WithLatestFrom#section-documentation)
[![](../../../assets/rx.svg?raw=true)](https://rxjs-dev.firebaseapp.com/api/operators/withLatestFrom)

**WithLatestFrom** will subscribe to all Observables and wait for all of them to emit before emitting
the first slice. The source observable determines the rate at which the values are emitted. The idea
is that observables that are faster than the source, don't determine the rate at which the resulting
observable emits. The observables that are combined with the source will be allowed to continue
emitting but only will have their last emitted value emitted whenever the source emits.

Note that any values emitted by the source before all other observables have emitted will
effectively be lost. The first emit will occur the first time the source emits after all other
observables have emitted.

![WithLatestFrom](../../../assets/WithLatestFrom.svg?raw=true)

## Example
```go
import _ "github.com/reactivego/rx/generic"
```
Code:
```go
a := FromInt(1, 2, 3, 4, 5).AsObservable()
b := FromString("A", "B", "C", "D", "E").AsObservable()
a.WithLatestFrom(b).Println()
```
Output:
```
[2 A]
[3 B]
[4 C]
[5 D]
```
