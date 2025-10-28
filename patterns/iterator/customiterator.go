package iterator

import (
	"fmt"
	"reflect"
)

type CustomIterator[T any] interface {
	Next() (T, bool)
}

// NewCustomIterator создает CustomIterator для любого типа
// Работает, только когда передается []T, где T совпадает с типом generic. Таким образом проще использовать NewSliceIterator.
// Например, NewCustomIterator[int]([]MyInt{…}) - не сработает.
// Generics в Go работают на этапе компиляции, а any — это динамический тип, проверяемый во время выполнения.
// Для настоящего any надо работать через рефлексию.
func NewCustomIterator[T any](src any) CustomIterator[T] {
	switch src.(type) {
	case []T:
		return NewSliceIterator(src.([]T))
	default:
		return nil
	}
}

// NewCustomIteratorReflect создает CustomIterator для любого типа
// @idiomatic: new(T) - создает нулевое значение типа T в куче и возвращает указатель на него, * - разыменовывание и получение нулевого значение.
// это эквивалентно записи `var zero T`
func NewCustomIteratorReflect[T any](src any) (CustomIterator[T], error) {
	// reflect.ValueOf(x) возвращает обёртку, которая знает фактический тип объекта на этапе выполнения.
	refSrc := reflect.ValueOf(src)
	zero := *new(T)
	refType := reflect.TypeOf(zero)

	switch refSrc.Kind() {
	case reflect.Slice:
		slice := make([]T, refSrc.Len())

		// Нельзя сделать просто slice := val.Interface().([]T), потому что компилятор не знает, что interface{} действительно содержит []T.
		// 1) Тип T известен только компилятору, на этапе выполнения (runtime) T — неизвестен
		for i := 0; i < refSrc.Len(); i++ {
			// Этот метод возвращает interface{}
			el := refSrc.Index(i)

			// Так нельзя:
			// typedEl, ok := el.(T)
			// Go не делает автоматическое "повышение" или "понижение" типов даже с одинаковым underlying type.
			// Кстати underlined types не совместимы между собой, например нельзя сравнить a = b и нужно явное именно "type conversion" a = int(b).
			// Пробовал "type conversion", компиляция проходит, но ошибка в runtime = cannot convert el.Interface()
			// (value of interface type any) to type T: need type assertion. А оно не работает с интерфейсами.
			// typedEl := T(el.Interface())
			// slice[i] = typedEl
			//
			// Заметим что:
			// "Приведение тип" = "type conversion" = T(x) - используется когда типы совместимы (underlined types).
			// "Утверждение типа" = "type assertion" = x.(T) - используется когда операнд interface.

			if !el.Type().ConvertibleTo(refType) {
				// Go не делает автоматическое "повышение" или "понижение" типов даже с одинаковым underlying type.
				return nil, fmt.Errorf("element at index %d cannot be converted to type %T", i, zero)
			}

			// Поэтому используем convert
			slice[i] = el.Convert(refType).Interface().(T)
		}

		return NewSliceIterator[T](slice), nil
	default:
		return nil, nil
	}
}

func NewSliceIterator[T any](src []T) CustomIterator[T] {
	return &sliceIterator[T]{src: src}
}

// @idiomatic: underlying types parameter
// @idiomatic: array parameter
type sliceIterator[T any] struct {
	src []T
	i   int
}

// Next returns the next value in the iterator.
// @idiomatic: zero value creation
func (it *sliceIterator[T]) Next() (T, bool) {
	if it.i >= len(it.src) {
		var zero T // или так zero := *new(T)
		return zero, false
	}

	v := it.src[it.i]
	it.i++
	return v, true
}
