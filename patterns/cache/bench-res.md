### map based

```
goos: darwin
goarch: arm64
pkg: github.com/gallyamow/golang-just-for-fun/patterns/cache
cpu: Apple M4
BenchmarkContainers/single-10               6378            190295 ns/op           27395 B/op       3488 allocs/op
BenchmarkContainers/sharded-8-10            7604            145777 ns/op           59409 B/op       5488 allocs/op
BenchmarkContainers/sharded-64-10          12619             98384 ns/op           59405 B/op       5488 allocs/op
BenchmarkContainers/sharded-128-10         13410             88370 ns/op           59408 B/op       5488 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/patterns/cache 8.065s
```

### slice based

```
goos: darwin
goarch: arm64
pkg: github.com/gallyamow/golang-just-for-fun/patterns/cache
cpu: Apple M4
BenchmarkContainers/single-10               6712            183666 ns/op           27395 B/op       3488 allocs/op
BenchmarkContainers/sharded-8-10            8622            135172 ns/op           59407 B/op       5488 allocs/op
BenchmarkContainers/sharded-64-10          15688             75829 ns/op           59403 B/op       5488 allocs/op
BenchmarkContainers/sharded-128-10         16170             73592 ns/op           59407 B/op       5488 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/patterns/cache 7.500s
```