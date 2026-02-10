package allocation

import (
	"slices"
	"testing"
)

func appendElements(sl []int, n int) []int {
	for i := range n {
		sl = append(sl, i)
	}
	return sl
}

func reallocHappened(prev, cur []int) bool {
	if len(prev) == 0 || len(cur) == 0 {
		return false
	}
	return &prev[0] != &cur[0]
}

func TestSlices(t *testing.T) {
	t.Run("using the same array", func(t *testing.T) {
		sl1 := []int{1, 2, 3}
		sl2 := sl1[:1]

		if len(sl2) != 1 {
			t.Fatalf("want 1, got %d", len(sl2))
		}

		if sl2[0] != 1 {
			t.Fatalf("want 1, got %d", sl2[0])
		}

		sl2 = append(sl2, 5)

		if !slices.Equal(sl1, []int{1, 5, 3}) {
			t.Fatalf("want [1, 5, 3], got %d", sl1)
		}

		if !slices.Equal(sl2, []int{1, 5}) {
			t.Fatalf("want [1, 5], got %d", sl1)
		}
	})

	t.Run("append few elements", func(t *testing.T) {
		sl1 := make([]int, 3, 10)
		for i := range 3 {
			sl1[i] = i
		}
		sl2 := sl1[:]

		p1 := &sl1[0]
		p2 := &sl2[0]

		if p1 != p2 {
			t.Fatalf("want equal pointers")
		}

		sl1 = appendElements(sl1, 3)

		p1After := &sl1[0]
		p2After := &sl2[0]

		if p1After != p2After {
			t.Fatalf("want array was not be reallocated")
		}

		// для если не превышаем текущий cap, то массив остается тем же самым
		if cap(sl1) != 10 || cap(sl2) != 10 {
			t.Fatalf("want both cap 10, got %d and %d", cap(sl1), cap(sl2))
		}
	})

	t.Run("append lots elements", func(t *testing.T) {
		sl1 := make([]int, 3, 10)
		for i := range 3 {
			sl1[i] = i
		}
		sl2 := sl1[:]

		p1 := &sl1[0]
		p2 := &sl2[0]

		if p1 != p2 {
			t.Fatalf("want equal pointers")
		}

		sl1 = appendElements(sl1, 1000)

		p1After := &sl1[0]
		p2After := &sl2[0]

		if p1After == p2After {
			t.Fatalf("want array be reallocated")
		}

		// для если превышаем текущий cap, то для второго массив остается тем же самым, а для первого - создается новый
		if cap(sl1) < 1000 || cap(sl2) != 10 {
			t.Fatalf("want both cap > 1000 and 10, got %d and %d", cap(sl1), cap(sl2))
		}
	})

	t.Run("append lots elements without assigning result", func(t *testing.T) {
		sl1 := make([]int, 3, 10)
		for i := range 3 {
			sl1[i] = i
		}
		sl2 := sl1[:]

		p1 := &sl1[0]
		p2 := &sl2[0]

		if p1 != p2 {
			t.Fatalf("want equal pointers")
		}

		appendElements(sl1, 1000)

		p1After := &sl1[0]
		p2After := &sl2[0]

		if p1After != p2After {
			t.Fatalf("want array was not be reallocated")
		}

		// даже если мы превышаем текущий cap, то массив остается тем же самым, потому что:
		// - мы передаем в appendElement копию структуры
		// - происходит reallocation
		// - но именно у копии структуры меняется указатель на backing array
		// - снаружи изменений нет, внешний slice продолжает указывать на тот же backing array
		//
		if cap(sl1) != 10 || cap(sl2) != 10 {
			t.Fatalf("want both cap 10, got %d and %d", cap(sl1), cap(sl2))
		}
	})
}
