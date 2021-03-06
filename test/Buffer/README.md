# Buffer

[![](../../../assets/godev.svg?raw=true)](https://pkg.go.dev/github.com/reactivego/rx/test/Buffer#section-documentation)
[![](../../../assets/rx.svg?raw=true)](http://reactivex.io/documentation/operators/buffer.html)

**Buffer** buffers the source Observable values until closingNotifier emits.

![Buffer](../../../assets/Buffer.svg?raw=true)

## Example
```go
import _ "github.com/reactivego/rx/generic"
```
Code:
```go
const ms = time.Millisecond
source := TimerInt(0*ms, 100*ms).Take(4).ConcatMap(func(i int) Observable {
    switch i {
    case 0:
        return From("a", "b")
    case 1:
        return From("c", "d", "e")
    case 3:
        return From("f", "g")
    }
    return Empty()
})
closingNotifier := Interval(100 * ms)
source.Buffer(closingNotifier).Println()
```
Output:
```
[a b]
[c d e]
[]
[f g]
```
