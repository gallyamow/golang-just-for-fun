### Golang idioms & patterns

* `fun/` - alternative implementations of some embedded functions and benchmarking against the original
* `patterns/` - implementations of some golang pattern
* `explore/` - exploring of some facts about golang runtime

### Tests

```sh
# Все тесты
go test -v ./...
go test -bench=. -benchmem ./...
go test -race ./...
```