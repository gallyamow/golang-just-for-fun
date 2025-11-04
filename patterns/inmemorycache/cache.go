package inmemorycache

// https://www.youtube.com/watch?v=QSfzdf3Dwb0
type Cache[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
}
