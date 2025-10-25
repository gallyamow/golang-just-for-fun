### Golang idioms & patterns

* `fun/` - alternative implementations of some `sync` primitives
* `patterns/` - implementations of some golang pattern

### Tests

```sh
# Все тесты
go test -v
go test -bench=. -benchmem > bench-res.md
go test -race
```