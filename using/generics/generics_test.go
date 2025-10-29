package generics

import (
	"slices"
	"testing"
)
import "golang.org/x/exp/constraints"

// Generics - появились в Go 1.18.
//
// Для чего они:
// Позволяю писать общие функции, которые могут работать с любыми типами. До их появления приходилось либо делать
// для каждого типа отдельную функцию (как в С), либо писать функции, которые работают с any.
//
// Как с ним работать:
//
// Как реализован:
// - Используют monomorphization + dictionaries:
// - Monomorphization - создается несколько вариантов функций и структур
// - Dictionary passing (передача словаря операций) — когда используется constraint (ограничение интерфейсного типа), компилятор передаёт специальную таблицу с функциями и метаданными.
// - Нет ковариантности (как в Java или C#).
// - Не влияют на runtime — всё разворачивается на этапе компиляции.
//
// Что делает компилятор Go при работе с generics:
// 1) Компилятор проверяет, что все операции, которые выполняются с T, разрешены constraint’ом.
// 2) Выводит типы (type inference)
// 3) Специализация (monomorphization) - Go создаёт специализированную версию функции для T. Если функция вызывается с
// другими типами, компилятор создаёт ещё одну версию, это делается только для реально используемых типов, не для всех возможных.
//  func Max_int(a, b int) int { ... }
//  func Max_float64(a, b float64) float64 { ... }
// 4) Когда constraint — не просто набор типов, а интерфейс, содержащий методы, Go не может сделать чистую специализированную,
// версию потому что неизвестно, какие конкретные типы будут использованы.
// func PrintAll[T Stringer](items []T) => func PrintAll(dict *dictionary, items []any)
// Где в dictionary - словарь содержащий указатели на реализацию String() для каждого конкретного типа.
//
// Таким образом:
// Для простых типов (int, float64) generics компилируются почти как обычные функции — без накладных расходов.
// Для interface-like constraints (типы с методами) используется dictionary dispatch → есть небольшой runtime overhead,
// аналогичный вызову метода через интерфейс.
// Компилятор Go делает inlining и escape analysis даже внутри generics, если это возможно.

// Собственные интерфейсы-ограничения (constraints)
type Numbers interface {
	// ~ - позволяет указать не только конкретный тип, но и все типы, чьим базовым (underlying) типом он является.
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 | ~float32 | ~float64
}

type MyInt32 int32

