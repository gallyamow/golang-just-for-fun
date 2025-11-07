package zerovalues

import "testing"

// Сравнение поведения ссылочных типов при инициализации в Go
//
// Тип        | var x T                 | x := new(T)                          | x := make(T)              | Примечание
// -----------|-------------------------|--------------------------------------|---------------------------|-----------------------------
// slice      | nil, len=0, cap=0       | *T → пустой слайс len=0, cap=0       | рабочий слайс             | append(*x, val) через new работает
// map        | nil                     | *T → nil                             | рабочий map               | запись в nil или *new → panic
// chan       | nil                     | *T → nil                             | рабочий канал             | запись/чтение в nil → panic
// struct     | zero value для полей    | *T → указатель на struct с нулями    | нет make                  | для struct make не используется
// pointer    | nil                     | *T → выделена память, zero value     | нет make                  | можно разыменовывать, присваивать значение
func TestZeroValues(t *testing.T) {
	// Все значения равны zero value типа. Для массивов будет массив заданной длины из zero value, для структур - структура
	// где все поля это zero-value.
	t.Run("value_types_equals_zero", func(t *testing.T) {
		var b bool
		if b != false {
			t.Errorf("bool is not false")
		}

		var r rune
		if r != 0 {
			t.Errorf("rune is not 0")
		}

		var c complex128
		if real(c) != 0 || imag(c) != 0 {
			t.Errorf("comples is not (0, 0)")
		}

		var s string
		if s != "" {
			t.Errorf("string is not empty")
		}

		var i int
		if i != 0 {
			t.Errorf("int is not 0")
		}

		var f32 float32
		if f32 != 0 {
			t.Errorf("float32 is not 0")
		}

		var f64 float64
		if f64 != 0 {
			t.Errorf("float64 is not 0")
		}

		// zero struct равна структуре где все поля равны их zero values
		var st struct{ x, y int }
		t.Logf("st = %#v, %T", st, st) // st = struct { x int; y int }{x:0, y:0}, struct { x int; y int }

		if st != struct{ x, y int }{x: 0, y: 0} {
			t.Errorf("struct is not zero-filled struct")
		}

		// zero array равна массиву заданной длины где все элементы равны их zero values
		var arr [2]int
		t.Logf("arr = %#v, %T", arr, arr) // arr = [2]int{0, 0}, [2]int

		if arr != [2]int{} {
			t.Errorf("array is not zero filled array")
		}

		var f func(int) int
		if f != nil {
			t.Errorf("func is not nil")
		}

		var e error
		if e != nil {
			t.Errorf("error is not nil")
		}
	})

	// Все ссылочные типы равны nil.
	// Только slice - готов к работе.
	t.Run("reference_types_equals_nil", func(t *testing.T) {
		var sl []int
		t.Logf("sl = %#v, %T", sl, sl) // sl = []int(nil), []int
		if sl != nil {
			t.Errorf("slice is not nil")
		}
		sl = append(sl, 1, 2) // можно писать

		var mp map[int]int
		t.Logf("mp = %#v, %T", mp, mp) // mp = map[int]int(nil), map[int]int
		if mp != nil {
			t.Errorf("map is not nil")
		}

		var ch chan int
		t.Logf("ch = %#v, %T", ch, ch) // ch = (chan int)(nil), chan int
		if ch != nil {
			t.Errorf("ch is not nil")
		}
		go func() {
			for range ch {
			}
		}()
		//ch <- 1 // запись не удается

		var p1 *int
		t.Logf("p1 = %#v, %T", p1, p1) // p1 = *int(nil), *int
		if p1 != nil {
			t.Errorf("p1 is not nil")
		}
	})

	// Создается указатель на элемент.
	// Только slice готов к работе и для pointer(T) где значение будет zero value T.
	t.Run("reference_types_after_new", func(t *testing.T) {
		// new([]int) создаёт указатель на слайс, тип *[]int.
		// Слайс внутри инициализируется нулевой длиной и нулевой ёмкостью (len=0, cap=0), но слайс не nil.
		var sl = new([]int)
		t.Logf("sl = %#v, %T", sl, sl) // sl = &[]int(nil), *[]int
		if sl == nil {
			t.Errorf("slice is nil")
		}
		*sl = append(*sl, 1, 2)

		var mp = new(map[int]int)
		t.Logf("sl = %#v, %T", mp, mp) // mp = map[int]int(nil), map[int]int
		if mp == nil {
			t.Errorf("map is nil")
		}

		var ch = new(chan int)
		t.Logf("ch = %#v, %T", ch, ch) // ch = (chan int)(nil), chan int
		if ch == nil {
			t.Errorf("ch is nil")
		}
		go func() {
			for range *ch {
			}
		}()
		// *ch <- 1 // запись не удается

		// Выделяет память и пишет туда zero-value типа
		var p2 = new(int)
		t.Logf("p2 = %#v, %T, val = %v", p2, p2, *p2) // p2 = (*int)(0xc0000aa190), *int, val = 0
		if p2 == nil {
			t.Errorf("p2 is nil")
		}
		if *p2 != 0 {
			t.Errorf("p2 val is not zero")
		}
	})

	t.Run("var_nil_slice", func(t *testing.T) {
		var s []int

		t.Logf("s = %#v, %T", s, s) // s = []int(nil), []int

		// Значение nil
		if s != nil {
			t.Errorf("nil slice is not nil")
		}

		// Сразу можно добавлять, даже если nil
		s = append(s, 1, 2)
		if len(s) != 2 {
			t.Errorf("len(s) is not 2")
		}
	})

	t.Run("empty_slice", func(t *testing.T) {
		// Warn: Empty slice declaration using a literal
		s := []int{}

		t.Logf("s = %#v, %T", s, s) // s = []int{}, []int

		// Значение не nil
		if s == nil {
			t.Errorf("empty slice is nil")
		}

		s = append(s, 1, 2)
		if len(s) != 2 {
			t.Errorf("len(s) is not 2")
		}
	})

	t.Run("var_nil_map", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("expected panic")
			}
		}()

		var m map[int]int

		t.Logf("m = %#v, %T", m, m) // m = map[int]int(nil), map[int]int

		// Значение nil
		if m != nil {
			t.Errorf("nil map is not nil")
		}

		// Нельзя добавлять значения, будет panic
		m[1] = 1
		m[2] = 2
		if len(m) != 2 {
			t.Errorf("len(s) is not 2")
		}
	})

	t.Run("empty_slice", func(t *testing.T) {
		m := map[int]int{}

		t.Logf("m = %#v", m)
		t.Logf("m = %#v, %T", m, m) // m = map[int]int{}, map[int]int

		// Значение не nil
		if m == nil {
			t.Errorf("empty map is nil")
		}

		// Можно добавлять значения
		m[1] = 1
		m[2] = 2
		if len(m) != 2 {
			t.Errorf("len(s) is not 2")
		}
	})
}
