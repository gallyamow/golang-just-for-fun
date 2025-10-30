```
goos: linux
goarch: amd64
pkg: github.com/gallyamow/golang-just-for-fun/patterns/latestvalue
cpu: Intel(R) Core(TM) i5-10400 CPU @ 2.90GHz
BenchmarkContainers/atomic-12           39183566                25.80 ns/op            0 B/op          0 allocs/op
BenchmarkContainers/rwmutex-12          15286070                80.29 ns/op            0 B/op          0 allocs/op
BenchmarkContainers/mutex-12             9111933               128.9 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/patterns/latestvalue   3.663s
```