// Обобщенная функция со своим ограничением
func Max[T Numbers](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Обобщенная функция со сторонним ограничением
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Встроенный constraint comparable для типов которые можно сравнить
func IsEqual[T comparable](a, b T) bool {
	return a == b
}

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() T {
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v
}

func (s *Stack[T]) Pad(n int) {
	// Можно создавать переменный указанного типа
	s.items = append(s.items, make([]T, n)...)
}

// Нельзя добавить параметр к методу: Method cannot have type parameters.
// Но как мы видим можно использовать тип самой структуры.
// func (s *Stack[T]) Convert[R any]() int {
// }

func TestUsingGenericsWithFunctions(t *testing.T) {
	t.Run("basic_types", func(t *testing.T) {
		var a float64 = 1
		var b float64 = 2

		// Есть вывод типов, T не нужно указывать явно
		if Max(a, b) != b {
			t.Errorf("got %v, want %v", Max(a, b), b)
		}

		if Min(a, b) != a {
			t.Errorf("got %v, want %v", Min(a, b), a)
		}

		if IsEqual(a, b) {
			t.Errorf("is not equal")
		}
	})

	t.Run("underlying_types", func(t *testing.T) {
		// без ~int32 (даже при наличии int32) MyInt32 does not satisfy Numbers (MyInt32 missing in ~int | ~int64 | ~uint | ~uint32 | ~uint64 | ~float32 | ~float64)
		var a MyInt32 = 1
		var b MyInt32 = 2

		if Max(a, b) != b {
			t.Errorf("got %v, want %v", Max(a, b), b)
		}

		if Min(a, b) != a {
			t.Errorf("got %v, want %v", Min(a, b), a)
		}

		// С константой работает без type conversion, потому что:
		// 1) Значение "a" - автоматически разворачивается до underlying type int32
		// 2) Число "3" является untyped int константой, которая может быть преобразована в int32
		if IsEqual(a, 3) {
			t.Errorf("is not equal")
		}

		// А вот тут будет ошибка: Cannot use 'x' (type int) as the type MyInt32.
		// if IsEqual(a, x) {
		// Ошибка сохранится даже если привести x к int32: Cannot use 'int32(x)' (type int32) as the type MyInt32
		// if IsEqual(a, int32(x)) {
		// В итоге сработает только если привести к MyInt
		x := 3
		if IsEqual(a, MyInt32(x)) {
			t.Errorf("is not equal")
		}
	})
}

func TestUsingGenericsWithStructs(t *testing.T) {
	t.Run("stack", func(t *testing.T) {
		stack := Stack[MyInt32]{}

		stack.Pad(5)
		if stack.Pop() != 0 {
			t.Errorf("got %v, want %v", stack.Pop(), 5)
		}

		// Для совместимых типов
		// автоматически приводи underlying типы
		stack.Push(1)
		stack.Push(2)

		if stack.Pop() != 2 {
			t.Errorf("got %v, want %v", stack.Pop(), 2)
		}
		if stack.Pop() != 1 {
			t.Errorf("got %v, want %v", stack.Pop(), 1)
		}
	})
}

// Collection фильтруемая коллекция.
// @idiomatic: create new instance of unknown class.
type Collection[T any] interface {
	// InitNew фабричный метод для создания пустого экземпляра
	InitEmpty() Collection[T]
	AppendItem(...T)
	GetItems() []T
	Range() func(yield func(T) bool)
}

// SliceCollection реализация на основе slice.
type SliceCollection[T any] struct {
	items []T
}

func (c *SliceCollection[T]) InitEmpty() Collection[T] {
	return &SliceCollection[T]{}
}

func (c *SliceCollection[T]) AppendItem(items ...T) {
	c.items = append(c.items, items...)
}

func (c *SliceCollection[T]) GetItems() []T {
	return c.items
}

func (c *SliceCollection[T]) Range() func(yield func(T) bool) {
	return func(yield func(T) bool) {
		for _, v := range c.items {
			if !yield(v) {
				break
			}
		}
	}
}

// Filter фильтрует коллекцию используя предикат.
func Filter[T any, C Collection[T]](coll C, predicate func(T) bool) Collection[T] {
	res := coll.InitEmpty()

	for _, item := range coll.GetItems() {
		if predicate(item) {
			res.AppendItem(item)
		}
	}

	return res
}

func TestFilterableCollection(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		c1 := &SliceCollection[string]{}
		c1.AppendItem("s1", "s2")

		c2 := c1.InitEmpty()
		if len(c2.GetItems()) != 0 {
			t.Errorf("got non empty")
		}
	})

	t.Run("append+get", func(t *testing.T) {
		c := &SliceCollection[int]{}

		c.AppendItem(1, 2, 3)
		c.AppendItem(4)

		if !slices.Equal(c.GetItems(), []int{1, 2, 3, 4}) {
			t.Errorf("got %v, want %v", c.GetItems(), []int{1, 2, 3, 4})
		}
	})

	t.Run("range", func(t *testing.T) {
		c := &SliceCollection[int]{}

		c.AppendItem(1, 2, 3)
		c.AppendItem(4, 5, 6)

		var res []int
		for v := range c.Range() {
			res = append(res, v)
		}

		if !slices.Equal(c.GetItems(), []int{1, 2, 3, 4, 5, 6}) {
			t.Errorf("got %v, want %v", c.GetItems(), []int{1, 2, 3, 4, 5, 6})
		}
	})
}
