package iterator

import (
	"slices"
	"testing"
)

func TestNewCustomIterator(t *testing.T) {
	t.Run("base_type", func(t *testing.T) {
		arr1 := []int{1, 2, 3}
		it1 := NewCustomIterator[int](arr1)

		for i := 0; i < len(arr1); i++ {
			v, ok := it1.Next()
			if !ok {
				break
			}

			if v != arr1[i] {
				t.Errorf("got %v, want %v", v, arr1[i])
			}
		}

		arr2 := []string{"string", "hellow", "world"}
		it2 := NewCustomIterator[string](arr2)

		for i := 0; i < len(arr1); i++ {
			v, ok := it2.Next()
			if !ok {
				break
			}

			if v != arr2[i] {
				t.Errorf("got %v, want %v", v, arr2[i])
			}
		}
	})

	t.Run("named_types", func(t *testing.T) {
		type MyInt int

		arr := []MyInt{1, 2, 3}
		it := NewCustomIterator[MyInt](arr)

		for i := 0; i < len(arr); i++ {
			v, ok := it.Next()
			if !ok {
				break
			}

			if v != arr[i] {
				t.Errorf("got %v, want %v", v, arr[i])
			}
		}
	})

	t.Run("type_alias", func(t *testing.T) {
		type MyInt = int

		arr := []MyInt{1, 2, 3}
		it := NewCustomIterator[MyInt](arr)

		for i := 0; i < len(arr); i++ {
			v, ok := it.Next()
			if !ok {
				break
			}

			if v != arr[i] {
				t.Errorf("got %v, want %v", v, arr[i])
			}
		}
	})

	t.Run("user_types", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		arr := []Person{{"Ivan", 10}, {"Mary", 43}}
		it := NewCustomIterator[Person](arr)

		for i := 0; i < len(arr); i++ {
			v, ok := it.Next()
			if !ok {
				break
			}

			if v.Name != arr[i].Name {
				t.Errorf("got %v, want %v", v, arr[i])
			}
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("no panic detected")
			}
			t.Log("expected")
		}()

		type MyInt int

		it := NewCustomIterator[int]([]MyInt{1, 2, 3})

		_, ok := it.Next()
		if !ok {
			t.Fatalf("empty iterator")
		}
	})
}

func TestNewCustomIteratorReflect(t *testing.T) {
	t.Run("mismatch", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic detected")
			}
		}()

		type MyInt int

		it, err := NewCustomIteratorReflect[int]([]MyInt{1, 2, 3})
		if err != nil {
			t.Fatalf("error = %v", err)
		}

		_, ok := it.Next()
		if !ok {
			t.Fatalf("empty iterator")
		}
	})
}

func TestRangeSliceIterator(t *testing.T) {
	t.Run("iterable", func(t *testing.T) {
		it := &RangeSliceIterator[int]{
			src: []int{1, 2, 3},
		}

		var got []int
		for v := range it.Range() {
			got = append(got, v)
		}

		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("got %v, want %v", got, []int{1, 2, 3})
		}
	})
}
