```
goos: linux
goarch: amd64
pkg: github.com/gallyamow/golang-just-for-fun/patterns/semaphore
cpu: Intel(R) Core(TM) i5-10400 CPU @ 2.90GHz
BenchmarkSpinSemaphore-12                 101658             11899 ns/op            1616 B/op         51 allocs/op
BenchmarkMutexSemaphore-12                 84264             13822 ns/op            1617 B/op         51 allocs/op
BenchmarkChannelSemaphore-12               61222             18866 ns/op            1616 B/op         51 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/patterns/semaphore     4.449s
```

```
goos: darwin
goarch: arm64
pkg: github.com/gallyamow/golang-just-for-fun/patterns/semaphore
cpu: Apple M4
BenchmarkSpinSemaphore-10                 177735              6788 ns/op            1616 B/op         51 allocs/op
BenchmarkMutexSemaphore-10                155905              7582 ns/op            1616 B/op         51 allocs/op
BenchmarkChannelSemaphore-10               97669             12074 ns/op            1616 B/op         51 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/patterns/semaphore     4.712s
```