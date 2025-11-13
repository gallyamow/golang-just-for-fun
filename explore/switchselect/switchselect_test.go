package switchselect

import "testing"

// TestFallthroughOperatorDifference иллюстрирует отличия работы break c С-подобными языками.
// (справедливо и в отношении к select)
//
// В отличие от C или Java, в Go каждый case автоматически завершается после своего тела.
// То есть break ставить не нужно, и нет «проваливания» (fallthrough) к следующему case.
// Если ты хочешь специально перейти к следующему case, можно написать fallthrough.
func TestFallthroughOperatorDifference(t *testing.T) {
	// Нет проваливания в следующий case по умолчанию:
	t.Run("no_default_fallthrough", func(t *testing.T) {
		var vals []int

		for i := range 10 {
			switch i {
			case 1:
			case 2:
				vals = append(vals, i*10)
			case 3:
				vals = append(vals, i*10)
			}
		}

		if len(vals) != 2 {
			t.Errorf("got %d, want %d", len(vals), 2)
		}

		// value 10 - skipped
		if !(vals[0] == 20 && vals[1] == 30) {
			t.Errorf("got %v, want %v", vals, []int{20, 30})
		}
	})

	// Для проваливания в следующий case нужно использовать явно fallthrough
	t.Run("explicit_fallthrough", func(t *testing.T) {
		var vals []int

		for i := range 10 {
			switch i {
			case 1:
				fallthrough
			case 2:
				vals = append(vals, i*10)
			case 3:
				vals = append(vals, i*10)
			}
		}

		if len(vals) != 3 {
			t.Errorf("got %d, want %d", len(vals), 3)
		}

		// value 10 - not skipped
		if !(vals[0] == 10 && vals[1] == 20 && vals[2] == 30) {
			t.Errorf("got %v, want %v", vals, []int{10, 20, 30})
		}
	})
}

// TestBreakOperatorConsistency иллюстрирует что break работает точно также как в C или Java.
// (справедливо и в отношении к select)
// (так как он не обязателен, я думал что он будет относиться к outer for, а по факту к ближайшему for, switch или select)
func TestBreakOperatorConsistency(t *testing.T) {
	// не останавливает outer for
	t.Run("break_refers_to_switch", func(t *testing.T) {
		var vals []int

		for i := range 10 {
			switch i {
			case 1:
				vals = append(vals, i*10)
			case 2:
				vals = append(vals, i*10)
				break
			case 3:
				vals = append(vals, i*10)
			}
		}

		if len(vals) != 3 {
			t.Errorf("got %d, want %d", len(vals), 3)
		}

		if !(vals[0] == 10 && vals[1] == 20 && vals[2] == 30) {
			t.Errorf("got %v, want %v", vals, []int{10, 20, 30})
		}
	})

	t.Run("explicit_breaking_outer", func(t *testing.T) {
		var vals []int

	outer:
		for i := range 10 {
			switch i {
			case 1:
				vals = append(vals, i*10)
			case 2:
				vals = append(vals, i*10)
				break outer
			case 3:
				vals = append(vals, i*10)
			}
		}

		if len(vals) != 2 {
			t.Errorf("got %d, want %d", len(vals), 2)
		}

		if !(vals[0] == 10 && vals[1] == 20) {
			t.Errorf("got %v, want %v", vals, []int{10, 20})
		}
	})
}
