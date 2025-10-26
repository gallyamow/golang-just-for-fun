Интересный момент, код практически одинаков, но c slowDo - быстрее.

```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-10400 CPU @ 2.90GHz
BenchmarkAtomicOnce-12          538804942                2.220 ns/op           0 B/op          0 allocs/op
BenchmarkSyncOnce-12            1000000000               0.5121 ns/op          0 B/op          0 allocs/op
PASS
ok      command-line-arguments  1.994s
```