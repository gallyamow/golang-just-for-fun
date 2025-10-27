Интересный момент, код практически одинаков, но c slowDo - быстрее.

```
finished after 2.291625msgoos: darwin
goarch: arm64
pkg: github.com/gallyamow/golang-just-for-fun/fun/funsync
cpu: Apple M4
BenchmarkAtomicOnce-10                  895626266                1.416 ns/op           0 B/op          0 allocs/op
BenchmarkSyncOnce-10                    1000000000               0.2330 ns/op          0 B/op          0 allocs/op
BenchmarkChannelWaitGroup-10             3845458               311.9 ns/op            16 B/op          1 allocs/op
BenchmarkStandardWaitGroup-10            8871572               135.0 ns/op            16 B/op          1 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/fun/funsync    5.352s
```