package iterator

type RangeSliceIterator[T any] struct {
	src []T
}

// Range возвращает функцию, которая итерируется по всем значениям в итераторе (Go 1.22
func (it *RangeSliceIterator[T]) Range() func(yield func(T) bool) {
	return func(yield func(T) bool) {
		for _, v := range it.src {
			if !yield(v) {
				// если yield вернул false — выходим из итерации
				return
			}
		}
	}
}
