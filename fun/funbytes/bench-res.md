```
oos: darwin
goarch: arm64
pkg: github.com/gallyamow/golang-just-for-fun/fun/funbytes
cpu: Apple M4
BenchmarkBytesBufferWrite-10              563829              2113 ns/op               0 B/op          0 allocs/op
BenchmarkFunBytesBufferWrite-10           327850              3578 ns/op               0 B/op          0 allocs/op
BenchmarkBuffer_Read-10                   250812              4730 ns/op               0 B/op          0 allocs/op
BenchmarkFunBuffer_Read-10                193197              6148 ns/op               1 B/op          0 allocs/op
PASS
ok      github.com/gallyamow/golang-just-for-fun/fun/funbytes   5.973s
```