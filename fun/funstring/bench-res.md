```
goos: darwin
goarch: arm64
pkg: github.com/gallyamow/golang-just-for-fun/funstring
cpu: Apple M4
BenchmarkFunBuilder-10         	  111872	     78201 ns/op	 2185841 B/op	       1 allocs/op
BenchmarkStandardBuilder-10    	52990755	        30.67 ns/op	     211 B/op	       0 allocs/op
PASS
ok  	github.com/gallyamow/golang-just-for-fun/funstring	10.747s
```